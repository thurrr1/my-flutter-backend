package repository

import (
	"my-flutter-backend/internal/model"

	"gorm.io/gorm"
)

type PerizinanRepository interface {
	CreateCuti(izin *model.PerizinanCuti) error
	GetByASNID(asnID uint) ([]model.PerizinanCuti, error)
	GetByAtasanID(atasanID uint) ([]model.PerizinanCuti, error)
	GetByID(id uint) (*model.PerizinanCuti, error)
	Update(izin *model.PerizinanCuti) error
}

type perizinanRepository struct {
	db *gorm.DB
}

func NewPerizinanRepository(db *gorm.DB) PerizinanRepository {
	return &perizinanRepository{db}
}

func (r *perizinanRepository) CreateCuti(izin *model.PerizinanCuti) error {
	return r.db.Create(izin).Error
}

func (r *perizinanRepository) GetByASNID(asnID uint) ([]model.PerizinanCuti, error) {
	var list []model.PerizinanCuti
	err := r.db.Where("asn_id = ?", asnID).Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *perizinanRepository) GetByAtasanID(atasanID uint) ([]model.PerizinanCuti, error) {
	var list []model.PerizinanCuti
	// Join ke tabel ASN untuk mencari bawahan dari atasanID ini
	err := r.db.Joins("JOIN asns ON asns.id = perizinan_cutis.asn_id").
		Where("asns.atasan_id = ?", atasanID).
		Preload("ASN"). // Load data nama pegawai
		Order("perizinan_cutis.created_at desc").
		Find(&list).Error
	return list, err
}

func (r *perizinanRepository) GetByID(id uint) (*model.PerizinanCuti, error) {
	var izin model.PerizinanCuti
	err := r.db.First(&izin, id).Error
	return &izin, err
}

func (r *perizinanRepository) Update(izin *model.PerizinanCuti) error {
	return r.db.Save(izin).Error
}
