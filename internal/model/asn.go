package model

import "gorm.io/gorm"

type ASN struct {
	gorm.Model
	AtasanID     *uint      `json:"atasan_id"` // Self-reference
	OrganisasiID uint       `json:"organisasi_id"`
	RoleID       uint       `json:"role_id"`
	Nama         string     `json:"nama"`
	NIP          string     `json:"nip" gorm:"column:nip;unique;not null"`
	Password     string     `json:"-"`
	Email        string     `json:"email"`
	NoHP         string     `json:"no_hp"`
	Foto         string     `json:"foto"`
	Jabatan      string     `json:"jabatan"`
	Bidang       string     `json:"bidang"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	
	// Relasi
	Atasan    *ASN       `json:"atasan" gorm:"foreignKey:AtasanID"`
	Bawahan   []ASN      `json:"bawahan" gorm:"foreignKey:AtasanID"`
	Devices   []Device   `json:"devices"`
	Jadwal    []Jadwal   `json:"jadwal"`
	Kehadiran []Kehadiran `json:"kehadiran"`
	Role       Role       `gorm:"foreignKey:RoleID"`
    Organisasi Organisasi `gorm:"foreignKey:OrganisasiID"`
}