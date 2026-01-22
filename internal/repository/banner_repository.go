package repository

import (
	"my-flutter-backend/internal/model"

	"gorm.io/gorm"
)

type BannerRepository interface {
	GetAllActive() ([]model.Banner, error) // Untuk Mobile
	GetAll() ([]model.Banner, error)       // Untuk Admin
	Create(banner *model.Banner) error
	Delete(id uint) error
	ToggleStatus(id uint) error
}

type bannerRepository struct {
	db *gorm.DB
}

func NewBannerRepository(db *gorm.DB) BannerRepository {
	return &bannerRepository{db}
}

func (r *bannerRepository) GetAll() ([]model.Banner, error) {
	var banners []model.Banner
	err := r.db.Order("created_at desc").Find(&banners).Error
	return banners, err
}

func (r *bannerRepository) GetAllActive() ([]model.Banner, error) {
	var banners []model.Banner
	err := r.db.Where("is_active = ?", true).Order("created_at desc").Find(&banners).Error
	return banners, err
}

func (r *bannerRepository) Create(banner *model.Banner) error {
	return r.db.Create(banner).Error
}

func (r *bannerRepository) Delete(id uint) error {
	// Jangan hapus data (Delete), tapi update is_active jadi false
	return r.db.Model(&model.Banner{}).Where("id = ?", id).Update("is_active", false).Error
}

func (r *bannerRepository) ToggleStatus(id uint) error {
	var banner model.Banner
	if err := r.db.First(&banner, id).Error; err != nil {
		return err
	}
	return r.db.Model(&banner).Update("is_active", !banner.IsActive).Error
}
