package handler

import (
	"my-flutter-backend/internal/model"
	"my-flutter-backend/internal/repository"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type RoleHandler struct {
	repo repository.RoleRepository
}

func NewRoleHandler(repo repository.RoleRepository) *RoleHandler {
	return &RoleHandler{repo: repo}
}

func (h *RoleHandler) GetAll(c *fiber.Ctx) error {
	roles, err := h.repo.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data role"})
	}

	// Filter: Sembunyikan "Super Admin"
	var filteredRoles []model.Role
	for _, r := range roles {
		if r.NamaRole != "Super Admin" {
			filteredRoles = append(filteredRoles, r)
		}
	}

	return c.JSON(fiber.Map{"data": filteredRoles})
}

func (h *RoleHandler) GetAllPermissions(c *fiber.Ctx) error {
	perms, err := h.repo.GetAllPermissions()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data permission"})
	}

	// Filter: Sembunyikan "kelola_organisasi"
	var filteredPerms []model.Permission
	for _, p := range perms {
		if p.NamaPermission != "kelola_organisasi" {
			filteredPerms = append(filteredPerms, p)
		}
	}

	return c.JSON(fiber.Map{"data": filteredPerms})
}

func (h *RoleHandler) GetDetail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	role, err := h.repo.GetByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Role tidak ditemukan"})
	}
	return c.JSON(fiber.Map{"data": role})
}

type RoleRequest struct {
	NamaRole      string `json:"nama_role"`
	PermissionIDs []uint `json:"permission_ids"`
}

func (h *RoleHandler) Create(c *fiber.Ctx) error {
	var req RoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	role := model.Role{NamaRole: req.NamaRole}
	if err := h.repo.Create(&role, req.PermissionIDs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuat role"})
	}

	return c.JSON(fiber.Map{"message": "Role berhasil dibuat"})
}

func (h *RoleHandler) Update(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var req RoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	role, err := h.repo.GetByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Role tidak ditemukan"})
	}

	role.NamaRole = req.NamaRole
	if err := h.repo.Update(role, req.PermissionIDs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal update role"})
	}

	return c.JSON(fiber.Map{"message": "Role berhasil diupdate"})
}

func (h *RoleHandler) Delete(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.repo.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menghapus role"})
	}
	return c.JSON(fiber.Map{"message": "Role berhasil dihapus"})
}
