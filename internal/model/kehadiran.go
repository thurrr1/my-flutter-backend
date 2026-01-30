package model

import "gorm.io/gorm"

type Kehadiran struct {
	gorm.Model
	ASNID                uint  `json:"asn_id"`
	JadwalID             uint  `json:"jadwal_id"`
	LokasiID             *uint `json:"lokasi_id"` // Lokasi terdekat saat absen
	PerizinanCutiID      *uint `json:"perizinan_cuti_id"`
	PerizinanKehadiranID *uint `json:"perizinan_kehadiran_id"` // Izin Status Keterlambatan/Pulang Cepat
	PerizinanLokasiID    *uint `json:"perizinan_lokasi_id"`    // Izin Lokasi/Luar Radius/Dinas Luar

	JamMasukReal       string `json:"jam_masuk_real"`
	JamPulangReal      string `json:"jam_pulang_real"`
	StatusMasuk        string `json:"status_masuk"`        // HADIR/TERLAMBAT/CUTI/IZIN/ALPHA
	StatusPulang       string `json:"status_pulang"`       // HADIR/PULANG_CEPAT/CUTI/IZIN
	StatusLokasiMasuk  string `json:"status_lokasi_masuk"` // VALID/INVALID
	StatusLokasiPulang string `json:"status_lokasi_pulang"`
	KoordinatMasuk     string `json:"koordinat_masuk"`
	KoordinatPulang    string `json:"koordinat_pulang"`

	Tanggal string `json:"tanggal"`
	Hari    string `json:"hari"`
	Bulan   string `json:"bulan"`
	Tahun   string `json:"tahun"`
}
