package model

import "gorm.io/gorm"

type PerizinanCuti struct {
	gorm.Model
	ASNID          uint   `json:"asn_id"`
	NIPAtasan      string `json:"nip_atasan"`
	Tipe           string `json:"tipe"`  // IZIN atau CUTI
	Jenis          string `json:"jenis"` // Sakit, Tahunan, dll
	TanggalMulai   string `json:"tanggal_mulai"`
	TanggalSelesai string `json:"tanggal_selesai"`
	Alasan         string `json:"alasan"`
	Status         string `json:"status" gorm:"default:PENDING"`
	PathFile       string `json:"path_file"`

	// Relasi untuk Preload data pemohon
	ASN ASN `gorm:"foreignKey:ASNID" json:"asn"`
}

type PerizinanKehadiran struct {
	gorm.Model
	ASNID            uint   `json:"asn_id"`
	NIPAtasan        string `json:"nip_atasan"`
	TanggalKehadiran string `json:"tanggal_kehadiran"`
	TipeKoreksi      string `json:"tipe_koreksi"` // TELAT, PULANG_CEPAT, LUAR_RADIUS
	Alasan           string `json:"alasan"`
	Status           string `json:"status" gorm:"default:PENDING"`
	PathFile         string `json:"path_file"`

	// Relasi agar Preload("ASN") di repository koreksi bekerja
	ASN ASN `gorm:"foreignKey:ASNID" json:"asn"`
}
