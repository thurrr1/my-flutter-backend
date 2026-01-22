package repository

import (
	"my-flutter-backend/internal/model"

	"gorm.io/gorm"
)

type RoleRepository interface {
	GetAll() ([]model.Role, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db}
}

func (r *roleRepository) GetAll() ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Find(&roles).Error
	return roles, err
}
