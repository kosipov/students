package models

type Group struct {
	ID        uint16
	GroupName string
	// Hidden groups are shown only in the admin panel.
	Hidden   bool `gorm:"not null;default:false"`
	Subjects []Subject
}
