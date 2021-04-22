package database

import (
	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/jinzhu/gorm"
)

// Repository is user repository for CRUD operations.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates new repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db}
}

// FindByID find user by id
func (repo *Repository) FindByID(id uint64) (*FileModel, error) {
	var file FileModel
	file.ID = id

	if err := repo.db.
		First(&file).
		Error; err != nil {
		return nil, err
	}
	return &file, nil
}

// GetByID returns only if exists
func (repo *Repository) GetByID(id uint64) (*FileModel, error) {
	var file FileModel

	if err := repo.db.Where("id = ?", id).First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

// FindAdminVisibleByUID find admin visible files by user id
func (repo *Repository) FindAdminVisibleByUID(uid string, excludeCategories []string) ([]*FileModel, error) {
	var files []*FileModel
	if err := repo.db.Where("user_id = ? AND category NOT IN (?)", uid, excludeCategories).Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

// FindClientVisibleByUID find client visible files by user id
func (repo *Repository) FindClientVisibleByUID(uid string) ([]*FileModel, error) {
	var files []*FileModel
	if err := repo.db.Where("user_id = ? AND is_admin_only IS NOT TRUE", uid).Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

func (repo *Repository) GetList(params *list_params.ListParams) ([]*FileModel, error) {
	var items []*FileModel

	str, arguments := params.GetWhereCondition()
	query := repo.db.Where(str, arguments...)

	query = query.Order(params.GetOrderByString())

	if params.GetLimit() != 0 {
		query = query.Limit(params.GetLimit())
	}
	query = query.Offset(params.GetOffset())

	query = query.Joins(params.GetJoinCondition())

	for _, preloadName := range params.GetPreloads() {
		query = query.Preload(preloadName)
	}

	if err := query.Find(&items).Error; err != nil {
		return items, err
	}

	interfaceFiles := make([]interface{}, len(items))
	for i, itemPtr := range items {
		interfaceFiles[i] = itemPtr
	}
	for _, customIncludesFunc := range params.GetCustomIncludesFunctions() {
		if err := customIncludesFunc(interfaceFiles); err != nil {
			return items, err
		}
	}

	return items, nil
}

func (repo *Repository) GetTotalSizeOfUserFiles(uid string) (float64, error) {
	var files []*FileModel
	var result struct {
		Size float64
	}
	if err := repo.db.Where("user_id = ?", uid).Find(&files).Select("SUM(size) as size").Scan(&result).Error; err != nil {
		return result.Size, err
	}
	return result.Size, nil
}

// Create creates a new file
func (repo *Repository) Create(file *FileModel) (*FileModel, error) {
	if err := repo.db.Create(file).Error; err != nil {
		return nil, err
	}
	return file, nil
}

// Update updates an existing user
func (repo *Repository) Update(file *FileModel) (*FileModel, error) {
	if err := repo.db.Model(file).Updates(file).Error; err != nil {
		return nil, err
	}
	return file, nil
}

// Delete delete an existing user
func (repo *Repository) Delete(file *FileModel) error {
	if err := repo.db.Delete(file).Error; err != nil {
		return err
	}
	return nil
}
