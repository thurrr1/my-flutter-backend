package handler

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"my-flutter-backend/internal/model"
	"my-flutter-backend/internal/repository"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
)

type ASNHandler struct {
	repo repository.ASNRepository
}

func NewASNHandler(repo repository.ASNRepository) *ASNHandler {
	return &ASNHandler{repo: repo}
}

type LoginRequest struct {
	NIP           string `json:"nip"`
	Password      string `json:"password"`
	DeviceID      string `json:"device_id"`      // UUID Unik Perangkat
	Brand         string `json:"brand"`          // Merk HP (e.g. Samsung)
	Series        string `json:"series"`         // Tipe HP (e.g. Galaxy S23)
	FirebaseToken string `json:"firebase_token"` // Untuk Notifikasi nanti
}

func (h *ASNHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format data salah"})
	}

	// 1. Cari ASN by NIP
	asn, err := h.repo.FindByNIP(req.NIP)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "NIP atau Password salah"})
	}

	// 2. Cek Password
	err = bcrypt.CompareHashAndPassword([]byte(asn.Password), []byte(req.Password))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "NIP atau Password salah"})
	}

	// 3. Cek Device Binding (Logika Keamanan)
	if req.DeviceID != "" {
		// Cek apakah akun ini sudah punya device terdaftar
		if len(asn.Devices) > 0 {
			isRegistered := false
			for _, d := range asn.Devices {
				if d.UUID == req.DeviceID {
					isRegistered = true
					break
				}
			}
			if !isRegistered {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Akun ini terkunci pada perangkat lain. Hubungi admin untuk reset.",
				})
			}
		} else {
			// Jika belum punya device, daftarkan device ini (Binding Pertama Kali)
			newDevice := model.Device{
				ASNID:         asn.ID,
				UUID:          req.DeviceID,
				Brand:         req.Brand,
				Series:        req.Series,
				FirebaseToken: req.FirebaseToken,
			}
			if err := h.repo.AddDevice(&newDevice); err != nil {
				// Kemungkinan error: UUID sudah dipakai user lain (karena unique)
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Perangkat ini sudah digunakan oleh akun lain.",
				})
			}
		}
	}

	// 4. Generate Token JWT
	accessToken, refreshToken, err := generateTokens(asn)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuat token"})
	}

	// Ambil list permission
	var permissions []string
	for _, p := range asn.Role.Permissions {
		permissions = append(permissions, p.NamaPermission)
	}

	nipAtasan := ""
	if asn.Atasan != nil {
		nipAtasan = asn.Atasan.NIP
	}

	// 4. Return Token ke Client
	return c.JSON(fiber.Map{
		"message":       "Login berhasil",
		"token":         accessToken,  // Access Token (15 Menit)
		"refresh_token": refreshToken, // Refresh Token (7 Hari)
		"data": fiber.Map{
			"nip":         asn.NIP,
			"nama":        asn.Nama,
			"role":        asn.Role.NamaRole,
			"permissions": permissions, // Kirim permission ke frontend
			"jabatan":     asn.Jabatan,
			"bidang":      asn.Bidang,
			"organisasi":  asn.Organisasi.NamaOrganisasi, // Tambahan untuk Dashboard
			"atasan_id":   asn.AtasanID,
			"nip_atasan":  nipAtasan,
			"foto":        asn.Foto,
		},
	})
}

// WebLogin: Login khusus via Web Admin (Tanpa Cek Device Binding)
func (h *ASNHandler) WebLogin(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format data salah"})
	}

	// 1. Cari ASN by NIP
	asn, err := h.repo.FindByNIP(req.NIP)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "NIP atau Password salah"})
	}

	// 2. Cek Password
	err = bcrypt.CompareHashAndPassword([]byte(asn.Password), []byte(req.Password))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "NIP atau Password salah"})
	}

	// 3. SKIP Cek Device Binding (Khusus Web)

	// 4. Generate Token JWT
	accessToken, refreshToken, err := generateTokens(asn)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuat token"})
	}

	// Ambil list permission
	var permissions []string
	for _, p := range asn.Role.Permissions {
		permissions = append(permissions, p.NamaPermission)
	}

	nipAtasan := ""
	if asn.Atasan != nil {
		nipAtasan = asn.Atasan.NIP
	}

	// 5. Return Token ke Client
	return c.JSON(fiber.Map{
		"message":       "Login berhasil",
		"token":         accessToken,  // Access Token (15 Menit)
		"refresh_token": refreshToken, // Refresh Token (7 Hari)
		"data": fiber.Map{
			"nip":         asn.NIP,
			"nama":        asn.Nama,
			"role":        asn.Role.NamaRole,
			"permissions": permissions,
			"jabatan":     asn.Jabatan,
			"bidang":      asn.Bidang,
			"organisasi":  asn.Organisasi.NamaOrganisasi,
			"atasan_id":   asn.AtasanID,
			"nip_atasan":  nipAtasan,
			"foto":        asn.Foto,
		},
	})
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *ASNHandler) RefreshToken(c *fiber.Ctx) error {
	var req RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format data salah"})
	}

	// Parse & Validasi Refresh Token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte("rahasia_negara"), nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Refresh token tidak valid atau kadaluwarsa"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token claims invalid"})
	}

	// Ambil User ID dari claims refresh token
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token invalid (user_id)"})
	}
	userID := uint(userIDFloat)
	asn, err := h.repo.FindByID(userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User tidak ditemukan"})
	}

	// Generate Token Baru
	newAccessToken, newRefreshToken, err := generateTokens(asn)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal generate token"})
	}

	return c.JSON(fiber.Map{
		"token":         newAccessToken,
		"refresh_token": newRefreshToken,
	})
}

func (h *ASNHandler) GetProfile(c *fiber.Ctx) error {
	asnID := uint(c.Locals("user_id").(float64))

	asn, err := h.repo.FindByID(asnID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User tidak ditemukan"})
	}

	return c.JSON(fiber.Map{
		"message": "Berhasil mengambil profil",
		"data":    asn,
	})
}

type UpdateProfileRequest struct {
	Email string `json:"email"`
	NoHP  string `json:"no_hp"`
	Foto  string `json:"foto"` // Base64 atau URL
}

func (h *ASNHandler) UpdateProfile(c *fiber.Ctx) error {
	asnID := uint(c.Locals("user_id").(float64))

	var req UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	asn, err := h.repo.FindByID(asnID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User tidak ditemukan"})
	}

	// Update field yang diizinkan
	if req.Email != "" {
		asn.Email = req.Email
	}
	if req.NoHP != "" {
		asn.NoHP = req.NoHP
	}
	if req.Foto != "" {
		// Cek apakah ini Base64 Image
		if strings.HasPrefix(req.Foto, "data:image") {
			// Format: "data:image/jpeg;base64,....."
			parts := strings.Split(req.Foto, ",")
			if len(parts) == 2 {
				// Decode Base64
				imgData, err := base64.StdEncoding.DecodeString(parts[1])
				if err == nil {
					// Simpan ke File
					uploadDir := "./uploads/profile"
					os.MkdirAll(uploadDir, 0755)

					// Tentukan ekstensi dari header
					ext := ".jpg" // Default
					if strings.Contains(parts[0], "png") {
						ext = ".png"
					}

					filename := fmt.Sprintf("%d_%d%s", asnID, time.Now().Unix(), ext)
					pathFile := fmt.Sprintf("uploads/profile/%s", filename)

					if err := os.WriteFile(pathFile, imgData, 0644); err == nil {
						asn.Foto = pathFile // Simpan path ke DB
					}
				}
			}
		} else {
			// Jika bukan base64, asumsikan ini URL atau Path yang sudah valid (atau update string biasa)
			asn.Foto = req.Foto
		}
	}

	if err := h.repo.Update(asn); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal update profil"})
	}

	return c.JSON(fiber.Map{
		"message": "Profil berhasil diperbarui",
		"data":    asn,
	})
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (h *ASNHandler) ChangePassword(c *fiber.Ctx) error {
	asnID := uint(c.Locals("user_id").(float64))

	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	asn, err := h.repo.FindByID(asnID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User tidak ditemukan"})
	}

	// Cek Password Lama
	if err := bcrypt.CompareHashAndPassword([]byte(asn.Password), []byte(req.OldPassword)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Password lama salah"})
	}

	// Hash Password Baru
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengenkripsi password"})
	}

	asn.Password = string(hashedPassword)
	if err := h.repo.Update(asn); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal update password"})
	}

	return c.JSON(fiber.Map{"message": "Password berhasil diubah"})
}

func (h *ASNHandler) GetAll(c *fiber.Ctx) error {
	// Ambil Organisasi ID dari user yang login (Admin)
	orgID := uint(c.Locals("organisasi_id").(float64))
	search := c.Query("search")

	asns, err := h.repo.GetAll(search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data pegawai"})
	}

	// Filter: Hanya tampilkan ASN yang satu organisasi dengan Admin
	var filteredASNs []model.ASN
	for _, a := range asns {
		if a.OrganisasiID == orgID {
			filteredASNs = append(filteredASNs, a)
		}
	}

	return c.JSON(fiber.Map{"data": filteredASNs})
}

func (h *ASNHandler) GetASNDetail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	asn, err := h.repo.FindByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Pegawai tidak ditemukan"})
	}
	return c.JSON(fiber.Map{"data": asn})
}

// --- FITUR BARU: CRUD ADMIN ---

type CreateASNRequest struct {
	Nama     string `json:"nama"`
	NIP      string `json:"nip"`
	Password string `json:"password"`
	Jabatan  string `json:"jabatan"`
	Bidang   string `json:"bidang"`
	RoleID   uint   `json:"role_id"`
	Email    string `json:"email"`
	NoHP     string `json:"no_hp"`
}

func (h *ASNHandler) CreateASN(c *fiber.Ctx) error {
	orgID := uint(c.Locals("organisasi_id").(float64))
	var req CreateASNRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	asn := model.ASN{
		Nama:         req.Nama,
		NIP:          req.NIP,
		Password:     string(hashedPassword),
		Jabatan:      req.Jabatan,
		Bidang:       req.Bidang,
		RoleID:       req.RoleID,
		OrganisasiID: orgID,
		IsActive:     true,
		Email:        req.Email,
		NoHP:         req.NoHP,
	}

	if err := h.repo.Create(&asn); err != nil {
		// Tampilkan error asli agar ketahuan penyebabnya (misal: Duplicate NIP)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Pegawai berhasil ditambahkan", "data": asn})
}

func (h *ASNHandler) ImportASN(c *fiber.Ctx) error {
	orgID := uint(c.Locals("organisasi_id").(float64))
	var reqs []CreateASNRequest
	if err := c.BodyParser(&reqs); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format data tidak valid"})
	}

	successCount := 0
	for _, req := range reqs {
		// Skip jika NIP kosong
		if req.NIP == "" {
			continue
		}

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

		asn := model.ASN{
			Nama:         req.Nama,
			NIP:          req.NIP,
			Password:     string(hashedPassword),
			Jabatan:      req.Jabatan,
			Bidang:       req.Bidang,
			RoleID:       req.RoleID,
			OrganisasiID: orgID,
			IsActive:     true,
			Email:        req.Email,
			NoHP:         req.NoHP,
		}

		// Abaikan error duplicate entry agar proses tetap lanjut untuk data lain
		if err := h.repo.Create(&asn); err == nil {
			successCount++
		}
	}

	return c.JSON(fiber.Map{"message": fmt.Sprintf("Berhasil mengimport %d data pegawai", successCount)})
}

func (h *ASNHandler) UpdateASN(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var req model.ASN
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	asn, err := h.repo.FindByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Pegawai tidak ditemukan"})
	}

	asn.Nama = req.Nama
	asn.Jabatan = req.Jabatan
	asn.Bidang = req.Bidang
	asn.IsActive = req.IsActive // Update Status Aktif/Nonaktif
	asn.RoleID = req.RoleID     // Update Role
	asn.Email = req.Email
	asn.NoHP = req.NoHP
	// NIP dan Password biasanya butuh endpoint khusus atau validasi ketat

	if err := h.repo.Update(asn); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal update pegawai"})
	}
	return c.JSON(fiber.Map{"message": "Data pegawai berhasil diupdate"})
}

func (h *ASNHandler) DeleteASN(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.repo.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menghapus pegawai"})
	}
	return c.JSON(fiber.Map{"message": "Pegawai berhasil dihapus"})
}

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password"`
}

func (h *ASNHandler) ResetUserPassword(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	if len(req.NewPassword) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Password minimal 6 karakter"})
	}

	asn, err := h.repo.FindByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Pegawai tidak ditemukan"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengenkripsi password"})
	}

	asn.Password = string(hashedPassword)
	if err := h.repo.Update(asn); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal reset password"})
	}

	return c.JSON(fiber.Map{"message": "Password berhasil di-reset"})
}

func (h *ASNHandler) ResetDevice(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.repo.ResetDevice(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal reset device"})
	}
	return c.JSON(fiber.Map{"message": "Device berhasil di-reset, user bisa login di HP baru"})
}

func (h *ASNHandler) GetListAtasan(c *fiber.Ctx) error {
	// Cari ASN yang punya permission 'approve_cuti'
	// Pastikan nama permission di database sesuai, misal: "approve_cuti" atau "approval_atasan"
	asns, err := h.repo.GetByPermission("approve_cuti")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data atasan"})
	}

	// Filter: Exclude diri sendiri dari list
	myID := uint(c.Locals("user_id").(float64))
	var result []model.ASN
	for _, a := range asns {
		if a.ID != myID {
			result = append(result, a)
		}
	}

	return c.JSON(fiber.Map{"data": result})
}

type UpdateAtasanRequest struct {
	AtasanID uint `json:"atasan_id"`
}

func (h *ASNHandler) UpdateAtasan(c *fiber.Ctx) error {
	asnID := uint(c.Locals("user_id").(float64))
	var req UpdateAtasanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	asn, err := h.repo.FindByID(asnID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User tidak ditemukan"})
	}

	if req.AtasanID != 0 {
		// Validasi: Atasan tidak boleh diri sendiri
		if req.AtasanID == asnID {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Tidak bisa menjadikan diri sendiri sebagai atasan"})
		}
		atasanID := req.AtasanID
		asn.AtasanID = &atasanID
	} else {
		asn.AtasanID = nil // Hapus atasan jika dikirim 0
	}

	// PENTING: Kosongkan struct Atasan yang ter-load (Preload) agar GORM tidak bingung
	// dan benar-benar mengupdate kolom atasan_id dengan nilai baru
	asn.Atasan = nil

	if err := h.repo.Update(asn); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal update atasan"})
	}

	return c.JSON(fiber.Map{"message": "Atasan berhasil diperbarui"})
}

// Helper function untuk membuat JWT
func generateTokens(asn *model.ASN) (string, string, error) {
	// 1. Access Token (15 Menit)
	accessClaims := jwt.MapClaims{
		"user_id":       asn.ID,
		"nip":           asn.NIP,
		"role":          asn.Role.NamaRole, // Tambahkan Role ke Token
		"organisasi_id": asn.OrganisasiID,
		"exp":           time.Now().Add(time.Minute * 15).Unix(), // Token berlaku 15 Menit
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte("rahasia_negara"))
	if err != nil {
		return "", "", err
	}

	// 2. Refresh Token (7 Hari)
	refreshClaims := jwt.MapClaims{
		"user_id": asn.ID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // Token berlaku 7 Hari
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte("rahasia_negara"))

	// Pastikan key "rahasia_negara" ini SAMA dengan yang ada di auth_middleware.go
	return accessTokenString, refreshTokenString, err
}

// --- FITUR LUPA PASSWORD (OTP) ---

var (
	otpStore = make(map[string]OTPData)
	otpMutex sync.Mutex
)

type OTPData struct {
	Code      string
	ExpiresAt time.Time
}

type ForgotPasswordRequest struct {
	NIP string `json:"nip"`
}

func (h *ASNHandler) RequestOTP(c *fiber.Ctx) error {
	var req ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	// 1. Cari ASN
	asn, err := h.repo.FindByNIP(req.NIP)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "NIP tidak ditemukan"})
	}

	// 2. Cek Email
	if asn.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email tidak terdaftar. Silakan hubungi Admin untuk update data."})
	}

	// 3. Generate OTP (6 Digit)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	otp := fmt.Sprintf("%06d", rng.Intn(1000000))

	// 4. Simpan OTP (In-Memory)
	otpMutex.Lock()
	otpStore[req.NIP] = OTPData{
		Code:      otp,
		ExpiresAt: time.Now().Add(5 * time.Minute), // Berlaku 5 menit
	}
	otpMutex.Unlock()

	// 5. Kirim Email Menggunakan Gomail
	if err := sendOTPEmail(asn.Email, asn.Nama, otp); err != nil {
		fmt.Printf("Error sending email: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengirim email OTP. Cek log server."})
	}

	// Masking email untuk response ke frontend
	maskedEmail := asn.Email
	if len(asn.Email) > 3 {
		maskedEmail = asn.Email[:3] + "****" + asn.Email[len(asn.Email)-4:]
	}

	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("Kode OTP telah dikirim ke email %s", maskedEmail),
	})
}

// Helper function untuk mengirim email menggunakan SMTP (Gomail)
func sendOTPEmail(toEmail, namaUser, otpCode string) error {
	// KONFIGURASI SMTP (Ganti dengan kredensial asli atau ambil dari ENV)
	// Jika menggunakan Gmail, pastikan menggunakan "App Password", bukan password login biasa.
	smtpHost := "smtp.gmail.com"
	smtpPort := 587
	smtpUser := "rahmanthur1@gmail.com" // GANTI DENGAN EMAIL PENGIRIM
	smtpPass := "jtso acbi quto aenp"   // GANTI DENGAN APP PASSWORD

	m := gomail.NewMessage()
	m.SetHeader("From", smtpUser)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Kode OTP Reset Password")
	m.SetBody("text/html", fmt.Sprintf(`
		<h3>Halo %s,</h3>
		<p>Anda melakukan permintaan reset password. Berikut adalah kode OTP Anda:</p>
		<h1 style="color: #3b82f6; letter-spacing: 5px;">%s</h1>
		<p>Kode ini berlaku selama 5 menit. Jangan berikan kepada siapapun.</p>
		<p>Jika ini bukan Anda, abaikan email ini.</p>
	`, namaUser, otpCode))

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	return d.DialAndSend(m)
}

type VerifyOTPRequest struct {
	NIP string `json:"nip"`
	OTP string `json:"otp"`
}

func (h *ASNHandler) VerifyOTP(c *fiber.Ctx) error {
	var req VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	otpMutex.Lock()
	data, exists := otpStore[req.NIP]
	otpMutex.Unlock()

	if !exists || time.Now().After(data.ExpiresAt) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Kode OTP kadaluwarsa atau tidak valid"})
	}

	if data.Code != req.OTP {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Kode OTP salah"})
	}

	return c.JSON(fiber.Map{"message": "OTP Valid"})
}

type ResetPasswordFinalRequest struct {
	NIP         string `json:"nip"`
	OTP         string `json:"otp"`
	NewPassword string `json:"new_password"`
}

func (h *ASNHandler) ResetPasswordFinal(c *fiber.Ctx) error {
	var req ResetPasswordFinalRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	// Validasi OTP lagi sebelum reset (Stateless Security)
	otpMutex.Lock()
	data, exists := otpStore[req.NIP]
	otpMutex.Unlock()

	if !exists || time.Now().After(data.ExpiresAt) || data.Code != req.OTP {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Sesi OTP tidak valid, silakan ulangi permintaan OTP"})
	}

	if len(req.NewPassword) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Password minimal 6 karakter"})
	}

	// Update Password
	asn, err := h.repo.FindByNIP(req.NIP)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Pegawai tidak ditemukan"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengenkripsi password"})
	}

	asn.Password = string(hashedPassword)
	if err := h.repo.Update(asn); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal reset password"})
	}

	// Hapus OTP agar tidak bisa dipakai lagi
	otpMutex.Lock()
	delete(otpStore, req.NIP)
	otpMutex.Unlock()

	return c.JSON(fiber.Map{"message": "Password berhasil diubah. Silakan login kembali."})
}

// --- FITUR UPLOAD FOTO PROFILE ---
func (h *ASNHandler) UploadFotoProfile(c *fiber.Ctx) error {
	asnID := uint(c.Locals("user_id").(float64))

	// 1. Ambil File dari Request
	file, err := c.FormFile("foto") // Key: "foto"
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File foto wajib diupload"})
	}

	// 2. Validasi Ukuran & Tipe (Opsional)
	// Max 2MB
	if file.Size > 2097152 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Ukuran file maksimal 2MB"})
	}

	// 3. Simpan File ke Folder ./uploads/profile
	// Pastikan folder ada
	uploadDir := "./uploads/profile"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
	}

	// Generate Nama Unik
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%d%s", asnID, time.Now().Unix(), ext)
	pathFile := fmt.Sprintf("uploads/profile/%s", filename)

	if err := c.SaveFile(file, pathFile); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan file"})
	}

	// 4. Update Database
	asn, err := h.repo.FindByID(asnID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User tidak ditemukan"})
	}

	asn.Foto = pathFile // Simpan relative path untuk diakses via URL
	if err := h.repo.Update(asn); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal update foto di database"})
	}

	return c.JSON(fiber.Map{
		"message": "Foto profil berhasil diupload",
		"data":    asn,
	})
}

// GetFotoProfile: Mengambil file foto profil berdasarkan ID ASN
func (h *ASNHandler) GetFotoProfile(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	asn, err := h.repo.FindByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("User tidak ditemukan")
	}

	// Jika foto kosong atau file tidak ada, return placeholder default (opsional)
	if asn.Foto == "" {
		return c.Status(fiber.StatusNotFound).SendString("Foto belum diatur")
	}

	// Cek apakah file benar-benar ada di disk
	if _, err := os.Stat(asn.Foto); os.IsNotExist(err) {
		// Jika path di DB ada tapi file fisik hilang
		return c.Status(fiber.StatusNotFound).SendString("File foto tidak ditemukan")
	}

	// Serve Static File
	return c.SendFile(asn.Foto)
}

// GetSubordinates: Mengambil list pegawai yang atasan-nya adalah user yang sedang login
func (h *ASNHandler) GetSubordinates(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	asns, err := h.repo.GetByAtasanID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data bawahan"})
	}

	return c.JSON(fiber.Map{"data": asns})
}
