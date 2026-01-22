package middleware

import "github.com/gofiber/fiber/v2"

func Role(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Ambil role user dari context (diset di Auth middleware)
		userRole, ok := c.Locals("role").(string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Akses ditolak: Role tidak valid"})
		}

		for _, role := range allowedRoles {
			if role == userRole {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Akses ditolak: Anda bukan Admin"})
	}
}
