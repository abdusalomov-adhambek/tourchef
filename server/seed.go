package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Real Wikimedia Commons image URLs reused from the project's design files.
const (
	imgPilaf              = "https://thumb.wikimedia.org/wikipedia/commons/thumb/4/4c/Uzbek_pilaf.jpg/960px-Uzbek_pilaf.jpg"
	imgManti              = "https://thumb.wikimedia.org/wikipedia/commons/thumb/6/6a/Uzbek_manti.JPG/960px-Uzbek_manti.JPG"
	imgSomsa              = "https://thumb.wikimedia.org/wikipedia/commons/thumb/1/14/Uzbek_national_Somsa.jpg/960px-Uzbek_national_Somsa.jpg"
	imgLagman             = "https://thumb.wikimedia.org/wikipedia/commons/thumb/d/dc/Lagman_Moscow_2023.jpg/960px-Lagman_Moscow_2023.jpg"
	imgShashlik           = "https://thumb.wikimedia.org/wikipedia/commons/thumb/9/90/Shashlik-national_food_of_Uzbekistan.jpg/960px-Shashlik-national_food_of_Uzbekistan.jpg"
	imgChuchvara          = "https://thumb.wikimedia.org/wikipedia/commons/thumb/9/9d/Chuchvara4.jpg/960px-Chuchvara4.jpg"
	imgShurpa             = "https://thumb.wikimedia.org/wikipedia/commons/thumb/e/e2/%D0%A1%D0%B0%D0%BC%D0%B0%D1%80%D0%BA%D0%B0%D0%BD%D0%B4%2C_%D1%81%D0%BE%D1%80%D0%BF%D0%B0_%D0%B2_%D1%83%D0%BB%D0%B8%D1%87%D0%BD%D0%BE%D0%BC_%D0%BA%D0%B0%D1%84%D0%B5.jpg/960px-%D0%A1%D0%B0%D0%BC%D0%B0%D1%80%D0%BA%D0%B0%D0%BD%D0%B4%2C_%D1%81%D0%BE%D1%80%D0%BF%D0%B0_%D0%B2_%D1%83%D0%BB%D0%B8%D1%87%D0%BD%D0%BE%D0%BC_%D0%BA%D0%B0%D1%84%D0%B5.jpg"
	imgNational           = "https://thumb.wikimedia.org/wikipedia/commons/thumb/0/04/Uzbek_national_cuisine.jpg/960px-Uzbek_national_cuisine.jpg"
	imgSumalak            = "https://thumb.wikimedia.org/wikipedia/commons/thumb/a/ac/Uzbek_national_dastarkhan_and_wheat_prepared_for_sumalak.jpg/960px-Uzbek_national_dastarkhan_and_wheat_prepared_for_sumalak.jpg"
	imgCauldron           = "https://thumb.wikimedia.org/wikipedia/commons/thumb/0/0a/Traditional_Uzbek_Pilaf_in_a_Cauldron_on_a_Potbelly_Stove_Moscow_Region.jpg/960px-Traditional_Uzbek_Pilaf_in_a_Cauldron_on_a_Potbelly_Stove_Moscow_Region.jpg"
	imgUnsplashRestaurant = "https://images.unsplash.com/photo-1517248135467-4c7edcad34c4?auto=format&fit=crop&w=800&q=80"
	imgUnsplashInterior   = "https://images.unsplash.com/photo-1414235077428-338989a2e8c0?auto=format&fit=crop&w=800&q=80"
	imgUnsplashCafe       = "https://images.unsplash.com/photo-1552566626-52f8b828add9?auto=format&fit=crop&w=800&q=80"
	devVideoURL           = "https://www.w3schools.com/html/mov_bbb.mp4"
)

// Fixed UUIDs make the whole seed idempotent: every INSERT uses ON CONFLICT DO NOTHING.
const (
	seedAdminID      = "11111111-1111-1111-1111-111111111111"
	seedChefUserID   = "22222222-2222-2222-2222-222222222222"
	seedChefID       = "22222222-2222-2222-2222-222222222223"
	seedClientID     = "33333333-3333-3333-3333-333333333333"
	seedPendingUser1 = "44444444-4444-4444-4444-444444444441"
	seedPendingChef1 = "44444444-4444-4444-4444-444444444442"
	seedPendingUser2 = "44444444-4444-4444-4444-444444444443"
	seedPendingChef2 = "44444444-4444-4444-4444-444444444444"

	seedRest1 = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1"
	seedRest2 = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2"
	seedRest3 = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3"

	seedDish1 = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb1"
	seedDish2 = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb2"
	seedDish3 = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb3"
	seedDish4 = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb4"
	seedDish5 = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb5"
	seedDish6 = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb6"
	seedDish7 = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb7"
	seedDish8 = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb8"

	seedMC1 = "cccccccc-cccc-cccc-cccc-ccccccccccc1"
	seedMC2 = "cccccccc-cccc-cccc-cccc-ccccccccccc2"
	seedMC3 = "cccccccc-cccc-cccc-cccc-ccccccccccc3"
	seedMC4 = "cccccccc-cccc-cccc-cccc-ccccccccccc4"
	seedMC5 = "cccccccc-cccc-cccc-cccc-ccccccccccc5"
)

func seedIfEnabled(ctx context.Context, pool *pgxpool.Pool, enabled bool) error {
	if !enabled {
		log.Println("seed disabled (SEED=false)")
		return nil
	}
	var n int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE email='admin@tourchef.uz'`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		// Asosiy seed mavjud bo'lsa ham, mintaqa katalogini har doim to'ldirib turamiz.
		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback(ctx) }()
		if err := seedRegionFoods(ctx, tx, seedAdminID); err != nil {
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		log.Println("seed data already present, region foods refreshed")
		return nil
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := insertSeedUsers(ctx, tx); err != nil {
		return err
	}
	if err := insertSeedContent(ctx, tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	log.Println("seed data inserted (admin@ / chef@ / client@tourchef.uz)")
	return nil
}

type seedUser struct {
	id, email, password, name, role string
}

func insertSeedUsers(ctx context.Context, tx pgx.Tx) error {
	users := []seedUser{
		{seedAdminID, "admin@tourchef.uz", "admin123", "TourChef Admin", "admin"},
		{seedChefUserID, "family@tourchef.uz", "family123", "Aziz Karimov", "chef"},
		{seedClientID, "client@tourchef.uz", "client123", "Emily Johnson", "client"},
		{seedPendingUser1, "pending1@tourchef.uz", "family123", "Malika Yusupova", "chef"},
		{seedPendingUser2, "pending2@tourchef.uz", "family123", "Rustam Ergashev", "chef"},
	}
	for _, u := range users {
		hash, err := hashPassword(u.password)
		if err != nil {
			return fmt.Errorf("hash password for %s: %w", u.email, err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO users(id, email, password_hash, full_name, role, status)
			VALUES($1::uuid, $2, $3, $4, $5, 'active')
			ON CONFLICT (id) DO NOTHING`, u.id, u.email, hash, u.name, u.role); err != nil {
			return fmt.Errorf("seed user %s: %w", u.email, err)
		}
	}

	chefs := []struct{ id, userID, name, city, specialty, bio, avatar, status string }{
		{seedChefID, seedChefUserID, "Aziz Karimov", "Tashkent", "Plov & traditional Uzbek cuisine",
			"Master chef with 15 years of experience cooking plov in a traditional kazan.", imgPilaf, "approved"},
		{seedPendingChef1, seedPendingUser1, "Malika Yusupova", "Samarkand", "Manti & dumplings",
			"Learned hand-folded manti from her grandmother and shares the tradition with travelers.", imgManti, "pending"},
		{seedPendingChef2, seedPendingUser2, "Rustam Ergashev", "Bukhara", "Tandoor baking",
			"Specializes in somsa and bread baked in a clay tandoor.", imgSomsa, "pending"},
	}
	for _, ch := range chefs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO chefs(id, user_id, display_name, city, specialty, bio, avatar_url, status)
			VALUES($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO NOTHING`, ch.id, ch.userID, ch.name, ch.city, ch.specialty,
			ch.bio, ch.avatar, ch.status); err != nil {
			return fmt.Errorf("seed chef %s: %w", ch.name, err)
		}
	}
	return nil
}

func insertSeedContent(ctx context.Context, tx pgx.Tx) error {
	restaurants := []struct {
		id, name, slug, city, cuisine, description, image string
		rating                                            float64
	}{
		{seedRest1, "Osh Markazi", "osh-markazi", "Tashkent", "Uzbek",
			"The central plov house of Tashkent, where kazan plov is cooked over an open fire every morning.", imgCauldron, 4.9},
		{seedRest2, "Samarkand Kitchen", "samarkand-kitchen", "Samarkand", "Silk Road cuisine",
			"A family restaurant serving recipes collected along the ancient Silk Road.", imgUnsplashRestaurant, 4.8},
		{seedRest3, "Bukhara Teahouse", "bukhara-teahouse", "Bukhara", "Traditional",
			"A courtyard teahouse famous for tandoor-baked somsa and slow-cooked shurpa.", imgUnsplashCafe, 4.7},
	}
	for _, r := range restaurants {
		if _, err := tx.Exec(ctx, `
			INSERT INTO restaurants(id, name, slug, city, cuisine, description, image_url, rating, status)
			VALUES($1::uuid, $2, $3, $4, $5, $6, $7, $8, 'approved')
			ON CONFLICT (id) DO NOTHING`, r.id, r.name, r.slug, r.city, r.cuisine,
			r.description, r.image, r.rating); err != nil {
			return fmt.Errorf("seed restaurant %s: %w", r.name, err)
		}
	}

	dishes := []struct {
		id, name, slug, category, region, description, ingredients, image string
		rating                                                            float64
	}{
		{seedDish1, "Plov", "plov", "Main course", "Tashkent",
			"Rice, meat and carrots slow-cooked in a cauldron — the heart of Uzbek hospitality.",
			"rice, lamb, carrots, onions, cumin, black pepper", imgPilaf, 4.9},
		{seedDish2, "Manti", "manti", "Dumplings", "Samarkand",
			"Steamed dumplings with spiced lamb, served with yogurt or melted butter.",
			"flour, lamb, onion, black pepper, yogurt", imgManti, 4.8},
		{seedDish3, "Somsa", "somsa", "Bakery", "Bukhara",
			"Golden, flaky pastry pockets baked in a tandoor, filled with meat and onions.",
			"flour, lamb, onion, sesame, butter", imgSomsa, 4.7},
		{seedDish4, "Lagman", "lagman", "Noodles", "Tashkent",
			"Hand-pulled noodles in a rich, savory broth with vegetables and meat.",
			"flour, beef, tomato, peppers, herbs", imgLagman, 4.6},
		{seedDish5, "Shashlik", "shashlik", "Grill", "Tashkent",
			"Marinated skewers grilled over charcoal until smoky and tender.",
			"lamb, onion, vinegar, spices", imgShashlik, 4.8},
		{seedDish6, "Chuchvara", "chuchvara", "Dumplings", "Fergana",
			"Tiny meat dumplings served in broth with vinegar and black pepper.",
			"flour, minced meat, broth, vinegar", imgChuchvara, 4.5},
		{seedDish7, "Shurpa", "shurpa", "Soup", "Bukhara",
			"A hearty lamb and vegetable soup, slow-simmered in a clay pot.",
			"lamb, potatoes, carrots, tomato, herbs", imgShurpa, 4.6},
		{seedDish8, "Sumalak", "sumalak", "Festive", "Karakalpakstan",
			"Sweet wheat sprout pudding cooked overnight for Navruz — a symbol of spring.",
			"wheat sprouts, flour, oil, sugar", imgSumalak, 4.9},
	}
	for _, d := range dishes {
		if _, err := tx.Exec(ctx, `
			INSERT INTO dishes(id, name, slug, category, region, description, ingredients,
				image_url, rating, status, created_by)
			VALUES($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, 'active', $10::uuid)
			ON CONFLICT (id) DO NOTHING`, d.id, d.name, d.slug, d.category, d.region,
			d.description, d.ingredients, d.image, d.rating, seedAdminID); err != nil {
			return fmt.Errorf("seed dish %s: %w", d.name, err)
		}
	}

	mcs := []struct {
		id, title, slug, description, dishID, restID, city, difficulty, language, image string
		duration                                                                        int
		price, rating                                                                   float64
	}{
		{seedMC1, "The Art of Uzbek Plov", "the-art-of-uzbek-plov",
			"Learn the secrets of a perfect kazan plov from a master chef: rice, meat, spice balance and open-fire technique.",
			seedDish1, seedRest1, "Tashkent", "beginner", "en", imgPilaf, 180, 180000, 4.9},
		{seedMC2, "Hand-folded Manti", "hand-folded-manti",
			"Fold delicate manti by hand and discover the Samarkand family recipes behind the perfect filling.",
			seedDish2, seedRest2, "Samarkand", "beginner", "en", imgManti, 150, 150000, 4.8},
		{seedMC3, "Somsa from the Tandoor", "somsa-from-the-tandoor",
			"Shape, fill and bake traditional somsa in a real clay tandoor with a Bukhara master.",
			seedDish3, seedRest3, "Bukhara", "intermediate", "uz", imgSomsa, 120, 120000, 4.7},
		{seedMC4, "Shashlik over Charcoal", "shashlik-over-charcoal",
			"Master the marinade, the skewer and the charcoal heat that make Uzbek shashlik unforgettable.",
			seedDish5, seedRest1, "Tashkent", "beginner", "en", imgShashlik, 150, 210000, 4.8},
		{seedMC5, "Lagman & Chuchvara Workshop", "lagman-and-chuchvara-workshop",
			"Pull noodles by hand and shape tiny chuchvara, then taste both in rich homemade broth.",
			seedDish4, seedRest2, "Samarkand", "intermediate", "en", imgLagman, 240, 165000, 4.6},
	}
	for _, m := range mcs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO masterclasses(id, title, slug, description, dish_id, chef_id, restaurant_id,
				city, duration_min, price, difficulty, language, rating, image_url, status)
			VALUES($1::uuid, $2, $3, $4, $5::uuid, $6::uuid, $7::uuid, $8, $9, $10,
				$11, $12, $13, $14, 'approved')
			ON CONFLICT (id) DO NOTHING`, m.id, m.title, m.slug, m.description, m.dishID,
			seedChefID, m.restID, m.city, m.duration, m.price, m.difficulty,
			m.language, m.rating, m.image); err != nil {
			return fmt.Errorf("seed masterclass %s: %w", m.title, err)
		}
	}

	now := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Hour)
	mcIDs := []string{seedMC1, seedMC2, seedMC3, seedMC4, seedMC5}
	prices := []float64{180000, 150000, 120000, 210000, 165000}
	for i, mcID := range mcIDs {
		for j := 0; j < 2; j++ {
			sessionID := fmt.Sprintf("dddddddd-dddd-dddd-dddd-dddddddddd%02d", i*2+j+1)
			startsAt := now.Add(time.Duration(i*24+j*72) * time.Hour)
			if _, err := tx.Exec(ctx, `
				INSERT INTO sessions(id, masterclass_id, starts_at, capacity, booked, price, status)
				VALUES($1::uuid, $2::uuid, $3, $4, $5, $6, 'scheduled')
				ON CONFLICT (id) DO NOTHING`,
				sessionID, mcID, startsAt, 10, j, prices[i]); err != nil {
				return fmt.Errorf("seed session: %w", err)
			}
		}
	}

	reels := []struct {
		id, title, caption, image, video, restID, mcID string
		views, likes                                   int
	}{
		{"eeeeeeee-eeee-eeee-eeee-eeeeeeeeeee1", "Plov masterclass in Tashkent",
			"A morning at the plov house: the kazan, the fire, the first spoonful.", imgCauldron,
			devVideoURL, seedRest1, seedMC1, 128000, 9400},
		{"eeeeeeee-eeee-eeee-eeee-eeeeeeeeeee2", "Shashlik over charcoal",
			"That first sizzle is the sound of Uzbek hospitality.", imgShashlik,
			devVideoURL, seedRest1, seedMC4, 92000, 6100},
		{"eeeeeeee-eeee-eeee-eeee-eeeeeeeeeee3", "Somsa from the tandoor",
			"Flaky, golden, straight from the clay oven in Bukhara.", imgSomsa,
			devVideoURL, seedRest3, seedMC3, 64000, 4800},
	}
	for _, r := range reels {
		if _, err := tx.Exec(ctx, `
			INSERT INTO reels(id, title, caption, image_url, video_url, views, likes,
				chef_id, restaurant_id, masterclass_id, status)
			VALUES($1::uuid, $2, $3, $4, $5, $6, $7, $8::uuid, $9::uuid, $10::uuid, 'published')
			ON CONFLICT (id) DO NOTHING`, r.id, r.title, r.caption, r.image, r.video,
			r.views, r.likes, seedChefID, r.restID, r.mcID); err != nil {
			return fmt.Errorf("seed reel: %w", err)
		}
	}

	// Sample engagement for the client account.
	if _, err := tx.Exec(ctx, `
		INSERT INTO favorites(id, user_id, masterclass_id)
		VALUES('ffffffff-ffff-ffff-ffff-fffffffffff1'::uuid, $1::uuid, $2::uuid),
		       ('ffffffff-ffff-ffff-ffff-fffffffffff2'::uuid, $1::uuid, $3::uuid)
		ON CONFLICT (id) DO NOTHING`, seedClientID, seedMC1, seedMC2); err != nil {
		return fmt.Errorf("seed favorites: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO reviews(id, user_id, masterclass_id, rating, comment)
		VALUES('99999999-9999-9999-9999-999999999991'::uuid, $1::uuid, $2::uuid, 5,
				'Cooking plov with a real master was the highlight of our trip.'),
		       ('99999999-9999-9999-9999-999999999992'::uuid, $1::uuid, $3::uuid, 4,
				'So much fun folding manti — and they were delicious!')
		ON CONFLICT (id) DO NOTHING`, seedClientID, seedMC1, seedMC2); err != nil {
		return fmt.Errorf("seed reviews: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO notifications(id, user_id, title, body, type)
		VALUES('77777777-7777-7777-7777-777777777771'::uuid, $1::uuid,
			'Welcome to TourChef', 'Discover authentic Uzbek cooking experiences.', 'info'),
		       ('77777777-7777-7777-7777-777777777772'::uuid, $1::uuid,
			'Booking confirmed', 'Your plov masterclass booking is confirmed.', 'booking')
		ON CONFLICT (id) DO NOTHING`, seedClientID); err != nil {
		return fmt.Errorf("seed notifications: %w", err)
	}
	if err := seedRegionFoods(ctx, tx, seedAdminID); err != nil {
		return fmt.Errorf("seed region foods: %w", err)
	}
	return nil
}
