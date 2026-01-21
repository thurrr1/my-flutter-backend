package repository

import (
	"my-flutter-backend/internal/model"

	"gorm.io/gorm"
)

type OrganisasiRepository interface {
	GetFirst() (*model.Organisasi, error)
	Update(org *model.Organisasi) error
	GetLokasiByID(id uint) (*model.Lokasi, error)
	UpdateLokasi(lokasi *model.Lokasi) error
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
	err := r.db.Preload("Lokasi").First(&org).Error
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
