package repository

import (
	"fmt"
	"my-flutter-backend/internal/model"
	"strconv"

	"gorm.io/gorm"
)

type DashboardRepository interface {
	GetDashboardStats(orgID uint, date string, month int, year int) (map[string]interface{}, error)
}

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{db}
}

// DetailStats dipindah ke level package agar lebih aman untuk reflection/serialization
type DetailStats struct {
	Tanggal              string `json:"tanggal"`
	StatusMasuk          string `json:"status_masuk"`
	StatusPulang         string `json:"status_pulang"`
	PerizinanKehadiranID *uint  `json:"perizinan_kehadiran_id"`
}

func (r *dashboardRepository) GetDashboardStats(orgID uint, date string, month int, year int) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 1. Total Pegawai Aktif
	var totalPegawai int64
	r.db.Model(&model.ASN{}).Where("organisasi_id = ? AND is_active = ?", orgID, true).Count(&totalPegawai)
	stats["total_pegawai"] = totalPegawai

	// 2. Statistik Harian (Hari Ini)
	var daily []struct {
		StatusMasuk string
		Count       int64
	}
	// Join dengan ASN untuk filter organisasi
	r.db.Table("kehadirans").
		Joins("JOIN asns ON asns.id = kehadirans.asn_id").
		Where("asns.organisasi_id = ? AND kehadirans.tanggal = ?", orgID, date).
		Group("status_masuk").Select("status_masuk, count(*) as count").Scan(&daily)

	dailyMap := map[string]int64{"HADIR": 0, "TERLAMBAT": 0, "IZIN": 0, "CUTI": 0}
	for _, d := range daily {
		dailyMap[d.StatusMasuk] = d.Count
	}

	// Hitung Pulang Cepat Hari Ini (Terpisah karena ada di kolom status_pulang)
	var pcDaily int64
	r.db.Table("kehadirans").
		Joins("JOIN asns ON asns.id = kehadirans.asn_id").
		Where("asns.organisasi_id = ? AND kehadirans.tanggal = ? AND status_pulang = ?", orgID, date, "PULANG_CEPAT").
		Count(&pcDaily)
	dailyMap["PULANG_CEPAT"] = pcDaily

	stats["hari_ini"] = dailyMap

	// 3. Statistik Bulanan (Bulan Ini)
	var monthly []struct {
		StatusMasuk          string
		PerizinanKehadiranID *uint
		Count                int64
	}
	monthStr := fmt.Sprintf("%02d", month)
	yearStr := strconv.Itoa(year)

	// Modified Query to include Perizinan checks
	r.db.Table("kehadirans").
		Joins("JOIN asns ON asns.id = kehadirans.asn_id").
		Where("asns.organisasi_id = ? AND kehadirans.bulan = ? AND kehadirans.tahun = ?", orgID, monthStr, yearStr).
		Select("status_masuk, perizinan_kehadiran_id, count(*) as count").
		Group("status_masuk, perizinan_kehadiran_id").Scan(&monthly)

	monthlyMap := map[string]int64{"HADIR": 0, "TERLAMBAT": 0, "IZIN": 0, "CUTI": 0, "TL_CP_DIIZINKAN": 0}

	for _, m := range monthly {
		if m.StatusMasuk == "TERLAMBAT" {
			if m.PerizinanKehadiranID != nil {
				monthlyMap["TL_CP_DIIZINKAN"] += m.Count
			} else {
				monthlyMap["TERLAMBAT"] += m.Count
			}
		} else {
			if count, ok := monthlyMap[m.StatusMasuk]; ok {
				monthlyMap[m.StatusMasuk] = count + m.Count
			} else {
				// Handle other statuses if necessary
				monthlyMap[m.StatusMasuk] += m.Count
			}
		}
	}

	// Hitung Pulang Cepat Bulan Ini
	var pcStats []struct {
		PerizinanKehadiranID *uint
		Count                int64
	}
	r.db.Table("kehadirans").
		Joins("JOIN asns ON asns.id = kehadirans.asn_id").
		Where("asns.organisasi_id = ? AND kehadirans.bulan = ? AND kehadirans.tahun = ? AND status_pulang = ?", orgID, monthStr, yearStr, "PULANG_CEPAT").
		Select("perizinan_kehadiran_id, count(*) as count").
		Group("perizinan_kehadiran_id").
		Scan(&pcStats)

	for _, pc := range pcStats {
		if pc.PerizinanKehadiranID != nil {
			monthlyMap["TL_CP_DIIZINKAN"] += pc.Count
		} else {
			// Asumsi Pulang Cepat masuk ke kategori TL/CP (Warning) jika tanpa izin
			monthlyMap["TERLAMBAT"] += pc.Count
		}
	}

	// Tambahan: Total Jadwal Bulan Ini (untuk persentase)
	var totalJadwal int64
	r.db.Table("jadwals").
		Joins("JOIN asns ON asns.id = jadwals.asn_id").
		Where("asns.organisasi_id = ? AND DATE_FORMAT(jadwals.tanggal, '%Y-%m') = ?", orgID, fmt.Sprintf("%d-%02d", year, month)).
		Where("jadwals.is_active = ?", true).
		Count(&totalJadwal)

	monthlyMap["total_jadwal"] = totalJadwal

	// Return map yang sesuai dng ekspektasi frontend
	stats["bulan_ini"] = map[string]interface{}{
		"hadir_tepat_waktu": monthlyMap["HADIR"],
		"tl_cp":             monthlyMap["TERLAMBAT"], // Tanpa Izin
		"tl_cp_diizinkan":   monthlyMap["TL_CP_DIIZINKAN"],
		"izin":              monthlyMap["IZIN"],
		"cuti":              monthlyMap["CUTI"],
		"total_jadwal":      totalJadwal,
		"alfa":              totalJadwal - (monthlyMap["HADIR"] + monthlyMap["TERLAMBAT"] + monthlyMap["TL_CP_DIIZINKAN"] + monthlyMap["IZIN"] + monthlyMap["CUTI"]), // Basic calc
		"belum_absen":       0,                                                                                                                                       // Placeholder needs proper logic if needed
	}

	// 4. Detail Harian Untuk Grafik
	var details []DetailStats

	r.db.Table("kehadirans").
		Joins("JOIN asns ON asns.id = kehadirans.asn_id").
		Where("asns.organisasi_id = ? AND kehadirans.bulan = ? AND kehadirans.tahun = ?", orgID, monthStr, yearStr).
		Select("kehadirans.tanggal, kehadirans.status_masuk, kehadirans.status_pulang, kehadirans.perizinan_kehadiran_id").
		Scan(&details)

	// Post-processing untuk menyesuaikan logic warna chart (Orange vs Kuning vs Hijau)
	for i := range details {
		// Logika: Jika Status Masuk terlambat ATAU Status Pulang Pulang Cepat
		if details[i].StatusMasuk == "TERLAMBAT" || details[i].StatusPulang == "PULANG_CEPAT" {
			if details[i].PerizinanKehadiranID != nil {
				// Jika ada izin -> Orange (TL_CP_DIIZINKAN)
				details[i].StatusMasuk = "TL_CP_DIIZINKAN"
			} else {
				// Jika TIDAK ada izin -> Kuning (TERLAMBAT / Tanpa Keterangan)
				details[i].StatusMasuk = "TERLAMBAT"
			}
		}
		// Sisa kondisi (HADIR, IZIN, CUTI) dibiarkan sesuai raw data
	}

	stats["bulan_ini"].(map[string]interface{})["detail"] = details

	return stats, nil
}
