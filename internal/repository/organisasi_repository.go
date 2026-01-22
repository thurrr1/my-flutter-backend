package repository

import (
	"my-flutter-backend/internal/model"

	"gorm.io/gorm"
)

type OrganisasiRepository interface {
	GetFirst() (*model.Organisasi, error)
	GetByID(id uint) (*model.Organisasi, error)
	Update(org *model.Organisasi) error
	GetLokasiByID(id uint) (*model.Lokasi, error)
	UpdateLokasi(lokasi *model.Lokasi) error
	CreateLokasi(lokasi *model.Lokasi) error
	DeleteLokasi(id uint) error
}

type organisasiRepository struct {
	db *gorm.DB
}

func NewOrganisasiRepository(db *gorm.DB) OrganisasiRepository {
	return &organisasiRepository{db}
}

func (r *organisasiRepository) GetFirst() (*model.Organisasi, error) {
	var org model.Organisasi
	// Asumsi aplikasi ini untuk 1 instansi, ambil yang pertama
	err := r.db.Preload("Lokasis").First(&org).Error
	return &org, err
}

func (r *organisasiRepository) GetByID(id uint) (*model.Organisasi, error) {
	var org model.Organisasi
	err := r.db.Preload("Lokasis").First(&org, id).Error
	return &org, err
}

func (r *organisasiRepository) Update(org *model.Organisasi) error {
	return r.db.Save(org).Error
}

func (r *organisasiRepository) GetLokasiByID(id uint) (*model.Lokasi, error) {
	var lokasi model.Lokasi
	err := r.db.First(&lokasi, id).Error
	return &lokasi, err
}

func (r *organisasiRepository) UpdateLokasi(lokasi *model.Lokasi) error {
	return r.db.Save(lokasi).Error
}

func (r *organisasiRepository) CreateLokasi(lokasi *model.Lokasi) error {
	return r.db.Create(lokasi).Error
}

func (r *organisasiRepository) DeleteLokasi(id uint) error {
	return r.db.Delete(&model.Lokasi{}, id).Error
}
