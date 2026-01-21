package repository

import (
	"my-flutter-backend/internal/model"

	"gorm.io/gorm"
)

type ASNRepository interface {
	FindByNIP(nip string) (*model.ASN, error)
	GetLokasiByOrganisasiID(orgID uint) (*model.Lokasi, error)
	FindByID(id uint) (*model.ASN, error)
	Update(asn *model.ASN) error
	AddDevice(device *model.Device) error
	GetAll() ([]model.ASN, error)
	ResetDevice(asnID uint) error
	Count() (int64, error)
}

type asnRepository struct {
	db *gorm.DB
}

func NewASNRepository(db *gorm.DB) ASNRepository {
	return &asnRepository{db}
}

func (r *asnRepository) FindByNIP(nip string) (*model.ASN, error) {
	var asn model.ASN
	// Kita Preload Role dan Organisasi agar datanya lengkap saat login
	err := r.db.Preload("Role").Preload("Organisasi").Preload("Devices").Where("nip = ?", nip).First(&asn).Error
	return &asn, err
}

func (r *asnRepository) GetLokasiByOrganisasiID(orgID uint) (*model.Lokasi, error) {
	var lokasi model.Lokasi
	err := r.db.Where("organisasi_id = ?", orgID).First(&lokasi).Error
	return &lokasi, err
}

func (r *asnRepository) FindByID(id uint) (*model.ASN, error) {
	var asn model.ASN
	err := r.db.Preload("Atasan").Preload("Devices").First(&asn, id).Error
	return &asn, err
}

func (r *asnRepository) Update(asn *model.ASN) error {
	return r.db.Save(asn).Error
}

func (r *asnRepository) AddDevice(device *model.Device) error {
	return r.db.Create(device).Error
}

func (r *asnRepository) GetAll() ([]model.ASN, error) {
	var asns []model.ASN
	err := r.db.Preload("Role").Preload("Organisasi").Find(&asns).Error
	return asns, err
}

func (r *asnRepository) ResetDevice(asnID uint) error {
	return r.db.Where("asn_id = ?", asnID).Delete(&model.Device{}).Error
}

func (r *asnRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.ASN{}).Where("is_active = ?", true).Count(&count).Error
	return count, err
}
