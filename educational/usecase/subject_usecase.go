package usecase

import (
	"context"
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/kosipov/students/educational"
	"github.com/kosipov/students/models"
	"golang.org/x/sync/singleflight"
	"log"
	"strconv"
	"time"
)

const (
	// contentRefreshInterval is how long a stored document is shown without checking the source.
	contentRefreshInterval = 5 * time.Minute
	// contentFetchTimeout limits a single refresh, including the time a student waits
	// for a document that has never been downloaded.
	contentFetchTimeout = 15 * time.Second
)

type SubjectUseCase struct {
	subjectRepo    educational.CommonSubjectRepository
	contentFetcher educational.ContentFetcher
	// refreshGroup makes concurrent views of the same outdated document share one download.
	refreshGroup singleflight.Group
	now          func() time.Time
}

func NewSubjectUseCase(subjectRepo educational.CommonSubjectRepository, contentFetcher educational.ContentFetcher) *SubjectUseCase {
	return &SubjectUseCase{
		subjectRepo:    subjectRepo,
		contentFetcher: contentFetcher,
		now:            time.Now,
	}
}

func (subjectUseCase *SubjectUseCase) GetSubjectsByGroup(ctx context.Context, groupId int) (*[]models.Subject, error) {
	group, err := subjectUseCase.subjectRepo.GetGroup(ctx, groupId)
	if err != nil {
		return nil, err
	}
	return subjectUseCase.subjectRepo.GetSubjectsByGroup(ctx, group)
}

func (subjectUseCase *SubjectUseCase) SubjectObjectListFromSubject(ctx context.Context, subjectId int) (*[]models.SubjectObject, error) {
	subject, err := subjectUseCase.GetSubjectById(ctx, subjectId)
	if err != nil {
		return nil, err
	}
	return subjectUseCase.subjectRepo.GetSubjectObjectsBySubject(ctx, subject)
}

func (subjectUseCase *SubjectUseCase) GetSubjectById(ctx context.Context, id int) (*models.Subject, error) {
	subject, err := subjectUseCase.subjectRepo.GetSubject(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, educational.ErrSubjectNotFound
		}
		return nil, err
	}
	return subject, nil
}

func (subjectUseCase *SubjectUseCase) GetAllSubject(ctx context.Context) (*[]models.Subject, error) {
	return subjectUseCase.subjectRepo.GetSubjects(ctx)
}

func (subjectUseCase *SubjectUseCase) CreateSubject(ctx context.Context, name string, groupId int) error {
	subject := &models.Subject{SubjectName: name, GroupId: groupId}

	return subjectUseCase.subjectRepo.CreateSubject(ctx, subject)
}

func (subjectUseCase *SubjectUseCase) CreateSubjectObject(ctx context.Context, name string, subjectId int, href string) (*models.SubjectObject, error) {
	subjectObject := &models.SubjectObject{SubjectId: subjectId, Name: name, Href: href}

	if err := subjectUseCase.subjectRepo.CreateSubjectObject(ctx, subjectObject); err != nil {
		return nil, err
	}

	subjectUseCase.prefetchContent(subjectObject)

	return subjectObject, nil
}

func (subjectUseCase *SubjectUseCase) GetSubjectObject(ctx context.Context, subjectId int, subjectObjectId int) (*models.SubjectObject, error) {
	subjectObject, err := subjectUseCase.subjectRepo.GetSubjectObject(ctx, subjectObjectId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, educational.ErrSubjectObjectNotFound
		}
		return nil, err
	}
	if subjectObject.SubjectId != subjectId {
		return nil, educational.ErrSubjectObjectNotFound
	}
	return subjectObject, nil
}

func (subjectUseCase *SubjectUseCase) UpdateSubjectObject(ctx context.Context, subjectId int, subjectObjectId int, name string, href string) (*models.SubjectObject, error) {
	subjectObject, err := subjectUseCase.GetSubjectObject(ctx, subjectId, subjectObjectId)
	if err != nil {
		return nil, err
	}

	hrefChanged := subjectObject.Href != href
	subjectObject.Name = name
	subjectObject.Href = href

	if err := subjectUseCase.subjectRepo.UpdateSubjectObject(ctx, subjectObject); err != nil {
		return nil, err
	}

	if hrefChanged {
		// The stored document belongs to the old link.
		resetContent(subjectObject)
		if err := subjectUseCase.subjectRepo.UpdateSubjectObjectContent(ctx, subjectObject); err != nil {
			return nil, err
		}
		subjectUseCase.prefetchContent(subjectObject)
	}

	return subjectObject, nil
}

func (subjectUseCase *SubjectUseCase) GetTask(ctx context.Context, subjectObjectId int) (*models.SubjectObject, error) {
	subjectObject, err := subjectUseCase.subjectRepo.GetSubjectObject(ctx, subjectObjectId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, educational.ErrSubjectObjectNotFound
		}
		return nil, err
	}

	if !subjectUseCase.IsDocument(subjectObject) {
		return subjectObject, educational.ErrSubjectObjectNotDocument
	}

	if !subjectUseCase.isContentOutdated(subjectObject) {
		return subjectObject, nil
	}

	if subjectObject.ContentFetchedAt == nil {
		// Nothing to show yet, so the student waits for the first download.
		subjectObject = subjectUseCase.refreshContent(*subjectObject, false)
		if subjectObject.ContentUnsupported {
			return subjectObject, educational.ErrSubjectObjectNotDocument
		}
		return subjectObject, nil
	}

	// The stored document is shown right away and updated for the next views.
	go subjectUseCase.refreshContent(*subjectObject, false)

	return subjectObject, nil
}

func (subjectUseCase *SubjectUseCase) RefreshSubjectObjectContent(ctx context.Context, subjectId int, subjectObjectId int) (*models.SubjectObject, error) {
	subjectObject, err := subjectUseCase.GetSubjectObject(ctx, subjectId, subjectObjectId)
	if err != nil {
		return nil, err
	}
	if !subjectUseCase.contentFetcher.Supports(subjectObject.Href) {
		return subjectObject, educational.ErrSubjectObjectNotDocument
	}

	return subjectUseCase.refreshContent(*subjectObject, true), nil
}

func (subjectUseCase *SubjectUseCase) IsDocument(subjectObject *models.SubjectObject) bool {
	return !subjectObject.ContentUnsupported && subjectUseCase.contentFetcher.Supports(subjectObject.Href)
}

// prefetchContent downloads the document of a new link in the background,
// so the first student who opens it doesn't wait for OneDrive.
func (subjectUseCase *SubjectUseCase) prefetchContent(subjectObject *models.SubjectObject) {
	if subjectUseCase.contentFetcher.Supports(subjectObject.Href) {
		go subjectUseCase.refreshContent(*subjectObject, false)
	}
}

func (subjectUseCase *SubjectUseCase) isContentOutdated(subjectObject *models.SubjectObject) bool {
	return subjectObject.ContentCheckedAt == nil ||
		subjectUseCase.now().Sub(*subjectObject.ContentCheckedAt) >= contentRefreshInterval
}

// refreshContent downloads the document and stores the result, including a failure.
// It works on a copy with its own timeout, so it is not interrupted when the request
// that triggered it finishes, and concurrent calls for the same subject object share one download.
// With force the document is downloaded even if it has not changed.
func (subjectUseCase *SubjectUseCase) refreshContent(subjectObject models.SubjectObject, force bool) *models.SubjectObject {
	result, _, _ := subjectUseCase.refreshGroup.Do(strconv.Itoa(subjectObject.ID), func() (interface{}, error) {
		ctx, cancel := context.WithTimeout(context.Background(), contentFetchTimeout)
		defer cancel()

		etag := subjectObject.ContentETag
		if force {
			etag = ""
		}

		content, err := subjectUseCase.contentFetcher.Fetch(ctx, subjectObject.Href, etag)
		now := subjectUseCase.now().UTC()
		subjectObject.ContentCheckedAt = &now

		switch {
		case err == nil:
			subjectObject.Content = string(content.Body)
			subjectObject.ContentETag = content.ETag
			subjectObject.ContentFetchedAt = &now
			subjectObject.ContentError = ""
			subjectObject.ContentUnsupported = false
		case errors.Is(err, educational.ErrContentNotModified):
			subjectObject.ContentError = ""
		case errors.Is(err, educational.ErrContentUnsupported):
			resetContent(&subjectObject)
			subjectObject.ContentCheckedAt = &now
			subjectObject.ContentUnsupported = true
		default:
			log.Printf("Failed to fetch content of subject object %d: %s", subjectObject.ID, err)
			subjectObject.ContentError = err.Error()
		}

		if err := subjectUseCase.subjectRepo.UpdateSubjectObjectContent(ctx, &subjectObject); err != nil {
			log.Printf("Failed to store content of subject object %d: %s", subjectObject.ID, err)
		}

		return &subjectObject, nil
	})

	return result.(*models.SubjectObject)
}

func resetContent(subjectObject *models.SubjectObject) {
	subjectObject.Content = ""
	subjectObject.ContentETag = ""
	subjectObject.ContentFetchedAt = nil
	subjectObject.ContentCheckedAt = nil
	subjectObject.ContentError = ""
	subjectObject.ContentUnsupported = false
}

func (subjectUseCase *SubjectUseCase) DeleteSubjectObject(ctx context.Context, subjectId int, subjectObjectId int) error {
	subjectObject, err := subjectUseCase.GetSubjectObject(ctx, subjectId, subjectObjectId)
	if err != nil {
		return err
	}
	return subjectUseCase.subjectRepo.DeleteSubjectObject(ctx, subjectObject)
}
