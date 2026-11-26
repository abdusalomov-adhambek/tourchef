package main

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type RegionJSON struct {
	ID           string   `json:"id"`
	Slug         string   `json:"slug"`
	GeoID        string   `json:"geo_id"`
	Name         string   `json:"name"`
	NameUZ       string   `json:"name_uz"`
	City         string   `json:"city"`
	DishName     string   `json:"dish_name"`
	DishImageURL string   `json:"dish_image_url"`
	Description  string   `json:"description"`
	Ingredients  []string `json:"ingredients"`
	CulturalInfo string   `json:"cultural_info"`
	Latitude     float64  `json:"latitude"`
	Longitude    float64  `json:"longitude"`
	Status       string   `json:"status"`
	CreatedAt    string   `json:"created_at"`
}

const regionCols = `id::text, slug, geo_id, name, name_uz, city, dish_name, dish_image_url,
	description, COALESCE(ingredients, '[]'::jsonb)::text, cultural_info, latitude, longitude, status, created_at::text`

func scanRegion(s scanner) (RegionJSON, error) {
	var r RegionJSON
	var ing string
	err := s.Scan(&r.ID, &r.Slug, &r.GeoID, &r.Name, &r.NameUZ, &r.City, &r.DishName,
		&r.DishImageURL, &r.Description, &ing, &r.CulturalInfo, &r.Latitude, &r.Longitude,
		&r.Status, &r.CreatedAt)
	if err != nil {
		return r, err
	}
	r.Ingredients = []string{}
	if ing != "" {
		_ = json.Unmarshal([]byte(ing), &r.Ingredients)
	}
	return r, nil
}

func handleRegions(c *fiber.Ctx) error {
	rows, err := db.Query(c.Context(), `SELECT `+regionCols+` FROM region_foods
		WHERE status='published' ORDER BY name`)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	defer rows.Close()
	out := make([]RegionJSON, 0)
	for rows.Next() {
		r, err := scanRegion(rows)
		if err != nil {
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		out = append(out, r)
	}
	return c.JSON(out)
}

func findRegion(c *fiber.Ctx, slug string, publishedOnly bool) (RegionJSON, error) {
	q := `SELECT ` + regionCols + ` FROM region_foods WHERE slug=$1`
	if publishedOnly {
		q += ` AND status='published'`
	}
	return scanRegion(db.QueryRow(c.Context(), q, slug))
}

func handleRegionDetail(c *fiber.Ctx) error {
	r, err := findRegion(c, c.Params("slug"), true)
	if err != nil {
		if err == sql.ErrNoRows {
			return fail(c, fiber.StatusNotFound, "region not found")
		}
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(r)
}

func handleRegionDish(c *fiber.Ctx) error {
	r, err := findRegion(c, c.Params("slug"), true)
	if err != nil {
		if err == sql.ErrNoRows {
			return fail(c, fiber.StatusNotFound, "region not found")
		}
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(fiber.Map{
		"name": r.DishName, "image_url": r.DishImageURL, "description": r.Description,
		"ingredients": r.Ingredients, "cultural_info": r.CulturalInfo, "region": r.Name,
	})
}

func handleRegionMasterclasses(c *fiber.Ctx) error {
	r, err := findRegion(c, c.Params("slug"), true)
	if err != nil {
		if err == sql.ErrNoRows {
			return fail(c, fiber.StatusNotFound, "region not found")
		}
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	f := &sqlFilter{}
	f.add(`m.city ILIKE ?`, "%"+r.City+"%")
	limit := queryInt(c, "limit", 12)
	if limit < 1 || limit > 50 {
		limit = 12
	}
	items, total, err := listMasterclasses(c.Context(), f, ` ORDER BY m.rating DESC, m.created_at DESC`, limit, 0)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(fiber.Map{"items": items, "total": total})
}

func handleRegionRestaurants(c *fiber.Ctx) error {
	r, err := findRegion(c, c.Params("slug"), true)
	if err != nil {
		if err == sql.ErrNoRows {
			return fail(c, fiber.StatusNotFound, "region not found")
		}
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	rows, err := db.Query(c.Context(), `SELECT `+restaurantSelectCols+` FROM restaurants r
		WHERE r.city ILIKE $1 AND r.status='approved' ORDER BY r.rating DESC`, "%"+r.City+"%")
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	defer rows.Close()
	out := make([]RestaurantJSON, 0)
	for rows.Next() {
		x, err := scanRestaurant(rows)
		if err != nil {
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		out = append(out, x)
	}
	return c.JSON(out)
}

// ---------------- Admin CRUD ----------------

func handleAdminRegions(c *fiber.Ctx) error {
	rows, err := db.Query(c.Context(), `SELECT `+regionCols+` FROM region_foods ORDER BY name`)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	defer rows.Close()
	out := make([]RegionJSON, 0)
	for rows.Next() {
		r, err := scanRegion(rows)
		if err != nil {
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		out = append(out, r)
	}
	return c.JSON(out)
}

type regionBody struct {
	GeoID        string   `json:"geo_id"`
	Name         string   `json:"name"`
	NameUZ       string   `json:"name_uz"`
	City         string   `json:"city"`
	DishName     string   `json:"dish_name"`
	DishImageURL string   `json:"dish_image_url"`
	Description  string   `json:"description"`
	Ingredients  []string `json:"ingredients"`
	CulturalInfo string   `json:"cultural_info"`
	Latitude     float64  `json:"latitude"`
	Longitude    float64  `json:"longitude"`
	Status       string   `json:"status"`
}

func handleAdminCreateRegion(c *fiber.Ctx) error {
	var b regionBody
	if err := c.BodyParser(&b); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	if strings.TrimSpace(b.Name) == "" || strings.TrimSpace(b.DishName) == "" {
		return fail(c, fiber.StatusBadRequest, "name and dish_name are required")
	}
	if b.Status == "" {
		b.Status = "published"
	}
	if b.Status != "published" && b.Status != "hidden" {
		return fail(c, fiber.StatusBadRequest, "status must be published or hidden")
	}
	ing, err := json.Marshal(b.Ingredients)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	r, err := scanRegion(db.QueryRow(c.Context(), `
		INSERT INTO region_foods(slug, geo_id, name, name_uz, city, dish_name, dish_image_url,
			description, ingredients, cultural_info, latitude, longitude, status)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10, $11, $12, $13)
		RETURNING `+regionCols,
		uniqueSlug(c.Context(), db, "region_foods", b.Name), b.GeoID, b.Name, b.NameUZ, b.City,
		b.DishName, b.DishImageURL, b.Description, string(ing), b.CulturalInfo,
		b.Latitude, b.Longitude, b.Status))
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.Status(fiber.StatusCreated).JSON(r)
}

func handleAdminUpdateRegion(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err.Error())
	}
	var b regionBody
	if err := c.BodyParser(&b); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	type assignment struct {
		col  string
		val  any
		cast string
	}
	sets := []assignment{}
	add := func(col string, val any) { sets = append(sets, assignment{col: col, val: val}) }
	if b.Name != "" {
		add("name", b.Name)
	}
	if b.NameUZ != "" {
		add("name_uz", b.NameUZ)
	}
	if b.City != "" {
		add("city", b.City)
	}
	if b.DishName != "" {
		add("dish_name", b.DishName)
	}
	if b.DishImageURL != "" {
		add("dish_image_url", b.DishImageURL)
	}
	if b.Description != "" {
		add("description", b.Description)
	}
	if b.CulturalInfo != "" {
		add("cultural_info", b.CulturalInfo)
	}
	if b.GeoID != "" {
		add("geo_id", b.GeoID)
	}
	if b.Latitude != 0 {
		add("latitude", b.Latitude)
	}
	if b.Longitude != 0 {
		add("longitude", b.Longitude)
	}
	if b.Status != "" {
		if b.Status != "published" && b.Status != "hidden" {
			return fail(c, fiber.StatusBadRequest, "invalid status")
		}
		add("status", b.Status)
	}
	if len(b.Ingredients) > 0 {
		raw, _ := json.Marshal(b.Ingredients)
		sets = append(sets, assignment{col: "ingredients", val: string(raw), cast: "::jsonb"})
	}
	if len(sets) == 0 {
		return fail(c, fiber.StatusBadRequest, "no fields to update")
	}
	query := "UPDATE region_foods SET "
	for i, s := range sets {
		if i > 0 {
			query += ", "
		}
		query += s.col + " = $" + itoa(i+1) + s.cast
	}
	query += ", updated_at=now() WHERE id=$" + itoa(len(sets)+1) + "::uuid RETURNING " + regionCols
	args := make([]any, 0, len(sets)+1)
	for _, s := range sets {
		args = append(args, s.val)
	}
	args = append(args, id)
	r, err := scanRegion(db.QueryRow(c.Context(), query, args...))
	if err != nil {
		if err == sql.ErrNoRows {
			return fail(c, fiber.StatusNotFound, "region not found")
		}
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(r)
}

func handleAdminDeleteRegion(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err.Error())
	}
	tag, err := db.Exec(c.Context(), `DELETE FROM region_foods WHERE id=$1::uuid`, id)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	if tag.RowsAffected() == 0 {
		return fail(c, fiber.StatusNotFound, "region not found")
	}
	return c.JSON(fiber.Map{"ok": true})
}
