package middleware

import (
	"my-flutter-backend/internal/model"

	"github.com/gofiber/fiber/v2"
)

func Permission(requiredPermission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Ambil Role user hari Context (Diset di Auht middleware)
		userRole, ok := c.Locals("role").(string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Akses ditolak: Role tidak valid"})
		}

		// 2. Jika Super Admin, Bypass Permission Check (Opsional, tapi aman)
		if userRole == "Super Admin" {
			return c.Next()
		}

		// 3. Cek Permission ke Database
		var role model.Role
		// Preload Permissions agar bisa dicek
		if err := DB.Preload("Permissions").Where("nama_role = ?", userRole).First(&role).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal memvalidasi permission"})
		}

		// 4. Loop permissions role ini
		isAllowed := false
		for _, p := range role.Permissions {
			if p.NamaPermission == requiredPermission {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Akses ditolak: Anda tidak memiliki izin " + requiredPermission})
		}

		return c.Next()
	}
}
