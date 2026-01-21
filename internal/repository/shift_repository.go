package repository

import (
	"my-flutter-backend/internal/model"

	"gorm.io/gorm"
)

type ShiftRepository interface {
	GetAll() ([]model.Shift, error)
	Create(shift *model.Shift) error
	Update(shift *model.Shift) error
	Delete(id uint) error
	GetByID(id uint) (*model.Shift, error)
}

type shiftRepository struct {
	db *gorm.DB
}

func NewShiftRepository(db *gorm.DB) ShiftRepository {
	return &shiftRepository{db}
}

func (r *shiftRepository) GetAll() ([]model.Shift, error) {
	var shifts []model.Shift
	err := r.db.Find(&shifts).Error
	return shifts, err
}

func (r *shiftRepository) Create(shift *model.Shift) error {
	return r.db.Create(shift).Error
}

func (r *shiftRepository) Update(shift *model.Shift) error {
	return r.db.Save(shift).Error
}

func (r *shiftRepository) Delete(id uint) error {
	return r.db.Delete(&model.Shift{}, id).Error
}

func (r *shiftRepository) GetByID(id uint) (*model.Shift, error) {
	var shift model.Shift
	err := r.db.First(&shift, id).Error
	return &shift, err
}
