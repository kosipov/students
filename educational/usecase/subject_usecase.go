package usecase

import (
	"context"
	"crypto/subtle"
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/kosipov/students/educational"
	"github.com/kosipov/students/models"
	"golang.org/x/sync/singleflight"
	"log"
	"strconv"
	"strings"
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

func (subjectUseCase *SubjectUseCase) GetSubjectWithSubjectObjects(ctx context.Context, id int) (*models.Subject, error) {
	subject, err := subjectUseCase.subjectRepo.GetSubjectWithSubjectObjects(ctx, id)
	if err != nil {
		return nil, notFound(err, educational.ErrSubjectNotFound)
	}
	return subject, nil
}

func (subjectUseCase *SubjectUseCase) CreateSubject(ctx context.Context, groupId int, input educational.SubjectInput) (*models.Subject, error) {
	if err := input.Normalize(); err != nil {
		return nil, err
	}
	if _, err := subjectUseCase.subjectRepo.GetGroup(ctx, groupId); err != nil {
		return nil, notFound(err, educational.ErrGroupNotFound)
	}

	subject := &models.Subject{SubjectName: input.Name, GroupId: groupId}
	if err := subjectUseCase.subjectRepo.CreateSubject(ctx, subject); err != nil {
		return nil, err
	}
	return subject, nil
}

func (subjectUseCase *SubjectUseCase) UpdateSubject(ctx context.Context, id int, input educational.SubjectInput) (*models.Subject, error) {
	if err := input.Normalize(); err != nil {
		return nil, err
	}
	subject, err := subjectUseCase.getSubject(ctx, id)
	if err != nil {
		return nil, err
	}

	subject.SubjectName = input.Name
	if err := subjectUseCase.subjectRepo.UpdateSubject(ctx, subject); err != nil {
		return nil, err
	}
	return subject, nil
}

func (subjectUseCase *SubjectUseCase) DeleteSubject(ctx context.Context, id int) error {
	subject, err := subjectUseCase.getSubject(ctx, id)
	if err != nil {
		return err
	}
	return subjectUseCase.subjectRepo.DeleteSubject(ctx, subject)
}

func (subjectUseCase *SubjectUseCase) CreateSubjectObject(ctx context.Context, subjectId int, input educational.SubjectObjectInput) (*models.SubjectObject, error) {
	if err := input.Normalize(); err != nil {
		return nil, err
	}
	if _, err := subjectUseCase.getSubject(ctx, subjectId); err != nil {
		return nil, err
	}

	categories, err := subjectUseCase.toCategories(ctx, input.Categories)
	if err != nil {
		return nil, err
	}

	subjectObject := &models.SubjectObject{
		SubjectId:  subjectId,
		Name:       input.Name,
		Href:       input.Href,
		Comment:    input.Comment,
		Hidden:     input.Hidden,
		Categories: categories,
		Password:   input.Password,
	}
	if err := subjectUseCase.subjectRepo.CreateSubjectObject(ctx, subjectObject); err != nil {
		return nil, err
	}

	subjectUseCase.prefetchContent(subjectObject)

	return subjectObject, nil
}

func (subjectUseCase *SubjectUseCase) UpdateSubjectObject(ctx context.Context, id int, patch educational.SubjectObjectPatch) (*models.SubjectObject, error) {
	if err := patch.Normalize(); err != nil {
		return nil, err
	}
	subjectObject, err := subjectUseCase.getSubjectObject(ctx, id)
	if err != nil {
		return nil, err
	}

	hrefChanged := patch.Href != nil && *patch.Href != subjectObject.Href
	if patch.Name != nil {
		subjectObject.Name = *patch.Name
	}
	if patch.Href != nil {
		subjectObject.Href = *patch.Href
	}
	if patch.Comment != nil {
		subjectObject.Comment = *patch.Comment
	}
	if patch.Hidden != nil {
		subjectObject.Hidden = *patch.Hidden
	}
	if patch.Password != nil && *patch.Password != subjectObject.Password {
		subjectObject.Password = *patch.Password
		subjectObject.PasswordVersion++
	}
	if patch.Categories != nil {
		categories, err := subjectUseCase.toCategories(ctx, *patch.Categories)
		if err != nil {
			return nil, err
		}
		subjectObject.Categories = categories
	}

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

func (subjectUseCase *SubjectUseCase) DeleteSubjectObject(ctx context.Context, id int) error {
	subjectObject, err := subjectUseCase.getSubjectObject(ctx, id)
	if err != nil {
		return err
	}
	return subjectUseCase.subjectRepo.DeleteSubjectObject(ctx, subjectObject)
}

func (subjectUseCase *SubjectUseCase) RefreshSubjectObjectContent(ctx context.Context, id int) (*models.SubjectObject, error) {
	subjectObject, err := subjectUseCase.getSubjectObject(ctx, id)
	if err != nil {
		return nil, err
	}
	if !subjectUseCase.contentFetcher.Supports(subjectObject.Href) {
		return subjectObject, educational.ErrSubjectObjectNotDocument
	}

	return subjectUseCase.refreshContent(*subjectObject, true), nil
}

func (subjectUseCase *SubjectUseCase) GetCategoryNames(ctx context.Context) ([]string, error) {
	return subjectUseCase.subjectRepo.GetCategoryNames(ctx)
}

// toCategories turns normalized names into categories. A name that already exists in another
// letter case takes the existing spelling, so students see one filter instead of two.
func (subjectUseCase *SubjectUseCase) toCategories(ctx context.Context, names []string) ([]models.SubjectObjectCategory, error) {
	if len(names) == 0 {
		return []models.SubjectObjectCategory{}, nil
	}

	existing, err := subjectUseCase.subjectRepo.GetCategoryNames(ctx)
	if err != nil {
		return nil, err
	}
	spelling := make(map[string]string, len(existing))
	for _, name := range existing {
		if _, ok := spelling[educational.CategoryKey(name)]; !ok {
			spelling[educational.CategoryKey(name)] = name
		}
	}

	categories := make([]models.SubjectObjectCategory, 0, len(names))
	for _, name := range names {
		if known, ok := spelling[educational.CategoryKey(name)]; ok {
			name = known
		}
		categories = append(categories, models.SubjectObjectCategory{Name: name})
	}
	return categories, nil
}

func (subjectUseCase *SubjectUseCase) IsDocument(subjectObject *models.SubjectObject) bool {
	return !subjectObject.ContentUnsupported && subjectUseCase.contentFetcher.Supports(subjectObject.Href)
}

func (subjectUseCase *SubjectUseCase) GetTask(ctx context.Context, subjectObjectId int, access educational.TaskAccess) (*educational.Task, error) {
	task, err := subjectUseCase.getVisibleTask(ctx, subjectObjectId)
	if err != nil {
		return nil, err
	}

	subjectObject := task.SubjectObject
	if subjectObject.IsProtected() && !access.IsUnlocked(subjectObject.ID, subjectObject.PasswordVersion) {
		return task, educational.ErrTaskLocked
	}

	return subjectUseCase.loadDocument(task)
}

func (subjectUseCase *SubjectUseCase) UnlockTask(ctx context.Context, subjectObjectId int, password string) (*models.SubjectObject, error) {
	task, err := subjectUseCase.getVisibleTask(ctx, subjectObjectId)
	if err != nil {
		return nil, err
	}

	subjectObject := task.SubjectObject
	if !subjectObject.IsProtected() {
		return subjectObject, nil
	}
	// Constant time, so response timing doesn't tell how much of a guess was right.
	if subtle.ConstantTimeCompare([]byte(strings.TrimSpace(password)), []byte(subjectObject.Password)) != 1 {
		return nil, educational.ErrWrongPassword
	}
	return subjectObject, nil
}

// getVisibleTask loads a task students can see: neither the task nor its group is hidden.
func (subjectUseCase *SubjectUseCase) getVisibleTask(ctx context.Context, subjectObjectId int) (*educational.Task, error) {
	task, err := subjectUseCase.getTask(ctx, subjectObjectId)
	if err != nil {
		return nil, err
	}
	if task.SubjectObject.Hidden || task.Group.Hidden {
		return nil, educational.ErrSubjectObjectNotFound
	}
	return task, nil
}

func (subjectUseCase *SubjectUseCase) PreviewTask(ctx context.Context, subjectObjectId int) (*educational.Task, error) {
	task, err := subjectUseCase.getTask(ctx, subjectObjectId)
	if err != nil {
		return nil, err
	}

	return subjectUseCase.loadDocument(task)
}

// getTask loads the subject object with its subject and group.
// A subject object whose subject or group no longer exists is not found.
func (subjectUseCase *SubjectUseCase) getTask(ctx context.Context, subjectObjectId int) (*educational.Task, error) {
	subjectObject, err := subjectUseCase.getSubjectObject(ctx, subjectObjectId)
	if err != nil {
		return nil, err
	}

	subject, err := subjectUseCase.subjectRepo.GetSubject(ctx, subjectObject.SubjectId)
	if err != nil {
		return nil, notFound(err, educational.ErrSubjectObjectNotFound)
	}

	group, err := subjectUseCase.subjectRepo.GetGroup(ctx, subject.GroupId)
	if err != nil {
		return nil, notFound(err, educational.ErrSubjectObjectNotFound)
	}

	return &educational.Task{SubjectObject: subjectObject, Subject: subject, Group: group}, nil
}

// loadDocument fills the task with its stored document, refreshing it when it is outdated.
func (subjectUseCase *SubjectUseCase) loadDocument(task *educational.Task) (*educational.Task, error) {
	subjectObject := task.SubjectObject
	if !subjectUseCase.IsDocument(subjectObject) {
		return task, educational.ErrSubjectObjectNotDocument
	}

	if !subjectUseCase.isContentOutdated(subjectObject) {
		return task, nil
	}

	if subjectObject.ContentFetchedAt == nil {
		// Nothing to show yet, so the student waits for the first download.
		task.SubjectObject = subjectUseCase.refreshContent(*subjectObject, false)
		if task.SubjectObject.ContentUnsupported {
			return task, educational.ErrSubjectObjectNotDocument
		}
		return task, nil
	}

	// The stored document is shown right away and updated for the next views.
	go subjectUseCase.refreshContent(*subjectObject, false)

	return task, nil
}

func (subjectUseCase *SubjectUseCase) getSubject(ctx context.Context, id int) (*models.Subject, error) {
	subject, err := subjectUseCase.subjectRepo.GetSubject(ctx, id)
	if err != nil {
		return nil, notFound(err, educational.ErrSubjectNotFound)
	}
	return subject, nil
}

func (subjectUseCase *SubjectUseCase) getSubjectObject(ctx context.Context, id int) (*models.SubjectObject, error) {
	subjectObject, err := subjectUseCase.subjectRepo.GetSubjectObject(ctx, id)
	if err != nil {
		return nil, notFound(err, educational.ErrSubjectObjectNotFound)
	}
	return subjectObject, nil
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
// that triggered it finishes, and concurrent calls for the same link share one download.
// With force the document is downloaded even if it has not changed.
func (subjectUseCase *SubjectUseCase) refreshContent(subjectObject models.SubjectObject, force bool) *models.SubjectObject {
	// The link is part of the key: a download of the old link must not be reused after the link changes.
	key := strconv.Itoa(subjectObject.ID) + " " + subjectObject.Href
	result, _, _ := subjectUseCase.refreshGroup.Do(key, func() (interface{}, error) {
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

// notFound replaces gorm's "record not found" with the domain error.
func notFound(err error, domainErr error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domainErr
	}
	return err
}
