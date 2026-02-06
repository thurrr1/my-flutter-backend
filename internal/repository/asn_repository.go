package repository

import (
	"my-flutter-backend/internal/model"

	"gorm.io/gorm"
)

type ASNRepository interface {
	FindByNIP(nip string) (*model.ASN, error)
	GetLokasiByOrganisasiID(orgID uint) (*model.Lokasi, error)
	FindByID(id uint) (*model.ASN, error)
	Create(asn *model.ASN) error
	Update(asn *model.ASN) error
	Delete(id uint) error
	AddDevice(device *model.Device) error
	GetAll(search string) ([]model.ASN, error)
	ResetDevice(asnID uint) error
	Count() (int64, error)
	GetByPermission(permissionName string) ([]model.ASN, error)
	GetAllByOrganisasiID(orgID uint) ([]model.ASN, error)
	GetByAtasanID(atasanID uint) ([]model.ASN, error)
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
	err := r.db.Preload("Role.Permissions").Preload("Organisasi").Preload("Devices").Preload("Atasan").Where("nip = ?", nip).First(&asn).Error
	return &asn, err
}

func (r *asnRepository) GetLokasiByOrganisasiID(orgID uint) (*model.Lokasi, error) {
	var lokasi model.Lokasi
	err := r.db.Where("organisasi_id = ?", orgID).First(&lokasi).Error
	return &lokasi, err
}

func (r *asnRepository) FindByID(id uint) (*model.ASN, error) {
	var asn model.ASN
	err := r.db.Preload("Role").Preload("Atasan").Preload("Devices").First(&asn, id).Error
	return &asn, err
}

func (r *asnRepository) Create(asn *model.ASN) error {
	return r.db.Create(asn).Error
}

func (r *asnRepository) Update(asn *model.ASN) error {
	return r.db.Save(asn).Error
}

func (r *asnRepository) Delete(id uint) error {
	return r.db.Delete(&model.ASN{}, id).Error
}

func (r *asnRepository) AddDevice(device *model.Device) error {
	return r.db.Create(device).Error
}

func (r *asnRepository) GetAll(search string) ([]model.ASN, error) {
	var asns []model.ASN
	query := r.db.Preload("Role").Preload("Organisasi")

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("nama LIKE ? OR nip LIKE ?", searchPattern, searchPattern)
	}

	err := query.Find(&asns).Error
	return asns, err
}

func (r *asnRepository) ResetDevice(asnID uint) error {
	// Gunakan Unscoped() untuk Hard Delete (Hapus Permanen)
	// Ini PENTING agar UUID device tersebut benar-benar hilang dan bisa didaftarkan ulang
	return r.db.Unscoped().Where("asn_id = ?", asnID).Delete(&model.Device{}).Error
}

func (r *asnRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.ASN{}).Where("is_active = ?", true).Count(&count).Error
	return count, err
}

func (r *asnRepository) GetByPermission(permissionName string) ([]model.ASN, error) {
	var asns []model.ASN
	// Join tabel untuk mencari ASN yang memiliki Role dengan Permission tertentu
	err := r.db.Distinct().Table("asns").
		Joins("JOIN roles ON roles.id = asns.role_id").
		Joins("JOIN role_permissions ON role_permissions.role_id = roles.id").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("permissions.nama_permission = ?", permissionName).
		Preload("Role").Preload("Organisasi").Find(&asns).Error
	return asns, err
}

func (r *asnRepository) GetAllByOrganisasiID(orgID uint) ([]model.ASN, error) {
	var asns []model.ASN
	err := r.db.Where("organisasi_id = ?", orgID).Find(&asns).Error
	return asns, err
}

func (r *asnRepository) GetByAtasanID(atasanID uint) ([]model.ASN, error) {
	var asns []model.ASN
	err := r.db.Where("atasan_id = ?", atasanID).Preload("Role").Preload("Organisasi").Find(&asns).Error
	return asns, err
}
