package repository

import (
	"my-flutter-backend/internal/model"

	"gorm.io/gorm"
)

type JadwalRepository interface {
	Create(jadwal *model.Jadwal) error
	GetByASNAndDate(asnID uint, date string) (*model.Jadwal, error)
	CreateMany(jadwal []model.Jadwal) error
}

type jadwalRepository struct {
	db *gorm.DB
}

func NewJadwalRepository(db *gorm.DB) JadwalRepository {
	return &jadwalRepository{db}
}

func (r *jadwalRepository) Create(jadwal *model.Jadwal) error {
	return r.db.Create(jadwal).Error
}

func (r *jadwalRepository) GetByASNAndDate(asnID uint, date string) (*model.Jadwal, error) {
	var jadwal model.Jadwal
	// Preload Shift penting untuk cek jam masuk nanti
	err := r.db.Preload("Shift").Where("asn_id = ? AND tanggal = ?", asnID, date).First(&jadwal).Error
	return &jadwal, err
}

func (r *jadwalRepository) CreateMany(jadwal []model.Jadwal) error {
	return r.db.Create(&jadwal).Error
}
