package repository

import (
	"my-flutter-backend/internal/model"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user model.User) error {
	return r.db.Create(&user).Error
}

func (r *UserRepository) AddDevice(device model.Device) error {
	return r.db.Create(&device).Error
}

// Tambahan: Pastikan GetByNIP kamu sekarang menarik data Devices juga (Preload)
func (r *UserRepository) GetByNIP(nip string) (model.User, error) {
	var user model.User
	// Gunakan Preload agar data user.Devices tidak kosong saat dicek di Usecase
	err := r.db.Preload("Devices").Where("nip = ?", nip).First(&user).Error
	return user, err
}
