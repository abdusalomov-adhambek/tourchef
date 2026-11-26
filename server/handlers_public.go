package main

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// ---------------------------------------------------------------------------
// Masterclasses
// ---------------------------------------------------------------------------

const mcSelect = `
SELECT m.id::text, m.title, m.slug, m.description,
       CASE WHEN m.dish_id IS NULL THEN NULL ELSE m.dish_id::text END,
       COALESCE(d.name, ''), COALESCE(d.slug, ''), COALESCE(d.image_url, ''),
       COALESCE(d.category, ''), COALESCE(d.region, ''),
       CASE WHEN m.chef_id IS NULL THEN NULL ELSE m.chef_id::text END,
       COALESCE(c.display_name, ''), COALESCE(c.avatar_url, ''), COALESCE(c.city, ''),
       COALESCE(c.rating::float8, 0),
       CASE WHEN m.restaurant_id IS NULL THEN NULL ELSE m.restaurant_id::text END,
       COALESCE(r.name, ''), COALESCE(r.slug, ''), COALESCE(r.city, ''),
       COALESCE(r.image_url, ''), COALESCE(r.rating::float8, 0),
       m.city, m.duration_min, m.price::float8, m.difficulty, m.language,
       m.rating::float8,
       (SELECT COUNT(*)::int FROM reviews rv WHERE rv.masterclass_id = m.id),
       m.image_url, m.status, m.created_at
FROM masterclasses m
LEFT JOIN dishes d ON d.id = m.dish_id
LEFT JOIN chefs c ON c.id = m.chef_id
LEFT JOIN restaurants r ON r.id = m.restaurant_id`

func scanMasterclass(s scanner) (MasterclassJSON, error) {
	var m MasterclassJSON
	var dishID, chefID, restID sql.NullString
	var dishSlug, dishImage, dishCat, dishRegion string
	var chefAvatar, chefCity string
	var chefRating float64
	var restSlug, restCity, restImage string
	var restRating float64

	err := s.Scan(
		&m.ID, &m.Title, &m.Slug, &m.Description,
		&dishID, &m.DishName, &dishSlug, &dishImage, &dishCat, &dishRegion,
		&chefID, &m.ChefName, &chefAvatar, &chefCity, &chefRating,
		&restID, &m.RestName, &restSlug, &restCity, &restImage, &restRating,
		&m.City, &m.DurationMin, &m.Price, &m.Difficulty, &m.Language, &m.Rating,
		&m.ReviewsCount, &m.ImageURL, &m.Status, &m.CreatedAt,
	)
	if err != nil {
		return m, err
	}

	if dishID.Valid {
		m.DishID = strPtr(dishID.String)
		m.Dish = map[string]any{
			"id": dishID.String, "name": m.DishName, "slug": dishSlug,
			"image_url": dishImage, "category": dishCat, "region": dishRegion,
		}
	}
	if chefID.Valid {
		m.ChefID = strPtr(chefID.String)
		m.Chef = map[string]any{
			"id": chefID.String, "display_name": m.ChefName,
			"avatar_url": chefAvatar, "city": chefCity, "rating": chefRating,
		}
	}
	if restID.Valid {
		m.RestaurantID = strPtr(restID.String)
		m.Restaurant = map[string]any{
			"id": restID.String, "name": m.RestName, "slug": restSlug,
			"city": restCity, "image_url": restImage, "rating": restRating,
		}
	}
	return m, nil
}

// applyMasterclassFilters translates query-string filters into SQL.
func applyMasterclassFilters(c *fiber.Ctx, f *sqlFilter, publicOnly bool) {
	if publicOnly {
		f.add(`m.status = 'approved'`)
	}

	if q := strings.TrimSpace(c.Query("q")); q != "" {
		needle := "%" + q + "%"
		f.add(`(m.title ILIKE ? OR m.description ILIKE ? OR d.name ILIKE ? OR c.display_name ILIKE ?)`,
			needle, needle, needle, needle)
	}
	if city := strings.TrimSpace(c.Query("city")); city != "" {
		f.add(`m.city ILIKE ?`, "%"+city+"%")
	}
	if dish := strings.TrimSpace(c.Query("dish")); dish != "" {
		f.add(`(d.name ILIKE ? OR d.slug = lower(?) OR m.dish_id::text = ? OR d.category ILIKE ?)`,
			"%"+dish+"%", dish, dish, "%"+dish+"%")
	}
	if category := strings.TrimSpace(c.Query("category")); category != "" {
		f.add(`d.category ILIKE ?`, "%"+category+"%")
	}
	if min := queryFloat(c, "min_price", 0); min > 0 {
		f.add(`m.price >= ?`, min)
	}
	if max := queryFloat(c, "max_price", 0); max > 0 {
		f.add(`m.price <= ?`, max)
	}
	if duration := c.Query("duration"); duration != "" {
		switch duration {
		case "1-2":
			f.add(`m.duration_min BETWEEN 60 AND 120`)
		case "2-4":
			f.add(`m.duration_min > 120 AND m.duration_min <= 240`)
		case "4+":
			f.add(`m.duration_min > 240`)
		}
	}
	if rating := queryFloat(c, "rating", 0); rating > 0 {
		f.add(`m.rating >= ?`, rating)
	}
	if date := strings.TrimSpace(c.Query("date")); date != "" {
		f.add(`EXISTS (SELECT 1 FROM sessions s WHERE s.masterclass_id = m.id
			AND s.status <> 'cancelled' AND s.starts_at::date = ?::date)`, date)
	}
}

func masterclassOrder(c *fiber.Ctx) string {
	switch c.Query("sort") {
	case "price_asc":
		return ` ORDER BY m.price ASC, m.created_at DESC`
	case "price_desc":
		return ` ORDER BY m.price DESC, m.created_at DESC`
	case "rating":
		return ` ORDER BY m.rating DESC, (SELECT COUNT(*) FROM reviews rv WHERE rv.masterclass_id = m.id) DESC, m.created_at DESC`
	case "new":
		return ` ORDER BY m.created_at DESC`
	default: // rec
		return ` ORDER BY m.rating DESC, (SELECT COUNT(*) FROM reviews rv WHERE rv.masterclass_id = m.id) DESC, m.created_at DESC`
	}
}

func listMasterclasses(ctx context.Context, f *sqlFilter, orderBy string, limit, offset int) ([]MasterclassJSON, int, error) {
	var total int
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM masterclasses m
		LEFT JOIN dishes d ON d.id = m.dish_id
		LEFT JOIN chefs c ON c.id = m.chef_id
		LEFT JOIN restaurants r ON r.id = m.restaurant_id`+f.where(), f.args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := mcSelect + f.where() + orderBy + ` LIMIT $LIMIT OFFSET $OFFSET`
	query = strings.Replace(query, "$LIMIT", itoa(limit), 1)
	query = strings.Replace(query, "$OFFSET", itoa(offset), 1)

	rows, err := db.Query(ctx, query, f.args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]MasterclassJSON, 0)
	for rows.Next() {
		mc, err := scanMasterclass(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, mc)
	}
	return items, total, rows.Err()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func handleMasterclasses(c *fiber.Ctx) error {
	f := &sqlFilter{}
	applyMasterclassFilters(c, f, true)

	page := queryInt(c, "page", 1)
	if page < 1 {
		page = 1
	}
	limit := queryInt(c, "limit", 12)
	if limit < 1 || limit > 100 {
		limit = 12
	}
	offset := (page - 1) * limit

	items, total, err := listMasterclasses(c.Context(), f, masterclassOrder(c), limit, offset)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	pages := 0
	if total > 0 {
		pages = (total + limit - 1) / limit
	}
	return c.JSON(fiber.Map{"items": items, "total": total, "page": page, "pages": pages})
}

func sessionsForMasterclass(ctx context.Context, id string) ([]SessionJSON, error) {
	rows, err := db.Query(ctx, `SELECT `+sessionSelectCols+` FROM sessions s
		WHERE s.masterclass_id=$1::uuid ORDER BY s.starts_at ASC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]SessionJSON, 0)
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func handleMasterclassDetail(c *fiber.Ctx) error {
	slug := c.Params("slug")
	mc, err := scanMasterclass(db.QueryRow(c.Context(), mcSelect+` WHERE m.slug=$1 AND m.status='approved'`, slug))
	if err == pgx.ErrNoRows {
		return fail(c, fiber.StatusNotFound, "masterclass not found")
	}
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	sessions, err := sessionsForMasterclass(c.Context(), mc.ID)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	mc.Sessions = sessions
	return c.JSON(mc)
}

func handleMasterclassSessions(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var id string
	if err := db.QueryRow(c.Context(), `SELECT id::text FROM masterclasses WHERE slug=$1 AND status='approved'`, slug).Scan(&id); err != nil {
		if err == pgx.ErrNoRows {
			return fail(c, fiber.StatusNotFound, "masterclass not found")
		}
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	sessions, err := sessionsForMasterclass(c.Context(), id)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(sessions)
}

// ---------------------------------------------------------------------------
// Dishes
// ---------------------------------------------------------------------------

func handleDishes(c *fiber.Ctx) error {
	f := &sqlFilter{}
	f.add(`d.status <> 'deleted'`)

	if q := strings.TrimSpace(c.Query("q")); q != "" {
		needle := "%" + q + "%"
		f.add(`(d.name ILIKE ? OR d.description ILIKE ? OR d.category ILIKE ? OR d.region ILIKE ?)`,
			needle, needle, needle, needle)
	}
	if category := strings.TrimSpace(c.Query("category")); category != "" {
		f.add(`d.category ILIKE ?`, "%"+category+"%")
	}
	if region := strings.TrimSpace(c.Query("region")); region != "" {
		f.add(`d.region ILIKE ?`, "%"+region+"%")
	}

	order := ` ORDER BY d.created_at DESC`
	switch c.Query("sort") {
	case "name":
		order = ` ORDER BY d.name ASC`
	case "rating":
		order = ` ORDER BY d.rating DESC, d.created_at DESC`
	case "new":
		order = ` ORDER BY d.created_at DESC`
	}

	limit := queryInt(c, "limit", 12)
	if limit < 1 || limit > 100 {
		limit = 12
	}
	offset := queryInt(c, "offset", 0)
	if offset < 0 {
		offset = 0
	}

	var total int
	if err := db.QueryRow(c.Context(), `SELECT COUNT(*) FROM dishes d`+f.where(), f.args...).Scan(&total); err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}

	query := `SELECT ` + dishSelectCols + ` FROM dishes d` + f.where() + order +
		` LIMIT ` + itoa(limit) + ` OFFSET ` + itoa(offset)
	rows, err := db.Query(c.Context(), query, f.args...)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	defer rows.Close()
	items := make([]DishJSON, 0)
	for rows.Next() {
		d, err := scanDish(rows)
		if err != nil {
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(fiber.Map{"items": items, "total": total})
}

func handleDishDetail(c *fiber.Ctx) error {
	d, err := scanDish(db.QueryRow(c.Context(), `SELECT `+dishSelectCols+` FROM dishes d
		WHERE d.slug=$1 AND d.status <> 'deleted'`, c.Params("slug")))
	if err == pgx.ErrNoRows {
		return fail(c, fiber.StatusNotFound, "dish not found")
	}
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(d)
}

// ---------------------------------------------------------------------------
// Chefs, restaurants, reels, reviews
// ---------------------------------------------------------------------------

func handleChefs(c *fiber.Ctx) error {
	f := &sqlFilter{}
	if city := strings.TrimSpace(c.Query("city")); city != "" {
		f.add(`c.city ILIKE ?`, "%"+city+"%")
	}
	if region := strings.TrimSpace(c.Query("region")); region != "" {
		f.add(`c.city ILIKE ?`, "%"+region+"%")
	}
	rows, err := db.Query(c.Context(), `SELECT `+chefSelectCols+` FROM chefs c
		LEFT JOIN users u ON u.id = c.user_id
		WHERE c.status='approved'`+f.where()+` ORDER BY c.rating DESC, c.created_at DESC`, f.args...)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	defer rows.Close()
	out := make([]ChefJSON, 0)
	for rows.Next() {
		chef, err := scanChef(rows)
		if err != nil {
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		out = append(out, chef)
	}
	return c.JSON(out)
}

func handleRestaurants(c *fiber.Ctx) error {
	rows, err := db.Query(c.Context(), `SELECT `+restaurantSelectCols+` FROM restaurants r
		WHERE r.status='approved' ORDER BY r.rating DESC, r.created_at DESC`)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	defer rows.Close()
	out := make([]RestaurantJSON, 0)
	for rows.Next() {
		r, err := scanRestaurant(rows)
		if err != nil {
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		out = append(out, r)
	}
	return c.JSON(out)
}

func handleRestaurantDetail(c *fiber.Ctx) error {
	slug := c.Params("slug")
	r, err := scanRestaurant(db.QueryRow(c.Context(), `SELECT `+restaurantSelectCols+`
		FROM restaurants r WHERE r.slug=$1 AND r.status='approved'`, slug))
	if err == pgx.ErrNoRows {
		return fail(c, fiber.StatusNotFound, "restaurant not found")
	}
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}

	var id string
	if err := db.QueryRow(c.Context(), `SELECT id::text FROM restaurants WHERE slug=$1`, slug).Scan(&id); err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	f := &sqlFilter{}
	f.add(`m.status = 'approved'`)
	f.add(`m.restaurant_id = ?`, id)
	mcs, _, err := listMasterclasses(c.Context(), f, ` ORDER BY m.rating DESC, m.created_at DESC`, 50, 0)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}

	return c.JSON(fiber.Map{
		"id": r.ID, "name": r.Name, "slug": r.Slug, "city": r.City,
		"cuisine": r.Cuisine, "description": r.Description,
		"image_url": r.ImageURL, "rating": r.Rating, "status": r.Status,
		"masterclasses": mcs,
	})
}

const reelSelect = `
SELECT rl.id::text, rl.title, rl.caption, rl.image_url, rl.video_url,
       rl.views, rl.likes, rl.status, rl.created_at,
       CASE WHEN c.id IS NULL THEN NULL ELSE c.id::text END,
       COALESCE(c.display_name, ''), COALESCE(c.avatar_url, ''),
       CASE WHEN r.id IS NULL THEN NULL ELSE r.id::text END,
       COALESCE(r.name, ''), COALESCE(r.slug, ''), COALESCE(r.image_url, ''),
       CASE WHEN m.id IS NULL THEN NULL ELSE m.id::text END,
       COALESCE(m.title, ''), COALESCE(m.slug, '')
FROM reels rl
LEFT JOIN chefs c ON c.id = rl.chef_id
LEFT JOIN restaurants r ON r.id = rl.restaurant_id
LEFT JOIN masterclasses m ON m.id = rl.masterclass_id`

func scanReel(s scanner) (ReelJSON, error) {
	var reel ReelJSON
	var chefID, restID, mcID sql.NullString
	var chefName, chefAvatar, restName, restSlug, restImage, mcTitle, mcSlug string
	err := s.Scan(&reel.ID, &reel.Title, &reel.Caption, &reel.ImageURL, &reel.VideoURL,
		&reel.Views, &reel.Likes, &reel.Status, &reel.CreatedAt,
		&chefID, &chefName, &chefAvatar,
		&restID, &restName, &restSlug, &restImage,
		&mcID, &mcTitle, &mcSlug)
	if err != nil {
		return reel, err
	}
	if chefID.Valid {
		reel.Chef = map[string]any{"id": chefID.String, "display_name": chefName, "avatar_url": chefAvatar}
	}
	if restID.Valid {
		reel.Restaurant = map[string]any{"id": restID.String, "name": restName, "slug": restSlug, "image_url": restImage}
	}
	if mcID.Valid {
		reel.Masterclass = map[string]any{"id": mcID.String, "title": mcTitle, "slug": mcSlug}
	}
	return reel, nil
}

func handleReels(c *fiber.Ctx) error {
	rows, err := db.Query(c.Context(), reelSelect+` WHERE rl.status='published'
		ORDER BY rl.created_at DESC`)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	defer rows.Close()
	out := make([]ReelJSON, 0)
	for rows.Next() {
		reel, err := scanReel(rows)
		if err != nil {
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		out = append(out, reel)
	}
	return c.JSON(out)
}

func handleReviews(c *fiber.Ctx) error {
	f := &sqlFilter{}
	if mcID := c.Query("masterclass_id"); mcID != "" {
		if !uuidPattern.MatchString(mcID) {
			return fail(c, fiber.StatusBadRequest, "invalid masterclass_id")
		}
		f.add(`rv.masterclass_id = ?`, mcID)
	}
	rows, err := db.Query(c.Context(), `SELECT rv.id::text, rv.masterclass_id::text, rv.user_id::text,
		COALESCE(u.full_name, ''), rv.rating, rv.comment, rv.created_at
		FROM reviews rv LEFT JOIN users u ON u.id = rv.user_id`+f.where()+` ORDER BY rv.created_at DESC`, f.args...)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	defer rows.Close()
	out := make([]fiber.Map, 0)
	for rows.Next() {
		var id, mcID, userID, name, comment string
		var rating int
		var created time.Time
		if err := rows.Scan(&id, &mcID, &userID, &name, &rating, &comment, &created); err != nil {
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		out = append(out, fiber.Map{
			"id": id, "masterclass_id": mcID, "user_id": userID,
			"user_name": name, "rating": rating, "comment": comment, "created_at": created,
		})
	}
	return c.JSON(out)
}
