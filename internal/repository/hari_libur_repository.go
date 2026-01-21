package repository

import (
	"my-flutter-backend/internal/model"

	"gorm.io/gorm"
)

type HariLiburRepository interface {
	GetAll() ([]model.HariLibur, error)
	Create(libur *model.HariLibur) error
	Delete(id uint) error
	IsHoliday(date string) (bool, error)
}

type hariLiburRepository struct {
	db *gorm.DB
}

func NewHariLiburRepository(db *gorm.DB) HariLiburRepository {
	return &hariLiburRepository{db}
}

func (r *hariLiburRepository) GetAll() ([]model.HariLibur, error) {
	var liburs []model.HariLibur
	err := r.db.Order("tanggal desc").Find(&liburs).Error
	return liburs, err
}

func (r *hariLiburRepository) Create(libur *model.HariLibur) error {
	return r.db.Create(libur).Error
}

func (r *hariLiburRepository) Delete(id uint) error {
	return r.db.Delete(&model.HariLibur{}, id).Error
}

func (r *hariLiburRepository) IsHoliday(date string) (bool, error) {
	var count int64
	err := r.db.Model(&model.HariLibur{}).Where("tanggal = ?", date).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
