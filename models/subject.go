package models

type Subject struct {
	Id          uint16
	SubjectName string
	GroupId     uint16
	Group       Group
}
