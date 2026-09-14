package models

import "time"

type SubjectObject struct {
	ID        int
	Name      string
	Comment   string
	Href      string
	SubjectId int
	Subject   Subject
	// Hidden subject objects are shown only in the admin panel.
	Hidden     bool                    `gorm:"not null;default:false"`
	Categories []SubjectObjectCategory `gorm:"foreignkey:SubjectObjectId"`

	// Stored copy of the markdown document Href points to (e.g. a file shared from OneDrive).
	// It is shown to students even when the source becomes unavailable.
	Content            string     `gorm:"column:content;type:mediumtext"`
	ContentETag        string     `gorm:"column:content_etag"`
	ContentFetchedAt   *time.Time `gorm:"column:content_fetched_at"`
	ContentCheckedAt   *time.Time `gorm:"column:content_checked_at"`
	ContentError       string     `gorm:"column:content_error;type:text"`
	ContentUnsupported bool       `gorm:"column:content_unsupported"`
}
