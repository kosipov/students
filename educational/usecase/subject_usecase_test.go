package usecase

import (
	"context"
	"errors"
	"github.com/kosipov/students/educational"
	"github.com/kosipov/students/models"
	"testing"
	"time"
)

// Documents

func TestGetTaskFirstViewDownloadsDocument(t *testing.T) {
	fetcher := &fakeFetcher{content: &educational.Content{Body: []byte("# Задание"), ETag: "v1"}}
	uc, repo := newTestUseCase(models.SubjectObject{ID: 1, Href: oneDriveHref}, fetcher)

	task, err := uc.GetTask(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if task.SubjectObject.Content != "# Задание" {
		t.Errorf("Content = %q, want downloaded document", task.SubjectObject.Content)
	}
	if task.Subject.ID != testSubjectId || int(task.Group.ID) != testGroupId {
		t.Errorf("task = %+v, want its subject and group", task)
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

	task, err := uc.GetTask(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if task.SubjectObject.Content != "old" || fetcher.callCount() != 0 {
		t.Errorf("Content = %q, fetch calls = %d; want stored document without download", task.SubjectObject.Content, fetcher.callCount())
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

	task, err := uc.GetTask(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if task.SubjectObject.Content != "old" {
		t.Errorf("Content = %q, want stored document shown right away", task.SubjectObject.Content)
	}

	<-fetcher.fetched
	waitFor(t, func() bool { return repo.stored(1).Content == "new" })
	if fetcher.call(0) != "v1" {
		t.Errorf("Fetch() etag = %q, want stored etag to skip unchanged documents", fetcher.call(0))
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

	task, err := uc.GetTask(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if task.SubjectObject.Content != "" || task.SubjectObject.ContentError == "" {
		t.Errorf("subject object = %+v, want empty content with error", task.SubjectObject)
	}

	checkedAt := *repo.stored(1).ContentCheckedAt
	uc.now = func() time.Time { return checkedAt.Add(time.Minute) }
	if _, err := uc.GetTask(context.Background(), 1); err != nil {
		t.Fatalf("second GetTask() error = %v", err)
	}
	if fetcher.callCount() != 1 {
		t.Errorf("fetch calls = %d, want 1", fetcher.callCount())
	}
}

func TestGetTaskNotDocument(t *testing.T) {
	uc, _ := newTestUseCase(models.SubjectObject{ID: 1, Href: "https://example.com/task"}, &fakeFetcher{})

	task, err := uc.GetTask(context.Background(), 1)
	if !errors.Is(err, educational.ErrSubjectObjectNotDocument) || task.SubjectObject.Href != "https://example.com/task" {
		t.Fatalf("GetTask() = %+v, %v; want link with ErrSubjectObjectNotDocument", task, err)
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
	uc, repo := newTestUseCase(models.SubjectObject{ID: 1, Href: oneDriveHref}, &fakeFetcher{})

	if _, err := uc.GetTask(context.Background(), 2); !errors.Is(err, educational.ErrSubjectObjectNotFound) {
		t.Errorf("GetTask() for missing task error = %v, want ErrSubjectObjectNotFound", err)
	}

	repo.addSubjectObject(models.SubjectObject{ID: 3, SubjectId: 404})
	if _, err := uc.GetTask(context.Background(), 3); !errors.Is(err, educational.ErrSubjectObjectNotFound) {
		t.Errorf("GetTask() for task without subject error = %v, want ErrSubjectObjectNotFound", err)
	}
}

func TestRefreshSubjectObjectContentForcesDownload(t *testing.T) {
	fetcher := &fakeFetcher{content: &educational.Content{Body: []byte("same"), ETag: "v1"}}
	uc, repo := newTestUseCase(models.SubjectObject{
		ID: 1, Href: oneDriveHref, Content: "stale", ContentETag: "v1",
		ContentFetchedAt: timeAgo(time.Minute), ContentCheckedAt: timeAgo(time.Minute),
	}, fetcher)

	subjectObject, err := uc.RefreshSubjectObjectContent(context.Background(), 1)
	if err != nil {
		t.Fatalf("RefreshSubjectObjectContent() error = %v", err)
	}
	if fetcher.call(0) != "" || repo.stored(1).Content != "same" || subjectObject.Content != "same" {
		t.Errorf("fetch etag = %q, content = %q; want forced download", fetcher.call(0), repo.stored(1).Content)
	}
}

func TestRefreshSubjectObjectContentOfLink(t *testing.T) {
	uc, _ := newTestUseCase(models.SubjectObject{ID: 1, Href: "https://example.com/task"}, &fakeFetcher{})

	if _, err := uc.RefreshSubjectObjectContent(context.Background(), 1); !errors.Is(err, educational.ErrSubjectObjectNotDocument) {
		t.Errorf("RefreshSubjectObjectContent() error = %v, want ErrSubjectObjectNotDocument", err)
	}
	if _, err := uc.RefreshSubjectObjectContent(context.Background(), 404); !errors.Is(err, educational.ErrSubjectObjectNotFound) {
		t.Errorf("RefreshSubjectObjectContent() for missing task error = %v, want ErrSubjectObjectNotFound", err)
	}
}

// Visibility

func TestGetTaskHidden(t *testing.T) {
	fetcher := &fakeFetcher{content: &educational.Content{Body: []byte("# Задание"), ETag: "v1"}}

	t.Run("hidden subject object", func(t *testing.T) {
		uc, _ := newTestUseCase(models.SubjectObject{ID: 1, Href: oneDriveHref, Hidden: true}, fetcher)
		if _, err := uc.GetTask(context.Background(), 1); !errors.Is(err, educational.ErrSubjectObjectNotFound) {
			t.Fatalf("GetTask() error = %v, want ErrSubjectObjectNotFound", err)
		}
	})

	t.Run("hidden group", func(t *testing.T) {
		uc, repo := newTestUseCase(models.SubjectObject{ID: 1, Href: "https://example.com/task"}, fetcher)
		repo.setGroupHidden(testGroupId, true)
		// Even external links must not be revealed.
		if task, err := uc.GetTask(context.Background(), 1); !errors.Is(err, educational.ErrSubjectObjectNotFound) || task != nil {
			t.Fatalf("GetTask() = %+v, %v; want ErrSubjectObjectNotFound without task", task, err)
		}
	})

	if fetcher.callCount() != 0 {
		t.Errorf("fetch calls = %d, hidden documents must not be downloaded for students", fetcher.callCount())
	}
}

func TestPreviewTaskShowsHidden(t *testing.T) {
	fetcher := &fakeFetcher{content: &educational.Content{Body: []byte("# Задание"), ETag: "v1"}}
	uc, repo := newTestUseCase(models.SubjectObject{ID: 1, Href: oneDriveHref, Hidden: true}, fetcher)
	repo.setGroupHidden(testGroupId, true)

	task, err := uc.PreviewTask(context.Background(), 1)
	if err != nil {
		t.Fatalf("PreviewTask() error = %v", err)
	}
	if task.SubjectObject.Content != "# Задание" {
		t.Errorf("Content = %q, want downloaded document", task.SubjectObject.Content)
	}
}

// Subject objects

func TestCreateSubjectObject(t *testing.T) {
	fetcher := &fakeFetcher{
		content: &educational.Content{Body: []byte("# Задание"), ETag: "v1"},
		fetched: make(chan struct{}, 1),
	}
	uc, repo := newTestUseCase(models.SubjectObject{ID: 1, Href: "https://example.com/task"}, fetcher)

	subjectObject, err := uc.CreateSubjectObject(context.Background(), testSubjectId, educational.SubjectObjectInput{
		Name: "  Новое задание ", Href: " " + oneDriveHref, Comment: " Пояснение ", Hidden: true,
	})
	if err != nil {
		t.Fatalf("CreateSubjectObject() error = %v", err)
	}
	stored := repo.stored(subjectObject.ID)
	if stored.Name != "Новое задание" || stored.Href != oneDriveHref || stored.Comment != "Пояснение" || !stored.Hidden {
		t.Errorf("stored subject object = %+v, want trimmed fields and hidden", stored)
	}

	// The document is downloaded before anyone opens the task.
	<-fetcher.fetched
	waitFor(t, func() bool { return repo.stored(subjectObject.ID).Content == "# Задание" })

	if _, err := uc.CreateSubjectObject(context.Background(), testSubjectId, educational.SubjectObjectInput{Name: "Ссылка", Href: "https://example.com/other"}); err != nil {
		t.Fatalf("CreateSubjectObject() error = %v", err)
	}
	if fetcher.callCount() != 1 {
		t.Errorf("fetch calls = %d, want external links not to be downloaded", fetcher.callCount())
	}
}

func TestCreateSubjectObjectValidation(t *testing.T) {
	uc, _ := newTestUseCase(models.SubjectObject{ID: 1}, &fakeFetcher{})

	_, err := uc.CreateSubjectObject(context.Background(), testSubjectId, educational.SubjectObjectInput{
		Name: " ", Href: "javascript:alert(1)",
	})
	var validationErr *educational.ValidationError
	if !errors.As(err, &validationErr) || validationErr.Fields["name"] == "" || validationErr.Fields["href"] == "" {
		t.Fatalf("CreateSubjectObject() error = %v, want name and href errors", err)
	}

	_, err = uc.CreateSubjectObject(context.Background(), 404, educational.SubjectObjectInput{Name: "Задание"})
	if !errors.Is(err, educational.ErrSubjectNotFound) {
		t.Errorf("CreateSubjectObject() for missing subject error = %v, want ErrSubjectNotFound", err)
	}
}

func TestUpdateSubjectObjectPatchesOnlySetFields(t *testing.T) {
	uc, repo := newTestUseCase(models.SubjectObject{
		ID: 1, Name: "Задание", Href: oneDriveHref, Comment: "Пояснение", Content: "old", ContentETag: "v1",
		ContentFetchedAt: timeAgo(time.Minute), ContentCheckedAt: timeAgo(time.Minute),
	}, &fakeFetcher{})

	hidden := true
	if _, err := uc.UpdateSubjectObject(context.Background(), 1, educational.SubjectObjectPatch{Hidden: &hidden}); err != nil {
		t.Fatalf("UpdateSubjectObject() error = %v", err)
	}
	stored := repo.stored(1)
	if !stored.Hidden || stored.Name != "Задание" || stored.Comment != "Пояснение" || stored.Content != "old" {
		t.Errorf("stored subject object = %+v, want only hidden changed", stored)
	}

	name, comment := "Задание 2", ""
	if _, err := uc.UpdateSubjectObject(context.Background(), 1, educational.SubjectObjectPatch{Name: &name, Comment: &comment}); err != nil {
		t.Fatalf("UpdateSubjectObject() error = %v", err)
	}
	if stored := repo.stored(1); stored.Name != "Задание 2" || stored.Comment != "" || stored.Content != "old" {
		t.Errorf("stored subject object = %+v, want name and cleared comment, content kept", stored)
	}

	empty := " "
	_, err := uc.UpdateSubjectObject(context.Background(), 1, educational.SubjectObjectPatch{Name: &empty})
	var validationErr *educational.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("UpdateSubjectObject() with empty name error = %v, want ValidationError", err)
	}
}

func TestUpdateSubjectObjectResetsContentWhenLinkChanges(t *testing.T) {
	uc, repo := newTestUseCase(models.SubjectObject{
		ID: 1, Name: "Задание", Href: oneDriveHref, Content: "old", ContentETag: "v1",
		ContentFetchedAt: timeAgo(time.Minute), ContentCheckedAt: timeAgo(time.Minute),
	}, &fakeFetcher{})

	href := "https://example.com/other"
	if _, err := uc.UpdateSubjectObject(context.Background(), 1, educational.SubjectObjectPatch{Href: &href}); err != nil {
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
		ID: 1, Name: "Задание", Href: oneDriveHref, Content: "old", ContentETag: "v1",
		ContentFetchedAt: timeAgo(time.Minute), ContentCheckedAt: timeAgo(time.Minute),
	}, fetcher)

	href := "https://1drv.ms/t/c/1/other"
	if _, err := uc.UpdateSubjectObject(context.Background(), 1, educational.SubjectObjectPatch{Href: &href}); err != nil {
		t.Fatalf("UpdateSubjectObject() error = %v", err)
	}

	<-fetcher.fetched
	waitFor(t, func() bool { return repo.stored(1).Content == "new document" })
	if fetcher.call(0) != "" {
		t.Errorf("Fetch() etag = %q, want the old link's etag not to be used", fetcher.call(0))
	}
}

func TestDownloadOfOldLinkIsNotStoredAfterLinkChange(t *testing.T) {
	release := make(chan struct{})
	fetcher := &blockingFetcher{release: release, started: make(chan struct{}, 2)}
	repo := newFakeRepo()
	repo.addGroup(models.Group{ID: testGroupId})
	repo.addSubject(models.Subject{ID: testSubjectId, GroupId: testGroupId})
	repo.addSubjectObject(models.SubjectObject{ID: 1, SubjectId: testSubjectId, Href: oneDriveHref})
	uc := NewSubjectUseCase(repo, fetcher)

	// A student opens the task, the download of the old link hangs.
	go func() { _, _ = uc.GetTask(context.Background(), 1) }()
	<-fetcher.started

	href := "https://example.com/other"
	if _, err := uc.UpdateSubjectObject(context.Background(), 1, educational.SubjectObjectPatch{Href: &href}); err != nil {
		t.Fatalf("UpdateSubjectObject() error = %v", err)
	}

	close(release)
	time.Sleep(50 * time.Millisecond)
	if stored := repo.stored(1); stored.Content != "" || stored.ContentCheckedAt != nil {
		t.Errorf("stored subject object = %+v, want the old link's document not stored", stored)
	}
}

type blockingFetcher struct {
	release chan struct{}
	started chan struct{}
}

func (f *blockingFetcher) Supports(href string) bool { return href == oneDriveHref }

func (f *blockingFetcher) Fetch(ctx context.Context, href string, etag string) (*educational.Content, error) {
	f.started <- struct{}{}
	<-f.release
	return &educational.Content{Body: []byte("old link document"), ETag: "old"}, nil
}

func TestDeleteSubjectObject(t *testing.T) {
	uc, repo := newTestUseCase(models.SubjectObject{ID: 1}, &fakeFetcher{})

	if err := uc.DeleteSubjectObject(context.Background(), 1); err != nil {
		t.Fatalf("DeleteSubjectObject() error = %v", err)
	}
	if _, err := repo.GetSubjectObject(context.Background(), 1); err == nil {
		t.Error("subject object is not deleted")
	}
	if err := uc.DeleteSubjectObject(context.Background(), 1); !errors.Is(err, educational.ErrSubjectObjectNotFound) {
		t.Errorf("second DeleteSubjectObject() error = %v, want ErrSubjectObjectNotFound", err)
	}
}

// Subjects

func TestSubjectLifecycle(t *testing.T) {
	uc, repo := newTestUseCase(models.SubjectObject{ID: 1}, &fakeFetcher{})
	ctx := context.Background()

	subject, err := uc.CreateSubject(ctx, testGroupId, educational.SubjectInput{Name: " Базы данных "})
	if err != nil || subject.SubjectName != "Базы данных" || subject.GroupId != testGroupId {
		t.Fatalf("CreateSubject() = %+v, %v", subject, err)
	}
	if _, err := uc.CreateSubject(ctx, 404, educational.SubjectInput{Name: "Предмет"}); !errors.Is(err, educational.ErrGroupNotFound) {
		t.Errorf("CreateSubject() for missing group error = %v, want ErrGroupNotFound", err)
	}

	if _, err := uc.UpdateSubject(ctx, subject.ID, educational.SubjectInput{Name: "СУБД"}); err != nil {
		t.Fatalf("UpdateSubject() error = %v", err)
	}
	withTasks, err := uc.GetSubjectWithSubjectObjects(ctx, testSubjectId)
	if err != nil || len(withTasks.SubjectObjects) != 1 || int(withTasks.Group.ID) != testGroupId {
		t.Errorf("GetSubjectWithSubjectObjects() = %+v, %v; want subject with its group and tasks", withTasks, err)
	}

	if err := uc.DeleteSubject(ctx, testSubjectId); err != nil {
		t.Fatalf("DeleteSubject() error = %v", err)
	}
	if _, err := repo.GetSubjectObject(ctx, 1); err == nil {
		t.Error("subject objects of the deleted subject are kept")
	}
	if _, err := uc.GetSubjectWithSubjectObjects(ctx, testSubjectId); !errors.Is(err, educational.ErrSubjectNotFound) {
		t.Errorf("GetSubjectWithSubjectObjects() after delete error = %v, want ErrSubjectNotFound", err)
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
