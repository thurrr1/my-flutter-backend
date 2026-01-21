package handler

import (
	"my-flutter-backend/internal/model"
	"my-flutter-backend/internal/repository"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
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
	token, err := generateToken(asn)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuat token"})
	}

	// 4. Return Token ke Client
	return c.JSON(fiber.Map{
		"message": "Login berhasil",
		"token":   token, // <--- Token ini yang nanti dicopy ke Bruno
		"data": fiber.Map{
			"nip":        asn.NIP,
			"nama":       asn.Nama,
			"role":       asn.Role.NamaRole,
			"jabatan":    asn.Jabatan,
			"organisasi": asn.Organisasi.NamaOrganisasi, // Tambahan untuk Dashboard
		},
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
		asn.Foto = req.Foto
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

	asns, err := h.repo.GetAll()
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

func (h *ASNHandler) ResetDevice(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.repo.ResetDevice(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal reset device"})
	}
	return c.JSON(fiber.Map{"message": "Device berhasil di-reset, user bisa login di HP baru"})
}

// Helper function untuk membuat JWT
func generateToken(asn *model.ASN) (string, error) {
	claims := jwt.MapClaims{
		"user_id":       asn.ID,
		"nip":           asn.NIP,
		"role":          asn.Role.NamaRole, // Tambahkan Role ke Token
		"organisasi_id": asn.OrganisasiID,
		"exp":           time.Now().Add(time.Hour * 24).Unix(), // Token berlaku 24 jam
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Pastikan key "rahasia_negara" ini SAMA dengan yang ada di auth_middleware.go
	return token.SignedString([]byte("rahasia_negara"))
}
