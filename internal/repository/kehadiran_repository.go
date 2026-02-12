package repository

import (
	"my-flutter-backend/internal/model"
	"time"

	"gorm.io/gorm"
)

type KehadiranRepository interface {
	Create(kehadiran model.Kehadiran) error
	GetTodayAttendance(asnID uint) (*model.Kehadiran, error)
	Update(kehadiran *model.Kehadiran) error
	GetHistory(asnID uint) ([]model.Kehadiran, error)
	CreateMany(kehadiran []model.Kehadiran) error
	GetByDate(asnID uint, date string) (*model.Kehadiran, error)
	GetByMonth(asnID uint, month string, year string) ([]model.Kehadiran, error)
	CountByStatus(date string, status string) (int64, error)
	GetByDateAndOrg(date string, orgID uint) ([]model.Kehadiran, error)
	GetByMonthAndOrg(month string, year string, orgID uint) ([]model.Kehadiran, error)
	DeleteByPerizinanID(perizinanID uint) error
}

type kehadiranRepository struct {
	db *gorm.DB
}

func NewKehadiranRepository(db *gorm.DB) KehadiranRepository {
	return &kehadiranRepository{db}
}

func (r *kehadiranRepository) Create(kehadiran model.Kehadiran) error {
	return r.db.Create(&kehadiran).Error
}

func (r *kehadiranRepository) GetTodayAttendance(asnID uint) (*model.Kehadiran, error) {
	var kehadiran model.Kehadiran
	// Cek apakah hari ini sudah ada record (untuk validasi double check-in)
	today := time.Now().Format("2006-01-02")
	err := r.db.Where("asn_id = ? AND DATE(tanggal) = ?", asnID, today).First(&kehadiran).Error
	if err != nil {
		return nil, err
	}
	return &kehadiran, nil
}

func (r *kehadiranRepository) Update(kehadiran *model.Kehadiran) error {
	return r.db.Save(kehadiran).Error
}

func (r *kehadiranRepository) GetHistory(asnID uint) ([]model.Kehadiran, error) {
	var history []model.Kehadiran
	err := r.db.Where("asn_id = ?", asnID).Order("created_at desc").Find(&history).Error
	return history, err
}

func (r *kehadiranRepository) CreateMany(kehadiran []model.Kehadiran) error {
	return r.db.Create(&kehadiran).Error
}

func (r *kehadiranRepository) GetByDate(asnID uint, date string) (*model.Kehadiran, error) {
	var kehadiran model.Kehadiran
	// Gunakan Find + Limit(1) agar GORM tidak mencetak log error "record not found"
	err := r.db.Where("asn_id = ? AND DATE(tanggal) = ?", asnID, date).Limit(1).Find(&kehadiran).Error
	if err != nil {
		return nil, err
	}
	if kehadiran.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &kehadiran, nil
}

func (r *kehadiranRepository) GetByMonth(asnID uint, month string, year string) ([]model.Kehadiran, error) {
	var list []model.Kehadiran
	err := r.db.Where("asn_id = ? AND bulan = ? AND tahun = ?", asnID, month, year).Find(&list).Error
	return list, err
}

func (r *kehadiranRepository) CountByStatus(date string, status string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Kehadiran{}).Where("DATE(tanggal) = ? AND status_masuk = ?", date, status).Count(&count).Error
	return count, err
}

func (r *kehadiranRepository) GetByDateAndOrg(date string, orgID uint) ([]model.Kehadiran, error) {
	var list []model.Kehadiran
	err := r.db.Joins("JOIN asns ON asns.id = kehadirans.asn_id").
		Where("kehadirans.tanggal = ? AND asns.organisasi_id = ?", date, orgID).
		Find(&list).Error
	return list, err
}

func (r *kehadiranRepository) GetByMonthAndOrg(month string, year string, orgID uint) ([]model.Kehadiran, error) {
	var list []model.Kehadiran
	// [FIX] Tambahkan Select("kehadirans.*") untuk menghindari GORM mencoba load relasi Jadwal
	// yang menyebabkan error jika ada data Kehadiran dengan JadwalID = 0.
	err := r.db.Select("kehadirans.*").Joins("JOIN asns ON asns.id = kehadirans.asn_id").
		Where("kehadirans.bulan = ? AND kehadirans.tahun = ? AND asns.organisasi_id = ?", month, year, orgID).
		Find(&list).Error
	return list, err
}

func (r *kehadiranRepository) DeleteByPerizinanID(perizinanID uint) error {
	// Menghapus semua record kehadiran yang terkait dengan ID perizinan cuti tertentu
	return r.db.Where("perizinan_cuti_id = ?", perizinanID).Delete(&model.Kehadiran{}).Error
}
