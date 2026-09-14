package gorm

import (
	"context"
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/kosipov/students/models"
)

type SubjectRepository struct {
	db *gorm.DB
}

func NewSubjectRepository(db *gorm.DB) *SubjectRepository {
	return &SubjectRepository{db: db}
}

func (subjectRepo *SubjectRepository) GetGroup(ctx context.Context, groupId int) (*models.Group, error) {
	var group models.Group
	result := subjectRepo.db.First(&group, groupId)
	return &group, result.Error
}

func (subjectRepo *SubjectRepository) GetSubject(ctx context.Context, id int) (*models.Subject, error) {
	var subject models.Subject
	result := subjectRepo.db.First(&subject, id)
	return &subject, result.Error
}

func (subjectRepo *SubjectRepository) GetSubjectWithSubjectObjects(ctx context.Context, id int) (*models.Subject, error) {
	var subject models.Subject
	result := subjectRepo.db.
		Preload("Group").
		Preload("SubjectObjects", func(db *gorm.DB) *gorm.DB {
			return db.Select(subjectObjectListColumns).Order("id")
		}).
		Preload("SubjectObjects.Categories", orderById).
		First(&subject, id)
	return &subject, result.Error
}

func (subjectRepo *SubjectRepository) CreateSubject(ctx context.Context, subject *models.Subject) error {
	return subjectRepo.db.Create(subject).Error
}

func (subjectRepo *SubjectRepository) UpdateSubject(ctx context.Context, subject *models.Subject) error {
	return subjectRepo.db.Model(&models.Subject{ID: subject.ID}).Update("subject_name", subject.SubjectName).Error
}

func (subjectRepo *SubjectRepository) DeleteSubject(ctx context.Context, subject *models.Subject) error {
	if subject.ID == 0 {
		// gorm deletes every row when the primary key is blank.
		return errors.New("delete subject: empty id")
	}

	return subjectRepo.db.Transaction(func(tx *gorm.DB) error {
		if err := deleteSubjectObjectsOfSubjects(tx, []int{subject.ID}); err != nil {
			return err
		}
		return tx.Where("id = ?", subject.ID).Delete(&models.Subject{}).Error
	})
}

func (subjectRepo *SubjectRepository) GetSubjectObject(ctx context.Context, subjectObjectId int) (*models.SubjectObject, error) {
	var subjectObject models.SubjectObject
	result := subjectRepo.db.Preload("Categories", orderById).First(&subjectObject, subjectObjectId)
	return &subjectObject, result.Error
}

func (subjectRepo *SubjectRepository) CreateSubjectObject(ctx context.Context, subjectObject *models.SubjectObject) error {
	return subjectRepo.db.Transaction(func(tx *gorm.DB) error {
		// Categories are created explicitly below, not by gorm's association saving.
		if err := tx.Set("gorm:save_associations", false).Create(subjectObject).Error; err != nil {
			return err
		}
		return createCategories(tx, subjectObject)
	})
}

// Updates below target a bare model with the id only: gorm would otherwise save
// associations loaded into the passed struct back to the database.

func (subjectRepo *SubjectRepository) UpdateSubjectObject(ctx context.Context, subjectObject *models.SubjectObject) error {
	return subjectRepo.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&models.SubjectObject{ID: subjectObject.ID}).Updates(map[string]interface{}{
			"name":    subjectObject.Name,
			"href":    subjectObject.Href,
			"comment": subjectObject.Comment,
			"hidden":  subjectObject.Hidden,
		}).Error
		if err != nil {
			return err
		}
		if err := tx.Where("subject_object_id = ?", subjectObject.ID).Delete(&models.SubjectObjectCategory{}).Error; err != nil {
			return err
		}
		return createCategories(tx, subjectObject)
	})
}

// UpdateSubjectObjectContent stores the document only while the subject object still has the same link,
// so a download that finishes after the link was changed doesn't overwrite the new state.
func (subjectRepo *SubjectRepository) UpdateSubjectObjectContent(ctx context.Context, subjectObject *models.SubjectObject) error {
	return subjectRepo.db.Model(&models.SubjectObject{ID: subjectObject.ID}).
		Where("href = ?", subjectObject.Href).
		Updates(map[string]interface{}{
			"content":             subjectObject.Content,
			"content_etag":        subjectObject.ContentETag,
			"content_fetched_at":  subjectObject.ContentFetchedAt,
			"content_checked_at":  subjectObject.ContentCheckedAt,
			"content_error":       subjectObject.ContentError,
			"content_unsupported": subjectObject.ContentUnsupported,
		}).Error
}

func (subjectRepo *SubjectRepository) DeleteSubjectObject(ctx context.Context, subjectObject *models.SubjectObject) error {
	if subjectObject.ID == 0 {
		// gorm deletes every row when the primary key is blank.
		return errors.New("delete subject object: empty id")
	}
	return subjectRepo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("subject_object_id = ?", subjectObject.ID).Delete(&models.SubjectObjectCategory{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", subjectObject.ID).Delete(&models.SubjectObject{}).Error
	})
}

func (subjectRepo *SubjectRepository) GetCategoryNames(ctx context.Context) ([]string, error) {
	var names []string
	err := subjectRepo.db.Raw("SELECT DISTINCT name FROM subject_object_categories ORDER BY name").Pluck("name", &names).Error
	return names, err
}

func createCategories(tx *gorm.DB, subjectObject *models.SubjectObject) error {
	for i := range subjectObject.Categories {
		category := &subjectObject.Categories[i]
		category.ID = 0
		category.SubjectObjectId = subjectObject.ID
		if err := tx.Create(category).Error; err != nil {
			return err
		}
	}
	return nil
}

// deleteSubjectObjectsOfSubjects deletes subject objects of the subjects with their categories.
func deleteSubjectObjectsOfSubjects(tx *gorm.DB, subjectIds []int) error {
	if len(subjectIds) == 0 {
		return nil
	}
	var subjectObjectIds []int
	if err := tx.Model(&models.SubjectObject{}).Where("subject_id IN (?)", subjectIds).Pluck("id", &subjectObjectIds).Error; err != nil {
		return err
	}
	if len(subjectObjectIds) > 0 {
		if err := tx.Where("subject_object_id IN (?)", subjectObjectIds).Delete(&models.SubjectObjectCategory{}).Error; err != nil {
			return err
		}
	}
	return tx.Where("subject_id IN (?)", subjectIds).Delete(&models.SubjectObject{}).Error
}

func orderById(db *gorm.DB) *gorm.DB {
	return db.Order("id")
}
