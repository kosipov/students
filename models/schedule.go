package models

import "time"

// ScheduleLesson is a lesson of the teacher copied from the university's schedule. Times are stored in UTC.
type ScheduleLesson struct {
	ID int
	// WeekStart is Monday of the week the lesson belongs to; lessons are replaced week by week.
	WeekStart time.Time `gorm:"type:date;not null;index"`
	Number    int       `gorm:"not null"`
	StartsAt  time.Time `gorm:"not null;index"`
	EndsAt    time.Time `gorm:"not null"`
	Title     string    `gorm:"type:varchar(255);not null"`
	// Kind is written as on the university's site: "л.", "пр.", "лаб.".
	Kind string `gorm:"type:varchar(50);not null;default:''"`
	// Location is the room as written on the site, e.g. "245/1"; Room and Building are its parts.
	Location string `gorm:"type:varchar(100);not null;default:''"`
	Room     string `gorm:"type:varchar(50);not null;default:''"`
	Building string `gorm:"type:varchar(50);not null;default:''"`
	Groups   string `gorm:"type:varchar(255);not null;default:''"`
}

// ScheduleSync is the state of copying the schedule from the university's site; there is one row.
type ScheduleSync struct {
	ID          int
	CheckedAt   *time.Time
	SucceededAt *time.Time
	// CampusUpdatedAt is when the university last changed schedules, as its site says.
	CampusUpdatedAt *time.Time
	Error           string `gorm:"type:text"`
}
