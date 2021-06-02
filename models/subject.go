package models

type Subject struct {
	ID             int
	SubjectName    string
	GroupId        int
	Group          Group
	SubjectObjects []SubjectObject
}
