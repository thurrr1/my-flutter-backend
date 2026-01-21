package model

import (
	"gorm.io/gorm"
)

type Jadwal struct {
	gorm.Model
	ASNID   uint   `json:"asn_id"`
	ShiftID uint   `json:"shift_id"`
	Tanggal string `json:"tanggal"`

	// Relasi
	Shift Shift `gorm:"foreignKey:ShiftID" json:"shift"`
}

type Shift struct {
	gorm.Model
	NamaShift string `json:"nama_shift"`
	JamMasuk  string `json:"jam_masuk"`
	JamPulang string `json:"jam_pulang"`
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
	Title    string `json:"title"`
	Foto     string `json:"foto"`
	IsActive bool   `json:"is_active" gorm:"default:true"`
}

type HariLibur struct {
	gorm.Model
	Tanggal    string `json:"tanggal" gorm:"unique;not null"` // Format YYYY-MM-DD
	Keterangan string `json:"keterangan"`
}
