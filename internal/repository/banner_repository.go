package repository

import (
	"my-flutter-backend/internal/model"

	"gorm.io/gorm"
)

type BannerRepository interface {
	GetAll() ([]model.Banner, error)
	Create(banner *model.Banner) error
	Delete(id uint) error
}

type bannerRepository struct {
	db *gorm.DB
}

func NewBannerRepository(db *gorm.DB) BannerRepository {
	return &bannerRepository{db}
}

func (r *bannerRepository) GetAll() ([]model.Banner, error) {
	var banners []model.Banner
	err := r.db.Where("is_active = ?", true).Order("created_at desc").Find(&banners).Error
	return banners, err
}

func (r *bannerRepository) Create(banner *model.Banner) error {
	return r.db.Create(banner).Error
}

func (r *bannerRepository) Delete(id uint) error {
	return r.db.Delete(&model.Banner{}, id).Error
}
