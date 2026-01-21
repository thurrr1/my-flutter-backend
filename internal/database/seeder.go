package database

import (
	"log"
	"my-flutter-backend/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedAll(db *gorm.DB) {
	// 1. Seed Organisasi
	org := model.Organisasi{NamaOrganisasi: "Dinas Komunikasi dan Informatika"}
	db.FirstOrCreate(&org, model.Organisasi{NamaOrganisasi: org.NamaOrganisasi})

	// 2. Seed Lokasi Kantor (Contoh Koordinat Kantor)
	lokasi := model.Lokasi{
		OrganisasiID: org.ID,
		NamaLokasi:   "Kantor Pusat Diskominfo",
		Latitude:     -0.9416, // Ganti dengan koordinat kantor aslimu
		Longitude:    100.3700,
		RadiusMeter:  50,
	}
	db.FirstOrCreate(&lokasi, model.Lokasi{NamaLokasi: lokasi.NamaLokasi})

	// 3. Seed Roles
	roles := []model.Role{
		{NamaRole: "Admin"},
		{NamaRole: "Atasan"},
		{NamaRole: "Pegawai"},
	}
	for _, r := range roles {
		db.FirstOrCreate(&r, model.Role{NamaRole: r.NamaRole})
	}

	// 4. Seed Permissions (Contoh)
	perms := []model.Permission{
		{NamaPermission: "edit_jadwal"},
		{NamaPermission: "approve_cuti"},
		{NamaPermission: "view_rekap"},
	}
	for _, p := range perms {
		db.FirstOrCreate(&p, model.Permission{NamaPermission: p.NamaPermission})
	}

	// 5. Seed Shift Default
	shiftNormal := model.Shift{
		NamaShift: "Normal (Senin-Jumat)",
		JamMasuk:  "07:30",
		JamPulang: "16:00",
	}
	db.FirstOrCreate(&shiftNormal, model.Shift{NamaShift: shiftNormal.NamaShift})

	// 6. Seed Akun Admin Pertama (ASN)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

	// Cari Role ID Admin
	var adminRole model.Role
	db.Where("nama_role = ?", "Admin").First(&adminRole)

	adminASN := model.ASN{
		Nama:         "Administrator Utama",
		NIP:          "123456789123456789",
		Password:     string(hashedPassword),
		Jabatan:      "Sekretaris",
		Bidang:       "Umum",
		RoleID:       adminRole.ID,
		OrganisasiID: org.ID,
		IsActive:     true,
	}

	result := db.FirstOrCreate(&adminASN, model.ASN{NIP: adminASN.NIP})
	if result.Error == nil {
		// Paksa update password agar selalu sinkron dengan "admin123" meskipun user sudah ada
		db.Model(&adminASN).Update("password", string(hashedPassword))
		log.Println("Seeding Admin berhasil!")
	}

	// 7. Seed Pegawai (Bawahan dari Admin)
	var pegawaiRole model.Role
	db.Where("nama_role = ?", "Pegawai").First(&pegawaiRole)

	pegawaiASN := model.ASN{
		Nama:         "Budi Pegawai",
		NIP:          "987654321", // NIP Beda
		Password:     string(hashedPassword),
		Jabatan:      "Staf Teknis",
		Bidang:       "Informatika",
		RoleID:       pegawaiRole.ID,
		OrganisasiID: org.ID,
		AtasanID:     &adminASN.ID, // PENTING: Link ke Admin sebagai atasan
		IsActive:     true,
	}
	db.FirstOrCreate(&pegawaiASN, model.ASN{NIP: pegawaiASN.NIP})
}
