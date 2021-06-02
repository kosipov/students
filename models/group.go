package models

type Group struct {
	ID        uint16
	GroupName string
	Subjects  []Subject
}
