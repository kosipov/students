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
	// Password protects the document and the link: students enter it to open the task.
	// It is stored as is, because the admin panel shows it. Empty means no password.
	Password string `gorm:"type:varchar(100);not null;default:''"`
	// PasswordVersion grows with every password change, so an old password no longer opens the task.
	PasswordVersion int `gorm:"not null;default:0"`

	// Stored copy of the markdown document Href points to (e.g. a file shared from OneDrive).
	// It is shown to students even when the source becomes unavailable.
	Content            string     `gorm:"column:content;type:mediumtext"`
	ContentETag        string     `gorm:"column:content_etag"`
	ContentFetchedAt   *time.Time `gorm:"column:content_fetched_at"`
	ContentCheckedAt   *time.Time `gorm:"column:content_checked_at"`
	ContentError       string     `gorm:"column:content_error;type:text"`
	ContentUnsupported bool       `gorm:"column:content_unsupported"`
}

func (so SubjectObject) IsProtected() bool {
	return so.Password != ""
}
