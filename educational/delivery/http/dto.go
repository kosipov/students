package http

import (
	"github.com/kosipov/students/educational"
	"github.com/kosipov/students/models"
	"time"
)

type ref struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Public API

type catalogResponse struct {
	Groups []catalogGroup `json:"groups"`
}

type catalogGroup struct {
	ID       int              `json:"id"`
	Name     string           `json:"name"`
	Subjects []catalogSubject `json:"subjects"`
}

type catalogSubject struct {
	ID    int           `json:"id"`
	Name  string        `json:"name"`
	Tasks []catalogTask `json:"tasks"`
}

type catalogTask struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Comment string `json:"comment"`
	Href    string `json:"href"`
	// IsDocument marks tasks shown as a page on the site; other tasks are opened by href.
	IsDocument bool     `json:"isDocument"`
	Categories []string `json:"categories"`
}

type taskResponse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Comment    string `json:"comment"`
	Href       string `json:"href"`
	Hidden     bool   `json:"hidden"`
	IsDocument bool   `json:"isDocument"`
	Group      ref    `json:"group"`
	Subject    ref    `json:"subject"`
	// ContentHTML is the rendered document, null when it has never been downloaded.
	ContentHTML *string    `json:"contentHtml"`
	UpdatedAt   *time.Time `json:"updatedAt"`
}

// Admin API

type adminGroup struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Hidden        bool   `json:"hidden"`
	SubjectsCount int    `json:"subjectsCount"`
	TasksCount    int    `json:"tasksCount"`
}

type adminSubject struct {
	ID         int    `json:"id"`
	GroupID    int    `json:"groupId"`
	Name       string `json:"name"`
	TasksCount int    `json:"tasksCount"`
}

type adminTask struct {
	ID                 int        `json:"id"`
	SubjectID          int        `json:"subjectId"`
	Name               string     `json:"name"`
	Href               string     `json:"href"`
	Comment            string     `json:"comment"`
	Hidden             bool       `json:"hidden"`
	Categories         []string   `json:"categories"`
	IsDocument         bool       `json:"isDocument"`
	ContentUnsupported bool       `json:"contentUnsupported"`
	FetchedAt          *time.Time `json:"fetchedAt"`
	CheckedAt          *time.Time `json:"checkedAt"`
	Error              string     `json:"error"`
}

type adminGroupsResponse struct {
	Groups []adminGroup `json:"groups"`
}

type adminGroupResponse struct {
	Group    adminGroup     `json:"group"`
	Subjects []adminSubject `json:"subjects"`
}

type adminSubjectResponse struct {
	Group   adminGroup   `json:"group"`
	Subject adminSubject `json:"subject"`
	Tasks   []adminTask  `json:"tasks"`
}

type overviewResponse struct {
	Stats           overviewStats  `json:"stats"`
	RecentlyUpdated []overviewTask `json:"recentlyUpdated"`
	Issues          []overviewTask `json:"issues"`
}

type overviewStats struct {
	Groups   int `json:"groups"`
	Subjects int `json:"subjects"`
	Tasks    int `json:"tasks"`
	Hidden   int `json:"hidden"`
}

type overviewTask struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	GroupName   string     `json:"groupName"`
	SubjectID   int        `json:"subjectId"`
	SubjectName string     `json:"subjectName"`
	FetchedAt   *time.Time `json:"fetchedAt"`
	Error       string     `json:"error"`
}

// Requests

type groupRequest struct {
	Name   *string `json:"name"`
	Hidden *bool   `json:"hidden"`
}

type subjectRequest struct {
	Name string `json:"name"`
}

type taskRequest struct {
	Name       *string   `json:"name"`
	Href       *string   `json:"href"`
	Comment    *string   `json:"comment"`
	Hidden     *bool     `json:"hidden"`
	Categories *[]string `json:"categories"`
}

type categoriesResponse struct {
	Categories []string `json:"categories"`
}

// Mapping

// safeHref hides links that can't be put into a page, e.g. javascript: URLs stored before validation existed.
func safeHref(href string) string {
	if educational.IsSafeHref(href) {
		return href
	}
	return ""
}

func valueOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func toAdminGroup(group *models.Group) adminGroup {
	result := adminGroup{
		ID:            int(group.ID),
		Name:          group.GroupName,
		Hidden:        group.Hidden,
		SubjectsCount: len(group.Subjects),
	}
	for _, subject := range group.Subjects {
		result.TasksCount += len(subject.SubjectObjects)
	}
	return result
}

func toAdminSubject(subject *models.Subject) adminSubject {
	return adminSubject{
		ID:         subject.ID,
		GroupID:    subject.GroupId,
		Name:       subject.SubjectName,
		TasksCount: len(subject.SubjectObjects),
	}
}

func (h *Handler) toAdminTask(subjectObject *models.SubjectObject) adminTask {
	return adminTask{
		ID:                 subjectObject.ID,
		SubjectID:          subjectObject.SubjectId,
		Name:               subjectObject.Name,
		Href:               safeHref(subjectObject.Href),
		Comment:            subjectObject.Comment,
		Hidden:             subjectObject.Hidden,
		Categories:         subjectObject.CategoryNames(),
		IsDocument:         h.subjectUseCase.IsDocument(subjectObject),
		ContentUnsupported: subjectObject.ContentUnsupported,
		FetchedAt:          subjectObject.ContentFetchedAt,
		CheckedAt:          subjectObject.ContentCheckedAt,
		Error:              subjectObject.ContentError,
	}
}

func toOverviewTasks(tasks []educational.Task) []overviewTask {
	result := make([]overviewTask, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, overviewTask{
			ID:          task.SubjectObject.ID,
			Name:        task.SubjectObject.Name,
			GroupName:   task.Group.GroupName,
			SubjectID:   task.Subject.ID,
			SubjectName: task.Subject.SubjectName,
			FetchedAt:   task.SubjectObject.ContentFetchedAt,
			Error:       task.SubjectObject.ContentError,
		})
	}
	return result
}
