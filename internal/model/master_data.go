package model

import (
	"gorm.io/gorm"
)

type Jadwal struct {
	gorm.Model
	ASNID    uint   `json:"asn_id" gorm:"uniqueIndex:idx_asn_tanggal"`
	ShiftID  uint   `json:"shift_id"`
	Tanggal  string `json:"tanggal" gorm:"type:varchar(20);uniqueIndex:idx_asn_tanggal"`
	IsActive bool   `json:"is_active"`

	// Relasi
	Shift Shift `gorm:"foreignKey:ShiftID" json:"shift"`
	ASN   ASN   `gorm:"foreignKey:ASNID" json:"asn"`
}

type Shift struct {
	gorm.Model
	OrganisasiID uint   `json:"organisasi_id"`
	NamaShift    string `json:"nama_shift"`
	JamMasuk     string `json:"jam_masuk"`
	JamPulang    string `json:"jam_pulang"`
}

type Device struct {
	gorm.Model
	ASNID         uint   `json:"asn_id"`
	UUID          string `json:"uuid" gorm:"unique"`
	Brand         string `json:"brand"`
	Series        string `json:"series"`
	FirebaseToken string `json:"firebase_token"`
}

type Banner struct {
	gorm.Model
	OrganisasiID uint   `json:"organisasi_id"`
	Title        string `json:"title"`
	Foto         string `json:"foto"`
	IsActive     bool   `json:"is_active" gorm:"default:true"`
}

type HariLibur struct {
	gorm.Model
	OrganisasiID uint   `json:"organisasi_id"`
	Tanggal      string `json:"tanggal" gorm:"not null"` // Hapus unique global, biar bisa beda per org
	Keterangan   string `json:"keterangan"`
}

type Organisasi struct {
	gorm.Model
	NamaOrganisasi string   `json:"nama_organisasi"`
	EmailAdmin     string   `json:"email_admin"`
	Lokasis        []Lokasi `json:"lokasis" gorm:"foreignKey:OrganisasiID"` // Relasi One-to-Many
}

type Lokasi struct {
	gorm.Model
	OrganisasiID uint    `json:"organisasi_id"`
	NamaLokasi   string  `json:"nama_lokasi"`
	Alamat       string  `json:"alamat"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	RadiusMeter  float64 `json:"radius_meter"`
}

type Role struct {
	gorm.Model
	NamaRole    string       `json:"nama_role"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
}

type Permission struct {
	gorm.Model
	NamaPermission string `json:"nama_permission"` // Contoh: "view_dashboard", "create_asn"
}
