package repository

import (
	"my-flutter-backend/internal/model"

	"gorm.io/gorm"
)

type PerizinanKehadiranRepository interface {
	Create(koreksi *model.PerizinanKehadiran) error
	GetByASNID(asnID uint) ([]model.PerizinanKehadiran, error)
	GetByAtasanID(atasanID uint) ([]model.PerizinanKehadiran, error)
	GetByID(id uint) (*model.PerizinanKehadiran, error)
	Update(koreksi *model.PerizinanKehadiran) error
}

type perizinanKehadiranRepository struct {
	db *gorm.DB
}

func NewPerizinanKehadiranRepository(db *gorm.DB) PerizinanKehadiranRepository {
	return &perizinanKehadiranRepository{db}
}

func (r *perizinanKehadiranRepository) Create(koreksi *model.PerizinanKehadiran) error {
	return r.db.Create(koreksi).Error
}

func (r *perizinanKehadiranRepository) GetByASNID(asnID uint) ([]model.PerizinanKehadiran, error) {
	var list []model.PerizinanKehadiran
	err := r.db.Where("asn_id = ?", asnID).Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *perizinanKehadiranRepository) GetByAtasanID(atasanID uint) ([]model.PerizinanKehadiran, error) {
	var list []model.PerizinanKehadiran
	err := r.db.Joins("JOIN asns ON asns.id = perizinan_kehadirans.asn_id").
		Where("asns.atasan_id = ?", atasanID).
		Preload("ASN"). // Pastikan ada relasi ASN di model PerizinanKehadiran jika ingin nama muncul
		Order("perizinan_kehadirans.created_at desc").
		Find(&list).Error
	return list, err
}

func (r *perizinanKehadiranRepository) GetByID(id uint) (*model.PerizinanKehadiran, error) {
	var koreksi model.PerizinanKehadiran
	err := r.db.First(&koreksi, id).Error
	return &koreksi, err
}

func (r *perizinanKehadiranRepository) Update(koreksi *model.PerizinanKehadiran) error {
	return r.db.Save(koreksi).Error
}
