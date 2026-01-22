package repository

import (
	"my-flutter-backend/internal/model"

	"gorm.io/gorm"
)

type RoleRepository interface {
	GetAll() ([]model.Role, error)
	GetByID(id uint) (*model.Role, error)
	Create(role *model.Role, permissionIDs []uint) error
	Update(role *model.Role, permissionIDs []uint) error
	Delete(id uint) error
	GetAllPermissions() ([]model.Permission, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db}
}

func (r *roleRepository) GetAll() ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Preload("Permissions").Find(&roles).Error
	return roles, err
}

func (r *roleRepository) GetByID(id uint) (*model.Role, error) {
	var role model.Role
	err := r.db.Preload("Permissions").First(&role, id).Error
	return &role, err
}

func (r *roleRepository) Create(role *model.Role, permissionIDs []uint) error {
	// 1. Buat Role
	if err := r.db.Create(role).Error; err != nil {
		return err
	}
	// 2. Assign Permissions
	if len(permissionIDs) > 0 {
		var perms []model.Permission
		r.db.Where("id IN ?", permissionIDs).Find(&perms)
		return r.db.Model(role).Association("Permissions").Replace(perms)
	}
	return nil
}

func (r *roleRepository) Update(role *model.Role, permissionIDs []uint) error {
	// Update Nama Role
	if err := r.db.Save(role).Error; err != nil {
		return err
	}
	// Update Relasi Permissions (Replace existing)
	var perms []model.Permission
	r.db.Where("id IN ?", permissionIDs).Find(&perms)
	return r.db.Model(role).Association("Permissions").Replace(perms)
}

func (r *roleRepository) Delete(id uint) error {
	// Hapus Role (Relasi di role_permissions otomatis terhapus jika gorm di-setup cascade, atau manual)
	// GORM many2many biasanya aman dihapus record utamanya.
	return r.db.Select("Permissions").Delete(&model.Role{Model: gorm.Model{ID: id}}).Error
}

func (r *roleRepository) GetAllPermissions() ([]model.Permission, error) {
	var perms []model.Permission
	err := r.db.Find(&perms).Error
	return perms, err
}
