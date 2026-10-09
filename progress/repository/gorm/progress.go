package gorm

import (
	"context"
	"errors"

	"github.com/jinzhu/gorm"
	"github.com/kosipov/students/models"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// GetGroup returns the group with subjects and their tasks. The stored documents are not loaded:
// the matrix and the public progress need only names and visibility.
func (r *Repository) GetGroup(ctx context.Context, id int, onlyVisible bool) (*models.Group, error) {
	var group models.Group
	query := r.db.
		// Subjects have no hidden flag; only groups and tasks do.
		Preload("Subjects", orderById).
		Preload("Subjects.SubjectObjects", func(db *gorm.DB) *gorm.DB {
			db = db.Select("id, name, comment, subject_id, hidden").Order("id")
			if onlyVisible {
				db = db.Where("hidden = ?", false)
			}
			return db
		})
	if onlyVisible {
		query = query.Where("hidden = ?", false)
	}
	result := query.First(&group, id)
	return &group, result.Error
}

func (r *Repository) GetStudent(ctx context.Context, id int) (*models.Student, error) {
	var student models.Student
	result := r.db.First(&student, id)
	return &student, result.Error
}

func (r *Repository) GetStudents(ctx context.Context, groupId int) ([]models.Student, error) {
	students := make([]models.Student, 0)
	result := r.db.Where("group_id = ?", groupId).Order("name").Find(&students)
	return students, result.Error
}

func (r *Repository) CreateStudent(ctx context.Context, student *models.Student) error {
	return r.db.Create(student).Error
}

func (r *Repository) RenameStudent(ctx context.Context, student *models.Student) error {
	return r.db.Model(&models.Student{ID: student.ID}).Update("name", student.Name).Error
}

func (r *Repository) DeleteStudent(ctx context.Context, student *models.Student) error {
	if student.ID == 0 {
		// gorm deletes every row when the primary key is blank.
		return errors.New("delete student: empty id")
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("student_id = ?", student.ID).Delete(&models.Mark{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", student.ID).Delete(&models.Student{}).Error
	})
}

func (r *Repository) GetMarks(ctx context.Context, studentIds []int) ([]models.Mark, error) {
	marks := make([]models.Mark, 0)
	if len(studentIds) == 0 {
		return marks, nil
	}
	result := r.db.Where("student_id IN (?)", studentIds).Find(&marks)
	return marks, result.Error
}

func (r *Repository) GetSubjectObject(ctx context.Context, id int) (*models.SubjectObject, error) {
	var subjectObject models.SubjectObject
	result := r.db.Select("id, subject_id").First(&subjectObject, id)
	if result.Error != nil {
		return nil, result.Error
	}
	if err := r.db.Model(&subjectObject).Related(&subjectObject.Subject).Error; err != nil {
		return nil, err
	}
	return &subjectObject, nil
}

// SetMark creates or updates the one mark a student has for a task.
func (r *Repository) SetMark(ctx context.Context, mark *models.Mark) error {
	var existing models.Mark
	err := r.db.Where("student_id = ? AND subject_object_id = ?", mark.StudentId, mark.SubjectObjectId).First(&existing).Error
	switch {
	case err == nil:
		return r.db.Model(&existing).Updates(map[string]interface{}{
			"status":  mark.Status,
			"comment": mark.Comment,
		}).Error
	case errors.Is(err, gorm.ErrRecordNotFound):
		return r.db.Create(mark).Error
	default:
		return err
	}
}

func (r *Repository) DeleteMark(ctx context.Context, studentId, subjectObjectId int) error {
	return r.db.Where("student_id = ? AND subject_object_id = ?", studentId, subjectObjectId).Delete(&models.Mark{}).Error
}

func orderById(db *gorm.DB) *gorm.DB {
	return db.Order("id")
}
