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
		{NamaRole: "Super Admin"}, // New Role
	}
	for _, r := range roles {
		db.FirstOrCreate(&r, model.Role{NamaRole: r.NamaRole})
	}

	// 4. Seed Permissions (Contoh)
	perms := []model.Permission{
		{NamaPermission: "kelola_organisasi"}, // New Permission
		{NamaPermission: "edit_jadwal"},
		{NamaPermission: "approve_cuti"},
		{NamaPermission: "view_rekap"},
	}
	for _, p := range perms {
		db.FirstOrCreate(&p, model.Permission{NamaPermission: p.NamaPermission})
	}

	// 4.1 Assign Permissions to Roles (Mapping Permission ke Role)
	var superAdminRole, adminRole, atasanRole, pegawaiRole model.Role
	db.Where("nama_role = ?", "Super Admin").First(&superAdminRole)
	db.Where("nama_role = ?", "Admin").First(&adminRole)     // ID 2
	db.Where("nama_role = ?", "Atasan").First(&atasanRole)   // ID 3
	db.Where("nama_role = ?", "Pegawai").First(&pegawaiRole) // ID 4

	var allPerms []model.Permission
	db.Find(&allPerms)

	// 0. Super Admin: Punya SEMUA permission
	db.Model(&superAdminRole).Association("Permissions").Replace(allPerms)

	// 1. Admin: Punya semua KECUALI 'kelola_organisasi'
	var adminPerms []model.Permission
	for _, p := range allPerms {
		if p.NamaPermission != "kelola_organisasi" {
			adminPerms = append(adminPerms, p)
		}
	}
	db.Model(&adminRole).Association("Permissions").Replace(adminPerms)

	// 2. Atasan: Semua kecuali 'edit_jadwal' & 'kelola_organisasi'
	var atasanPerms []model.Permission
	for _, p := range allPerms {
		if p.NamaPermission != "edit_jadwal" && p.NamaPermission != "kelola_organisasi" {
			atasanPerms = append(atasanPerms, p)
		}
	}
	db.Model(&atasanRole).Association("Permissions").Replace(atasanPerms)

	// 3. Pegawai (ID 3): Cuma punya permission 'view_rekap'
	var pegawaiPerms []model.Permission
	for _, p := range allPerms {
		if p.NamaPermission == "view_rekap" {
			pegawaiPerms = append(pegawaiPerms, p)
		}
	}
	db.Model(&pegawaiRole).Association("Permissions").Replace(pegawaiPerms)

	// 5. Seed Shift Default
	shiftNormal := model.Shift{
		NamaShift:    "Normal (Senin-Jumat)",
		JamMasuk:     "07:30",
		JamPulang:    "16:00",
		OrganisasiID: org.ID, // Assign ke Organisasi Default
	}
	db.FirstOrCreate(&shiftNormal, model.Shift{NamaShift: shiftNormal.NamaShift})

	// 6. Seed Akun Super Admin
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

	superAdminASN := model.ASN{
		Nama:         "Super Administrator",
		NIP:          "000000001", // NIP Khusus
		Password:     string(hashedPassword),
		Jabatan:      "IT Super Admin",
		Bidang:       "Pusat Data",
		RoleID:       superAdminRole.ID,
		OrganisasiID: org.ID, // Masih attach ke Org pertama, tapi punya permission lintas org
		IsActive:     true,
	}
	db.FirstOrCreate(&superAdminASN, model.ASN{NIP: superAdminASN.NIP})

	// Force Update Role & Password agar jika user sudah ada, tetap ter-update
	db.Model(&superAdminASN).Updates(map[string]interface{}{
		"role_id":   superAdminRole.ID,
		"password":  string(hashedPassword),
		"is_active": true,
	})

	// 7. Seed Akun Admin Pertama (ASN)
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

	// 8. Seed Pegawai (Bawahan dari Admin)
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
