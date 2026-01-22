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
		StatusMasuk string
		Count       int64
	}
	monthStr := fmt.Sprintf("%02d", month)
	yearStr := strconv.Itoa(year)

	r.db.Table("kehadirans").
		Joins("JOIN asns ON asns.id = kehadirans.asn_id").
		Where("asns.organisasi_id = ? AND kehadirans.bulan = ? AND kehadirans.tahun = ?", orgID, monthStr, yearStr).
		Group("status_masuk").Select("status_masuk, count(*) as count").Scan(&monthly)

	monthlyMap := map[string]int64{"HADIR": 0, "TERLAMBAT": 0, "IZIN": 0, "CUTI": 0}
	for _, m := range monthly {
		monthlyMap[m.StatusMasuk] = m.Count
	}

	// Hitung Pulang Cepat Bulan Ini
	var pcMonthly int64
	r.db.Table("kehadirans").
		Joins("JOIN asns ON asns.id = kehadirans.asn_id").
		Where("asns.organisasi_id = ? AND kehadirans.bulan = ? AND kehadirans.tahun = ? AND status_pulang = ?", orgID, monthStr, yearStr, "PULANG_CEPAT").
		Count(&pcMonthly)
	monthlyMap["PULANG_CEPAT"] = pcMonthly

	stats["bulan_ini"] = monthlyMap

	return stats, nil
}
