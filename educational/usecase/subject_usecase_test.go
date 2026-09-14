package usecase

import (
	"context"
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/kosipov/students/educational"
	"github.com/kosipov/students/models"
	"strings"
	"sync"
	"testing"
	"time"
)

const oneDriveHref = "https://1drv.ms/t/c/1/abc"

type fakeSubjectRepo struct {
	educational.CommonSubjectRepository

	mu             sync.Mutex
	subjectObjects map[int]models.SubjectObject
}

func (r *fakeSubjectRepo) GetSubjectObject(ctx context.Context, id int) (*models.SubjectObject, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	subjectObject, ok := r.subjectObjects[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &subjectObject, nil
}

func (r *fakeSubjectRepo) CreateSubjectObject(ctx context.Context, subjectObject *models.SubjectObject) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	subjectObject.ID = len(r.subjectObjects) + 1
	r.subjectObjects[subjectObject.ID] = *subjectObject
	return nil
}

func (r *fakeSubjectRepo) UpdateSubjectObject(ctx context.Context, subjectObject *models.SubjectObject) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := r.subjectObjects[subjectObject.ID]
	stored.Name = subjectObject.Name
	stored.Href = subjectObject.Href
	r.subjectObjects[subjectObject.ID] = stored
	return nil
}

func (r *fakeSubjectRepo) UpdateSubjectObjectContent(ctx context.Context, subjectObject *models.SubjectObject) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := r.subjectObjects[subjectObject.ID]
	stored.Content = subjectObject.Content
	stored.ContentETag = subjectObject.ContentETag
	stored.ContentFetchedAt = subjectObject.ContentFetchedAt
	stored.ContentCheckedAt = subjectObject.ContentCheckedAt
	stored.ContentError = subjectObject.ContentError
	stored.ContentUnsupported = subjectObject.ContentUnsupported
	r.subjectObjects[subjectObject.ID] = stored
	return nil
}

func (r *fakeSubjectRepo) stored(id int) models.SubjectObject {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.subjectObjects[id]
}

type fakeFetcher struct {
	mu      sync.Mutex
	content *educational.Content
	err     error
	calls   []string // etags passed to Fetch
	fetched chan struct{}
}

func (f *fakeFetcher) Supports(href string) bool {
	return strings.HasPrefix(href, "https://1drv.ms/")
}

func (f *fakeFetcher) Fetch(ctx context.Context, href string, etag string) (*educational.Content, error) {
	f.mu.Lock()
	f.calls = append(f.calls, etag)
	f.mu.Unlock()
	if f.fetched != nil {
		defer func() { f.fetched <- struct{}{} }()
	}
	if f.err != nil {
		return nil, f.err
	}
	if etag != "" && etag == f.content.ETag {
		return nil, educational.ErrContentNotModified
	}
	return f.content, nil
}

func (f *fakeFetcher) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

var testNow = time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)

func newTestUseCase(subjectObject models.SubjectObject, fetcher *fakeFetcher) (*SubjectUseCase, *fakeSubjectRepo) {
	repo := &fakeSubjectRepo{subjectObjects: map[int]models.SubjectObject{subjectObject.ID: subjectObject}}
	uc := NewSubjectUseCase(repo, fetcher)
	uc.now = func() time.Time { return testNow }
	return uc, repo
}

func timeAgo(d time.Duration) *time.Time {
	t := testNow.Add(-d)
	return &t
}

func TestGetTaskFirstViewDownloadsDocument(t *testing.T) {
	fetcher := &fakeFetcher{content: &educational.Content{Body: []byte("# Задание"), ETag: "v1"}}
	uc, repo := newTestUseCase(models.SubjectObject{ID: 1, Href: oneDriveHref}, fetcher)

	subjectObject, err := uc.GetTask(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if subjectObject.Content != "# Задание" {
		t.Errorf("Content = %q, want downloaded document", subjectObject.Content)
	}

	stored := repo.stored(1)
	if stored.Content != "# Задание" || stored.ContentETag != "v1" || stored.ContentFetchedAt == nil || stored.ContentCheckedAt == nil {
		t.Errorf("stored subject object = %+v, want downloaded document with timestamps", stored)
	}
}

func TestGetTaskFreshDocumentIsNotDownloaded(t *testing.T) {
	fetcher := &fakeFetcher{content: &educational.Content{Body: []byte("new"), ETag: "v2"}}
	uc, _ := newTestUseCase(models.SubjectObject{
		ID: 1, Href: oneDriveHref, Content: "old", ContentETag: "v1",
		ContentFetchedAt: timeAgo(time.Minute), ContentCheckedAt: timeAgo(time.Minute),
	}, fetcher)

	subjectObject, err := uc.GetTask(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if subjectObject.Content != "old" || fetcher.callCount() != 0 {
		t.Errorf("Content = %q, fetch calls = %d; want stored document without download", subjectObject.Content, fetcher.callCount())
	}
}

func TestGetTaskOutdatedDocumentIsShownAndRefreshedInBackground(t *testing.T) {
	fetcher := &fakeFetcher{
		content: &educational.Content{Body: []byte("new"), ETag: "v2"},
		fetched: make(chan struct{}, 1),
	}
	uc, repo := newTestUseCase(models.SubjectObject{
		ID: 1, Href: oneDriveHref, Content: "old", ContentETag: "v1",
		ContentFetchedAt: timeAgo(time.Hour), ContentCheckedAt: timeAgo(time.Hour),
	}, fetcher)

	subjectObject, err := uc.GetTask(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if subjectObject.Content != "old" {
		t.Errorf("Content = %q, want stored document shown right away", subjectObject.Content)
	}

	<-fetcher.fetched
	waitFor(t, func() bool { return repo.stored(1).Content == "new" })
	if fetcher.calls[0] != "v1" {
		t.Errorf("Fetch() etag = %q, want stored etag to skip unchanged documents", fetcher.calls[0])
	}
}

func TestGetTaskKeepsDocumentWhenSourceFails(t *testing.T) {
	fetcher := &fakeFetcher{err: errors.New("onedrive is down"), fetched: make(chan struct{}, 1)}
	uc, repo := newTestUseCase(models.SubjectObject{
		ID: 1, Href: oneDriveHref, Content: "old", ContentETag: "v1",
		ContentFetchedAt: timeAgo(time.Hour), ContentCheckedAt: timeAgo(time.Hour),
	}, fetcher)

	if _, err := uc.GetTask(context.Background(), 1); err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	<-fetcher.fetched
	waitFor(t, func() bool { return repo.stored(1).ContentError != "" })
	if stored := repo.stored(1); stored.Content != "old" || !stored.ContentCheckedAt.Equal(testNow) {
		t.Errorf("stored subject object = %+v, want old document kept and check time updated", stored)
	}
}

func TestGetTaskFailedFirstDownloadIsNotRetriedOnEveryView(t *testing.T) {
	fetcher := &fakeFetcher{err: errors.New("onedrive is down")}
	uc, repo := newTestUseCase(models.SubjectObject{ID: 1, Href: oneDriveHref}, fetcher)

	subjectObject, err := uc.GetTask(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if subjectObject.Content != "" || subjectObject.ContentError == "" {
		t.Errorf("subject object = %+v, want empty content with error", subjectObject)
	}

	repoObject := repo.stored(1)
	uc.now = func() time.Time { return repoObject.ContentCheckedAt.Add(time.Minute) }
	if _, err := uc.GetTask(context.Background(), 1); err != nil {
		t.Fatalf("second GetTask() error = %v", err)
	}
	if fetcher.callCount() != 1 {
		t.Errorf("fetch calls = %d, want 1", fetcher.callCount())
	}
}

func TestGetTaskNotDocument(t *testing.T) {
	uc, _ := newTestUseCase(models.SubjectObject{ID: 1, Href: "https://example.com/task"}, &fakeFetcher{})

	subjectObject, err := uc.GetTask(context.Background(), 1)
	if !errors.Is(err, educational.ErrSubjectObjectNotDocument) || subjectObject.Href != "https://example.com/task" {
		t.Fatalf("GetTask() = %+v, %v; want link with ErrSubjectObjectNotDocument", subjectObject, err)
	}
}

func TestGetTaskUnsupportedFile(t *testing.T) {
	fetcher := &fakeFetcher{err: educational.ErrContentUnsupported}
	uc, repo := newTestUseCase(models.SubjectObject{ID: 1, Href: oneDriveHref}, fetcher)

	if _, err := uc.GetTask(context.Background(), 1); !errors.Is(err, educational.ErrSubjectObjectNotDocument) {
		t.Fatalf("GetTask() error = %v, want ErrSubjectObjectNotDocument", err)
	}
	if !repo.stored(1).ContentUnsupported {
		t.Fatal("unsupported file is not remembered")
	}

	if _, err := uc.GetTask(context.Background(), 1); !errors.Is(err, educational.ErrSubjectObjectNotDocument) {
		t.Fatalf("second GetTask() error = %v, want ErrSubjectObjectNotDocument", err)
	}
	if fetcher.callCount() != 1 {
		t.Errorf("fetch calls = %d, want unsupported file not to be downloaded again", fetcher.callCount())
	}
}

func TestGetTaskNotFound(t *testing.T) {
	uc, _ := newTestUseCase(models.SubjectObject{ID: 1, Href: oneDriveHref}, &fakeFetcher{})

	if _, err := uc.GetTask(context.Background(), 2); !errors.Is(err, educational.ErrSubjectObjectNotFound) {
		t.Fatalf("GetTask() error = %v, want ErrSubjectObjectNotFound", err)
	}
}

func TestRefreshSubjectObjectContentForcesDownload(t *testing.T) {
	fetcher := &fakeFetcher{content: &educational.Content{Body: []byte("same"), ETag: "v1"}}
	uc, repo := newTestUseCase(models.SubjectObject{
		ID: 1, SubjectId: 7, Href: oneDriveHref, Content: "stale", ContentETag: "v1",
		ContentFetchedAt: timeAgo(time.Minute), ContentCheckedAt: timeAgo(time.Minute),
	}, fetcher)

	if _, err := uc.RefreshSubjectObjectContent(context.Background(), 7, 1); err != nil {
		t.Fatalf("RefreshSubjectObjectContent() error = %v", err)
	}
	if fetcher.calls[0] != "" || repo.stored(1).Content != "same" {
		t.Errorf("fetch etag = %q, content = %q; want forced download", fetcher.calls[0], repo.stored(1).Content)
	}

	if _, err := uc.RefreshSubjectObjectContent(context.Background(), 8, 1); !errors.Is(err, educational.ErrSubjectObjectNotFound) {
		t.Errorf("refresh with another subject error = %v, want ErrSubjectObjectNotFound", err)
	}
}

func TestUpdateSubjectObjectResetsContentWhenLinkChanges(t *testing.T) {
	uc, repo := newTestUseCase(models.SubjectObject{
		ID: 1, SubjectId: 7, Name: "Задание", Href: oneDriveHref, Content: "old", ContentETag: "v1",
		ContentFetchedAt: timeAgo(time.Minute), ContentCheckedAt: timeAgo(time.Minute),
	}, &fakeFetcher{})

	if _, err := uc.UpdateSubjectObject(context.Background(), 7, 1, "Задание 2", oneDriveHref); err != nil {
		t.Fatalf("UpdateSubjectObject() error = %v", err)
	}
	if repo.stored(1).Content != "old" {
		t.Error("content is reset although the link has not changed")
	}

	if _, err := uc.UpdateSubjectObject(context.Background(), 7, 1, "Задание 2", "https://example.com/other"); err != nil {
		t.Fatalf("UpdateSubjectObject() error = %v", err)
	}
	if stored := repo.stored(1); stored.Content != "" || stored.ContentETag != "" || stored.ContentCheckedAt != nil {
		t.Errorf("stored subject object = %+v, want content reset for the new link", stored)
	}
}

func TestUpdateSubjectObjectDownloadsNewDocument(t *testing.T) {
	fetcher := &fakeFetcher{
		content: &educational.Content{Body: []byte("new document"), ETag: "n1"},
		fetched: make(chan struct{}, 1),
	}
	uc, repo := newTestUseCase(models.SubjectObject{
		ID: 1, SubjectId: 7, Name: "Задание", Href: oneDriveHref, Content: "old", ContentETag: "v1",
		ContentFetchedAt: timeAgo(time.Minute), ContentCheckedAt: timeAgo(time.Minute),
	}, fetcher)

	if _, err := uc.UpdateSubjectObject(context.Background(), 7, 1, "Задание", "https://1drv.ms/t/c/1/other"); err != nil {
		t.Fatalf("UpdateSubjectObject() error = %v", err)
	}

	<-fetcher.fetched
	waitFor(t, func() bool { return repo.stored(1).Content == "new document" })
	if fetcher.calls[0] != "" {
		t.Errorf("Fetch() etag = %q, want the old link's etag not to be used", fetcher.calls[0])
	}
}

func TestCreateSubjectObjectDownloadsDocument(t *testing.T) {
	fetcher := &fakeFetcher{
		content: &educational.Content{Body: []byte("# Задание"), ETag: "v1"},
		fetched: make(chan struct{}, 1),
	}
	uc, repo := newTestUseCase(models.SubjectObject{ID: 1, SubjectId: 7, Href: "https://example.com/task"}, fetcher)

	subjectObject, err := uc.CreateSubjectObject(context.Background(), "Новое задание", 7, oneDriveHref)
	if err != nil {
		t.Fatalf("CreateSubjectObject() error = %v", err)
	}

	<-fetcher.fetched
	waitFor(t, func() bool { return repo.stored(subjectObject.ID).Content == "# Задание" })

	if _, err := uc.CreateSubjectObject(context.Background(), "Ссылка", 7, "https://example.com/other"); err != nil {
		t.Fatalf("CreateSubjectObject() error = %v", err)
	}
	if fetcher.callCount() != 1 {
		t.Errorf("fetch calls = %d, want external links not to be downloaded", fetcher.callCount())
	}
}

func waitFor(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatal("condition was not met in time")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
