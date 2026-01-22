package repository

import (
	"my-flutter-backend/internal/model"

	"gorm.io/gorm"
)

type JadwalRepository interface {
	Create(jadwal *model.Jadwal) error
	GetByASNAndDate(asnID uint, date string) (*model.Jadwal, error)
	GetByDate(date string, orgID uint) ([]model.Jadwal, error)
	GetByID(id uint) (*model.Jadwal, error)
	Update(jadwal *model.Jadwal) error
	Delete(id uint) error
	CreateMany(jadwal []model.Jadwal) error
	CountByShiftID(shiftID uint) (int64, error)
	DeleteByDate(date string, orgID uint) error
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

func (r *jadwalRepository) GetByDate(date string, orgID uint) ([]model.Jadwal, error) {
	var jadwals []model.Jadwal
	// Join dengan ASN untuk filter organisasi dan Preload data yang dibutuhkan
	err := r.db.Preload("Shift").Preload("ASN").
		Joins("JOIN asns ON asns.id = jadwals.asn_id").
		Where("jadwals.tanggal = ? AND asns.organisasi_id = ?", date, orgID).
		Find(&jadwals).Error
	return jadwals, err
}

func (r *jadwalRepository) GetByID(id uint) (*model.Jadwal, error) {
	var jadwal model.Jadwal
	err := r.db.First(&jadwal, id).Error
	return &jadwal, err
}

func (r *jadwalRepository) Update(jadwal *model.Jadwal) error {
	return r.db.Save(jadwal).Error
}

func (r *jadwalRepository) Delete(id uint) error {
	return r.db.Delete(&model.Jadwal{}, id).Error
}

func (r *jadwalRepository) CreateMany(jadwal []model.Jadwal) error {
	return r.db.Create(&jadwal).Error
}

func (r *jadwalRepository) CountByShiftID(shiftID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Jadwal{}).Where("shift_id = ?", shiftID).Count(&count).Error
	return count, err
}

func (r *jadwalRepository) DeleteByDate(date string, orgID uint) error {
	// Hapus jadwal pada tanggal tertentu untuk semua pegawai di organisasi tersebut
	return r.db.Where("tanggal = ? AND asn_id IN (SELECT id FROM asns WHERE organisasi_id = ?)", date, orgID).Delete(&model.Jadwal{}).Error
}
