package models

type SubjectObject struct {
	ID        int
	Name      string
	Comment   string
	Href      string
	SubjectId int
	Subject   Subject
}
