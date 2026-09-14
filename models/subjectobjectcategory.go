package models

// SubjectObjectCategory is a free-form label of a subject object, e.g. "Курсовые".
// A subject object may have several; students filter tasks of a subject by them.
type SubjectObjectCategory struct {
	ID              int
	SubjectObjectId int `gorm:"not null;unique_index:idx_subject_object_category"`
	// The binary collation compares names exactly. Which names count as the same category
	// ("Курсовые" and "курсовые") is decided by educational.CategoryKey, not by the database,
	// whose default collation would also merge names that differ only in accents.
	Name string `gorm:"type:varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;not null;unique_index:idx_subject_object_category"`
}

// CategoryNames returns the names of the subject object's categories in their order.
func (so SubjectObject) CategoryNames() []string {
	names := make([]string, 0, len(so.Categories))
	for _, category := range so.Categories {
		names = append(names, category.Name)
	}
	return names
}
