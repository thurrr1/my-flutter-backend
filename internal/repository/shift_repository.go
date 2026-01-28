package repository

import (
	"fmt"
	"my-flutter-backend/internal/model"

	"gorm.io/gorm"
)

type ShiftRepository interface {
	GetAll() ([]model.Shift, error)
	Create(shift *model.Shift) error
	Update(shift *model.Shift) error
	Delete(id uint) error
	GetByID(id uint) (*model.Shift, error)
	FindOrCreate(jamMasuk, jamPulang string) (*model.Shift, error)
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

func (r *shiftRepository) FindOrCreate(jamMasuk, jamPulang string) (*model.Shift, error) {
	var shift model.Shift
	// Cek apakah shift dengan jam tersebut sudah ada
	err := r.db.Where("jam_masuk = ? AND jam_pulang = ?", jamMasuk, jamPulang).First(&shift).Error
	if err == nil {
		return &shift, nil
	}

	if err == gorm.ErrRecordNotFound {
		// Buat shift baru jika belum ada
		newShift := model.Shift{
			NamaShift: fmt.Sprintf("%s-%s", jamMasuk, jamPulang),
			JamMasuk:  jamMasuk,
			JamPulang: jamPulang,
		}
		if err := r.db.Create(&newShift).Error; err != nil {
			return nil, err
		}
		return &newShift, nil
	}

	return nil, err
}
