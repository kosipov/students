package educational

import "github.com/kosipov/students/models"

// Task is a subject object together with the subject and group it belongs to.
type Task struct {
	SubjectObject *models.SubjectObject
	Subject       *models.Subject
	Group         *models.Group
}

// TaskAccess tells which password-protected tasks the student has opened.
type TaskAccess interface {
	// IsUnlocked reports whether the student entered the password of this version of the task.
	IsUnlocked(subjectObjectId int, passwordVersion int) bool
}

// Overview summarizes all groups for the admin panel.
type Overview struct {
	Groups      int
	Subjects    int
	Tasks       int
	HiddenTasks int
	// RecentlyUpdated lists tasks whose documents were downloaded most recently, newest first.
	RecentlyUpdated []Task
	// Issues lists tasks whose documents failed to download.
	Issues []Task
}
