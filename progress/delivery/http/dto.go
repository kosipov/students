package http

import (
	"time"

	"github.com/kosipov/students/models"
	"github.com/kosipov/students/progress"
)

type groupJSON struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Hidden bool   `json:"hidden"`
}

type taskJSON struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Hidden bool   `json:"hidden"`
}

type subjectJSON struct {
	ID    int        `json:"id"`
	Name  string     `json:"name"`
	Tasks []taskJSON `json:"tasks"`
}

type markJSON struct {
	TaskId    int       `json:"taskId"`
	Status    string    `json:"status"`
	Comment   string    `json:"comment,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type studentJSON struct {
	ID    int        `json:"id"`
	Name  string     `json:"name"`
	Marks []markJSON `json:"marks"`
}

type matrixResponse struct {
	Group    groupJSON     `json:"group"`
	Subjects []subjectJSON `json:"subjects"`
	Students []studentJSON `json:"students"`
}

func toMatrixResponse(matrix *progress.Matrix) matrixResponse {
	response := matrixResponse{
		Group:    groupJSON{ID: int(matrix.Group.ID), Name: matrix.Group.GroupName, Hidden: matrix.Group.Hidden},
		Subjects: make([]subjectJSON, 0, len(matrix.Subjects)),
		Students: make([]studentJSON, 0, len(matrix.Students)),
	}
	for _, st := range matrix.Subjects {
		subject := subjectJSON{ID: st.Subject.ID, Name: st.Subject.SubjectName, Tasks: make([]taskJSON, 0, len(st.Tasks))}
		for i := range st.Tasks {
			subject.Tasks = append(subject.Tasks, taskJSON{ID: st.Tasks[i].ID, Name: st.Tasks[i].Name, Hidden: st.Tasks[i].Hidden})
		}
		response.Subjects = append(response.Subjects, subject)
	}
	for _, sm := range matrix.Students {
		response.Students = append(response.Students, toStudentJSON(sm.Student, sm.Marks))
	}
	return response
}

func toStudentJSON(student *models.Student, marks []models.Mark) studentJSON {
	result := studentJSON{ID: student.ID, Name: student.Name, Marks: make([]markJSON, 0, len(marks))}
	for i := range marks {
		result.Marks = append(result.Marks, markJSON{
			TaskId:    marks[i].SubjectObjectId,
			Status:    marks[i].Status,
			Comment:   marks[i].Comment,
			UpdatedAt: marks[i].UpdatedAt,
		})
	}
	return result
}

// Public

type rosterResponse struct {
	Group    groupJSON      `json:"group"`
	Students []studentBrief `json:"students"`
}

type studentBrief struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type taskProgressJSON struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Comment     string     `json:"comment"`
	Status      string     `json:"status"`
	MarkComment string     `json:"markComment,omitempty"`
	UpdatedAt   *time.Time `json:"updatedAt"`
}

type subjectProgressJSON struct {
	ID    int                `json:"id"`
	Name  string             `json:"name"`
	Tasks []taskProgressJSON `json:"tasks"`
}

type progressResponse struct {
	Student  studentBrief          `json:"student"`
	Group    groupJSON             `json:"group"`
	Subjects []subjectProgressJSON `json:"subjects"`
}

func toProgressResponse(studentProgress *progress.StudentProgress) progressResponse {
	response := progressResponse{
		Student:  studentBrief{ID: studentProgress.Student.ID, Name: studentProgress.Student.Name},
		Group:    groupJSON{ID: int(studentProgress.Group.ID), Name: studentProgress.Group.GroupName},
		Subjects: make([]subjectProgressJSON, 0, len(studentProgress.Subjects)),
	}
	for _, sp := range studentProgress.Subjects {
		subject := subjectProgressJSON{ID: sp.Subject.ID, Name: sp.Subject.SubjectName, Tasks: make([]taskProgressJSON, 0, len(sp.Tasks))}
		for _, tp := range sp.Tasks {
			task := taskProgressJSON{ID: tp.Task.ID, Name: tp.Task.Name, Comment: tp.Task.Comment, Status: ""}
			if tp.Mark != nil {
				task.Status = tp.Mark.Status
				task.MarkComment = tp.Mark.Comment
				updatedAt := tp.Mark.UpdatedAt
				task.UpdatedAt = &updatedAt
			}
			subject.Tasks = append(subject.Tasks, task)
		}
		response.Subjects = append(response.Subjects, subject)
	}
	return response
}

// Requests

type studentRequest struct {
	Name *string `json:"name"`
}

type importRequest struct {
	Names string `json:"names"`
}

type markRequest struct {
	Status  *string `json:"status"`
	Comment *string `json:"comment"`
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
