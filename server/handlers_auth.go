package main

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// flexibleString qabul qiladi: "suv, tuz" yoki ["suv", "tuz"].
type flexibleString struct{ Value string }

func (f *flexibleString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		f.Value = ""
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		f.Value = s
		return nil
	}
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		f.Value = strings.Join(arr, ", ")
		return nil
	}
	f.Value = strings.Trim(string(data), `"`)
	return nil
}

const dishReturnCols = `id::text, name, slug, category, region, description, ingredients,
	image_url, rating::float8, status,
	CASE WHEN created_by IS NULL THEN NULL ELSE created_by::text END, created_at,
	COALESCE(multiimage, '[]'::jsonb)::text`

func fetchMasterclassByID(c *fiber.Ctx, id string) (MasterclassJSON, error) {
	return scanMasterclass(db.QueryRow(c.Context(), mcSelect+` WHERE m.id=$1::uuid`, id))
}

// ---------------------------------------------------------------------------
// Dishes (admin)
// ---------------------------------------------------------------------------

type dishCreateBody struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Region      string `json:"region"`
	Description string `json:"description"`
	Ingredients flexibleString `json:"ingredients"`
	ImageURL    string         `json:"image_url"`
	Status      string         `json:"status"`
	MultiImage  []string       `json:"multiimage"`
}

type dishPatchBody struct {
	Name        *string         `json:"name"`
	Category    *string         `json:"category"`
	Region      *string         `json:"region"`
	Description *string         `json:"description"`
	Ingredients *flexibleString `json:"ingredients"`
	ImageURL    *string         `json:"image_url"`
	Status      *string         `json:"status"`
	MultiImage  *[]string       `json:"multiimage"`
}

func handleCreateDish(c *fiber.Ctx) error {
	var body dishCreateBody
	if err := c.BodyParser(&body); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	if body.Name = trim(body.Name); body.Name == "" {
		return fail(c, fiber.StatusBadRequest, "name is required")
	}
	if body.Status == "" {
		body.Status = "active"
	}
	if body.Status != "active" && body.Status != "hidden" {
		return fail(c, fiber.StatusBadRequest, "invalid status")
	}
	slug := uniqueSlug(c.Context(), db, "dishes", body.Name)
	user := currentUser(c)

	images := body.MultiImage
	if len(images) == 0 && body.ImageURL != "" {
		images = []string{body.ImageURL}
	}
	if len(images) > 0 && body.ImageURL == "" {
		body.ImageURL = images[0]
	}
	multiJSON, err := json.Marshal(images)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}

	d, err := scanDish(db.QueryRow(c.Context(), `
		INSERT INTO dishes(name, slug, category, region, description, ingredients, image_url, status, created_by, multiimage)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9::uuid, $10::jsonb)
		RETURNING `+dishReturnCols,
		body.Name, slug, body.Category, body.Region, body.Description, body.Ingredients.Value,
		body.ImageURL, body.Status, user.ID, string(multiJSON)))
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.Status(fiber.StatusCreated).JSON(d)
}

func handleUpdateDish(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err.Error())
	}
	var body dishPatchBody
	if err := c.BodyParser(&body); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid JSON body")
	}

	type assignment struct {
		col  string
		val  any
		cast string
	}
	sets := make([]assignment, 0, 8)
	add := func(col string, val any) { sets = append(sets, assignment{col: col, val: val}) }
	addJSON := func(col string, val any) { sets = append(sets, assignment{col: col, val: val, cast: "::jsonb"}) }
	if body.Name != nil {
		if *body.Name = trim(*body.Name); *body.Name == "" {
			return fail(c, fiber.StatusBadRequest, "name cannot be empty")
		}
		add("name", *body.Name)
		var currentSlug string
		if err := db.QueryRow(c.Context(), `SELECT slug FROM dishes WHERE id=$1::uuid`, id).Scan(&currentSlug); err != nil {
			if err == pgx.ErrNoRows {
				return fail(c, fiber.StatusNotFound, "dish not found")
			}
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		if slugify(*body.Name) != currentSlug {
			add("slug", uniqueSlug(c.Context(), db, "dishes", *body.Name))
		}
	}
	if body.Category != nil {
		add("category", *body.Category)
	}
	if body.Region != nil {
		add("region", *body.Region)
	}
	if body.Description != nil {
		add("description", *body.Description)
	}
	if body.Ingredients != nil {
		add("ingredients", body.Ingredients.Value)
	}
	if body.ImageURL != nil {
		add("image_url", *body.ImageURL)
	}
	if body.MultiImage != nil {
		images := *body.MultiImage
		raw, err := json.Marshal(images)
		if err != nil {
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		addJSON("multiimage", string(raw))
		if body.ImageURL == nil && len(images) > 0 {
			add("image_url", images[0])
		}
	}
	if body.Status != nil {
		if *body.Status != "active" && *body.Status != "hidden" && *body.Status != "deleted" {
			return fail(c, fiber.StatusBadRequest, "invalid status")
		}
		add("status", *body.Status)
	}
	if len(sets) == 0 {
		return fail(c, fiber.StatusBadRequest, "no fields to update")
	}

	query := "UPDATE dishes SET "
	for i, s := range sets {
		if i > 0 {
			query += ", "
		}
		query += s.col + " = $" + itoa(i+1) + s.cast
	}
	query += ", updated_at = now() WHERE id=$" + itoa(len(sets)+1) + "::uuid RETURNING " + dishReturnCols
	args := make([]any, 0, len(sets)+1)
	for _, s := range sets {
		args = append(args, s.val)
	}
	args = append(args, id)

	d, err := scanDish(db.QueryRow(c.Context(), query, args...))
	if err == pgx.ErrNoRows {
		return fail(c, fiber.StatusNotFound, "dish not found")
	}
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(d)
}

func handleDeleteDish(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err.Error())
	}
	tag, err := db.Exec(c.Context(), `DELETE FROM dishes WHERE id=$1::uuid`, id)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	if tag.RowsAffected() == 0 {
		return fail(c, fiber.StatusNotFound, "dish not found")
	}
	return c.JSON(fiber.Map{"ok": true})
}

func trim(s string) string {
	out := s
	for len(out) > 0 && (out[0] == ' ' || out[0] == '\t' || out[0] == '\n' || out[0] == '\r') {
		out = out[1:]
	}
	for len(out) > 0 {
		last := out[len(out)-1]
		if last == ' ' || last == '\t' || last == '\n' || last == '\r' {
			out = out[:len(out)-1]
		} else {
			break
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Chef masterclasses and sessions
// ---------------------------------------------------------------------------

type mcCreateBody struct {
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	DishID       string  `json:"dish_id"`
	City         string  `json:"city"`
	DurationMin  int     `json:"duration_min"`
	Price        float64 `json:"price"`
	Difficulty   string  `json:"difficulty"`
	Language     string  `json:"language"`
	ImageURL     string  `json:"image_url"`
	RestaurantID string  `json:"restaurant_id"`
	Status       string  `json:"status"`
	ChefID       string  `json:"chef_id"` // admin only
}

// chefIDForCaller resolves the chefs row for the authenticated caller.
func chefIDForCaller(c *fiber.Ctx) (string, error) {
	u := currentUser(c)
	var id string
	err := db.QueryRow(c.Context(), `SELECT id::text FROM chefs WHERE user_id=$1::uuid`, u.ID).Scan(&id)
	return id, err
}

func handleCreateMasterclass(c *fiber.Ctx) error {
	var body mcCreateBody
	if err := c.BodyParser(&body); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	if body.Title = trim(body.Title); body.Title == "" {
		return fail(c, fiber.StatusBadRequest, "title is required")
	}
	if !uuidPattern.MatchString(body.DishID) {
		return fail(c, fiber.StatusBadRequest, "valid dish_id is required")
	}
	if body.DurationMin <= 0 {
		body.DurationMin = 120
	}
	if body.Price < 0 {
		return fail(c, fiber.StatusBadRequest, "price cannot be negative")
	}
	if body.Difficulty == "" {
		body.Difficulty = "beginner"
	}
	if body.Language == "" {
		body.Language = "uz"
	}

	user := currentUser(c)
	chefID := body.ChefID
	status := body.Status
	if user.Role == "chef" {
		own, err := chefIDForCaller(c)
		if err == pgx.ErrNoRows {
			return fail(c, fiber.StatusForbidden, "chef profile not found")
		}
		if err != nil {
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		chefID = own
		// Chefs may submit drafts/pending items but cannot self-approve.
		if status != "draft" && status != "pending" {
			status = "pending"
		}
	} else {
		if !uuidPattern.MatchString(chefID) {
			if err := db.QueryRow(c.Context(), `SELECT id::text FROM chefs WHERE status='approved'
				ORDER BY rating DESC LIMIT 1`).Scan(&chefID); err != nil {
				return fail(c, fiber.StatusBadRequest, "no chef available for this masterclass")
			}
		}
		if status == "" {
			status = "approved"
		}
	}
	if status != "draft" && status != "pending" && status != "approved" {
		status = "pending"
	}

	var restaurantID any
	if uuidPattern.MatchString(body.RestaurantID) {
		restaurantID = body.RestaurantID
	}

	slug := uniqueSlug(c.Context(), db, "masterclasses", body.Title)
	var id string
	err := db.QueryRow(c.Context(), `
		INSERT INTO masterclasses(title, slug, description, dish_id, chef_id, restaurant_id,
			city, duration_min, price, difficulty, language, image_url, status)
		VALUES($1, $2, $3, $4::uuid, $5::uuid, $6::uuid, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id::text`,
		body.Title, slug, body.Description, body.DishID, chefID, restaurantID,
		body.City, body.DurationMin, body.Price, body.Difficulty, body.Language,
		body.ImageURL, status).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			return fail(c, fiber.StatusConflict, "masterclass slug already exists")
		}
		return failErr(c, fiber.StatusInternalServerError, err)
	}

	mc, err := fetchMasterclassByID(c, id)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.Status(fiber.StatusCreated).JSON(mc)
}

func handleChefMasterclasses(c *fiber.Ctx) error {
	chefID, err := chefIDForCaller(c)
	if err == pgx.ErrNoRows {
		return c.JSON([]MasterclassJSON{})
	}
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	f := &sqlFilter{}
	f.add(`m.chef_id = ?`, chefID)
	items, _, err := listMasterclasses(c.Context(), f, ` ORDER BY m.created_at DESC`, 100, 0)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(items)
}

type sessionCreateBody struct {
	MasterclassID string    `json:"masterclass_id"`
	StartsAt      time.Time `json:"starts_at"`
	Capacity      int       `json:"capacity"`
	Price         float64   `json:"price"`
	Status        string    `json:"status"`
}

const sessionReturnCols = `id::text, masterclass_id::text, starts_at, capacity, booked, price::float8, status`

func handleCreateSession(c *fiber.Ctx) error {
	var body sessionCreateBody
	if err := c.BodyParser(&body); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	if !uuidPattern.MatchString(body.MasterclassID) {
		return fail(c, fiber.StatusBadRequest, "valid masterclass_id is required")
	}
	if body.StartsAt.IsZero() {
		return fail(c, fiber.StatusBadRequest, "starts_at is required (RFC3339)")
	}
	if body.Capacity <= 0 {
		return fail(c, fiber.StatusBadRequest, "capacity must be greater than zero")
	}
	if body.Price < 0 {
		return fail(c, fiber.StatusBadRequest, "price cannot be negative")
	}
	if body.Status == "" {
		body.Status = "scheduled"
	}
	if body.Status != "scheduled" && body.Status != "full" && body.Status != "completed" && body.Status != "cancelled" {
		return fail(c, fiber.StatusBadRequest, "invalid status")
	}

	user := currentUser(c)
	chefID, err := chefIDForCaller(c)
	if err != nil {
		return fail(c, fiber.StatusForbidden, "chef profile not found")
	}
	var ownerCheck int
	err = db.QueryRow(c.Context(), `SELECT COUNT(*) FROM masterclasses m
		WHERE m.id=$1::uuid AND (m.chef_id=$2::uuid OR $3='admin')`,
		body.MasterclassID, chefID, user.Role).Scan(&ownerCheck)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	if ownerCheck == 0 {
		return fail(c, fiber.StatusForbidden, "masterclass does not belong to this chef")
	}

	if body.Price == 0 {
		if err := db.QueryRow(c.Context(), `SELECT price::float8 FROM masterclasses WHERE id=$1::uuid`,
			body.MasterclassID).Scan(&body.Price); err != nil {
			return failErr(c, fiber.StatusInternalServerError, err)
		}
	}

	s, err := scanSession(db.QueryRow(c.Context(), `
		INSERT INTO sessions(masterclass_id, starts_at, capacity, price, status)
		VALUES($1::uuid, $2, $3, $4, $5)
		RETURNING `+sessionReturnCols,
		body.MasterclassID, body.StartsAt, body.Capacity, body.Price, body.Status))
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.Status(fiber.StatusCreated).JSON(s)
}

func handleChefSessions(c *fiber.Ctx) error {
	chefID, err := chefIDForCaller(c)
	if err == pgx.ErrNoRows {
		return c.JSON([]SessionJSON{})
	}
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return listSessionsBy(c, `
		WHERE m.chef_id=$1::uuid`, chefID)
}

func listSessionsBy(c *fiber.Ctx, where string, args ...any) error {
	query := `SELECT ` + sessionSelectCols + ` FROM sessions s
		JOIN masterclasses m ON m.id = s.masterclass_id ` + where + ` ORDER BY s.starts_at ASC`
	rows, err := db.Query(c.Context(), query, args...)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	defer rows.Close()
	out := make([]SessionJSON, 0)
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(out)
}
