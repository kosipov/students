package models

import "time"

// Student is a member of a group whose lab progress the teacher tracks.
type Student struct {
	ID      int
	Name    string `gorm:"type:varchar(120);not null"`
	GroupId int    `gorm:"not null;index"`
	Group   Group
	Marks   []Mark `gorm:"foreignkey:StudentId"`
}

// Mark is the status a student has for one subject object (a lab work).
// A row exists only while a status is set; an empty cell has no row.
type Mark struct {
	ID              int
	StudentId       int    `gorm:"not null;unique_index:idx_marks_student_task"`
	SubjectObjectId int    `gorm:"not null;unique_index:idx_marks_student_task"`
	Status          string `gorm:"type:varchar(20);not null"`
	// Comment is a short note for the student, e.g. what to fix. Shown on the public page.
	Comment   string `gorm:"type:varchar(500);not null;default:''"`
	UpdatedAt time.Time
}
