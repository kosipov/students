package models

type SubjectObject struct {
	Id        uint16
	Name      string
	SubjectId uint16
	Subject   Subject
}
