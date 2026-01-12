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

func (r *UserRepository) GetByNIP(nip string) (model.User, error) {
	var user model.User
	err := r.db.Where("nip = ?", nip).First(&user).Error
	return user, err
}
