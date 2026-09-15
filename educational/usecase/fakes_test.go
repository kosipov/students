package usecase

import (
	"context"
	"github.com/jinzhu/gorm"
	"github.com/kosipov/students/educational"
	"github.com/kosipov/students/models"
	"strings"
	"sync"
	"time"
)

const (
	oneDriveHref  = "https://1drv.ms/t/c/1/abc"
	testGroupId   = 1
	testSubjectId = 7
)

var testNow = time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)

// taskAccess lists unlocked task ids with password versions.
type taskAccess map[int]int

func (a taskAccess) IsUnlocked(id int, version int) bool {
	v, ok := a[id]
	return ok && v == version
}

var noAccess = taskAccess{}

func timeAgo(d time.Duration) *time.Time {
	t := testNow.Add(-d)
	return &t
}

// fakeRepo keeps groups, subjects and subject objects in memory for both repositories.
type fakeRepo struct {
	mu             sync.Mutex
	groups         map[int]models.Group
	subjects       map[int]models.Subject
	subjectObjects map[int]models.SubjectObject
	nextId         int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		groups:         map[int]models.Group{},
		subjects:       map[int]models.Subject{},
		subjectObjects: map[int]models.SubjectObject{},
		nextId:         100,
	}
}

func (r *fakeRepo) id() int {
	r.nextId++
	return r.nextId
}

func (r *fakeRepo) addGroup(group models.Group) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.groups[int(group.ID)] = group
}

func (r *fakeRepo) addSubject(subject models.Subject) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.subjects[subject.ID] = subject
}

func (r *fakeRepo) addSubjectObject(subjectObject models.SubjectObject) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.subjectObjects[subjectObject.ID] = subjectObject
}

func (r *fakeRepo) stored(id int) models.SubjectObject {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.subjectObjects[id]
}

func (r *fakeRepo) setGroupHidden(id int, hidden bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	group := r.groups[id]
	group.Hidden = hidden
	r.groups[id] = group
}

// tree builds a group with its subjects and subject objects ordered by id. The caller holds the lock.
func (r *fakeRepo) tree(group models.Group) models.Group {
	group.Subjects = nil
	for id := 1; id <= r.nextId; id++ {
		subject, ok := r.subjects[id]
		if !ok || subject.GroupId != int(group.ID) {
			continue
		}
		for objectId := 1; objectId <= r.nextId; objectId++ {
			if subjectObject, ok := r.subjectObjects[objectId]; ok && subjectObject.SubjectId == subject.ID {
				subjectObject.Content = ""
				subject.SubjectObjects = append(subject.SubjectObjects, subjectObject)
			}
		}
		group.Subjects = append(group.Subjects, subject)
	}
	return group
}

// Group repository

func (r *fakeRepo) GetGroups(ctx context.Context) (*[]models.Group, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var groups []models.Group
	for id := 1; id <= r.nextId; id++ {
		if group, ok := r.groups[id]; ok {
			groups = append(groups, r.tree(group))
		}
	}
	return &groups, nil
}

func (r *fakeRepo) GetVisibleGroups(ctx context.Context) (*[]models.Group, error) {
	all, _ := r.GetGroups(ctx)
	var groups []models.Group
	for _, group := range *all {
		if group.Hidden {
			continue
		}
		for si := range group.Subjects {
			var visible []models.SubjectObject
			for _, subjectObject := range group.Subjects[si].SubjectObjects {
				if !subjectObject.Hidden {
					visible = append(visible, subjectObject)
				}
			}
			group.Subjects[si].SubjectObjects = visible
		}
		groups = append(groups, group)
	}
	return &groups, nil
}

func (r *fakeRepo) GetGroupById(ctx context.Context, id int) (*models.Group, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	group, ok := r.groups[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	group = r.tree(group)
	return &group, nil
}

func (r *fakeRepo) CreateGroup(ctx context.Context, group *models.Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	group.ID = uint16(r.id())
	r.groups[int(group.ID)] = *group
	return nil
}

func (r *fakeRepo) UpdateGroup(ctx context.Context, group *models.Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := r.groups[int(group.ID)]
	stored.GroupName = group.GroupName
	stored.Hidden = group.Hidden
	r.groups[int(group.ID)] = stored
	return nil
}

func (r *fakeRepo) DeleteGroup(ctx context.Context, group *models.Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, subject := range r.subjects {
		if subject.GroupId == int(group.ID) {
			r.deleteSubject(id)
		}
	}
	delete(r.groups, int(group.ID))
	return nil
}

// Subject repository

func (r *fakeRepo) GetGroup(ctx context.Context, groupId int) (*models.Group, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	group, ok := r.groups[groupId]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &group, nil
}

func (r *fakeRepo) GetSubject(ctx context.Context, id int) (*models.Subject, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	subject, ok := r.subjects[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &subject, nil
}

func (r *fakeRepo) GetSubjectWithSubjectObjects(ctx context.Context, id int) (*models.Subject, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	subject, ok := r.subjects[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	for _, s := range r.tree(r.groups[subject.GroupId]).Subjects {
		if s.ID == id {
			subject = s
		}
	}
	subject.Group = r.groups[subject.GroupId]
	return &subject, nil
}

func (r *fakeRepo) CreateSubject(ctx context.Context, subject *models.Subject) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	subject.ID = r.id()
	r.subjects[subject.ID] = *subject
	return nil
}

func (r *fakeRepo) UpdateSubject(ctx context.Context, subject *models.Subject) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := r.subjects[subject.ID]
	stored.SubjectName = subject.SubjectName
	r.subjects[subject.ID] = stored
	return nil
}

func (r *fakeRepo) DeleteSubject(ctx context.Context, subject *models.Subject) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deleteSubject(subject.ID)
	return nil
}

func (r *fakeRepo) deleteSubject(id int) {
	for objectId, subjectObject := range r.subjectObjects {
		if subjectObject.SubjectId == id {
			delete(r.subjectObjects, objectId)
		}
	}
	delete(r.subjects, id)
}

func (r *fakeRepo) GetSubjectObject(ctx context.Context, id int) (*models.SubjectObject, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	subjectObject, ok := r.subjectObjects[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &subjectObject, nil
}

func (r *fakeRepo) CreateSubjectObject(ctx context.Context, subjectObject *models.SubjectObject) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	subjectObject.ID = r.id()
	r.subjectObjects[subjectObject.ID] = *subjectObject
	return nil
}

func (r *fakeRepo) UpdateSubjectObject(ctx context.Context, subjectObject *models.SubjectObject) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := r.subjectObjects[subjectObject.ID]
	stored.Name = subjectObject.Name
	stored.Href = subjectObject.Href
	stored.Comment = subjectObject.Comment
	stored.Hidden = subjectObject.Hidden
	stored.Password = subjectObject.Password
	stored.PasswordVersion = subjectObject.PasswordVersion
	stored.Categories = append([]models.SubjectObjectCategory(nil), subjectObject.Categories...)
	r.subjectObjects[subjectObject.ID] = stored
	return nil
}

func (r *fakeRepo) GetCategoryNames(ctx context.Context) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	seen := map[string]bool{}
	var names []string
	for id := 1; id <= r.nextId; id++ {
		for _, category := range r.subjectObjects[id].Categories {
			if !seen[category.Name] {
				seen[category.Name] = true
				names = append(names, category.Name)
			}
		}
	}
	return names, nil
}

func (r *fakeRepo) UpdateSubjectObjectContent(ctx context.Context, subjectObject *models.SubjectObject) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored, ok := r.subjectObjects[subjectObject.ID]
	if !ok || stored.Href != subjectObject.Href {
		// Same as the database: the document of an old link is not stored.
		return nil
	}
	stored.Content = subjectObject.Content
	stored.ContentETag = subjectObject.ContentETag
	stored.ContentFetchedAt = subjectObject.ContentFetchedAt
	stored.ContentCheckedAt = subjectObject.ContentCheckedAt
	stored.ContentError = subjectObject.ContentError
	stored.ContentUnsupported = subjectObject.ContentUnsupported
	r.subjectObjects[subjectObject.ID] = stored
	return nil
}

func (r *fakeRepo) DeleteSubjectObject(ctx context.Context, subjectObject *models.SubjectObject) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.subjectObjects, subjectObject.ID)
	return nil
}

var (
	_ educational.CommonGroupRepository   = (*fakeRepo)(nil)
	_ educational.CommonSubjectRepository = (*fakeRepo)(nil)
)

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

func (f *fakeFetcher) call(i int) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[i]
}

// newTestUseCase stores the subject object with a visible subject (testSubjectId) and group (testGroupId).
func newTestUseCase(subjectObject models.SubjectObject, fetcher *fakeFetcher) (*SubjectUseCase, *fakeRepo) {
	repo := newFakeRepo()
	repo.addGroup(models.Group{ID: testGroupId, GroupName: "ИСП-301"})
	repo.addSubject(models.Subject{ID: testSubjectId, SubjectName: "Веб-разработка", GroupId: testGroupId})
	if subjectObject.SubjectId == 0 {
		subjectObject.SubjectId = testSubjectId
	}
	repo.addSubjectObject(subjectObject)

	uc := NewSubjectUseCase(repo, fetcher)
	uc.now = func() time.Time { return testNow }
	return uc, repo
}
