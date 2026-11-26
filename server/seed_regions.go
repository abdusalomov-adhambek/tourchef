package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// getIdempotent region-food seed. geo_id mos keladi app/public/geo/uzbekistan-regions.geojson dagi id ga.
type regionFoodSeed struct {
	id, slug, geoID, name, nameUZ, city          string
	dish, image, description, cultural          string
	ingredients                                 []string
	lat, lon                                    float64
}

func seedRegionFoods(ctx context.Context, tx pgx.Tx, adminID string) error {
	rows := []regionFoodSeed{
		{"cccccccc-cccc-cccc-cccc-cccccccccd01", "tashkent-city", "14", "Tashkent city", "Toshkent shahri", "Tashkent", "Tashkent Plov", imgCauldron, "The capital's signature plov — rice, lamb and spices cooked in a kazan over open fire.", "Plov is the centerpiece of Tashkent hospitality, served at weddings, holidays and family gatherings.", []string{"rice", "lamb", "carrots", "onions", "cumin"}, 41.2810, 69.2647},
		{"cccccccc-cccc-cccc-cccc-cccccccccd02", "namangan", "6", "Namangan region", "Namangan viloyati", "Namangan", "Namangan Manti", imgManti, "Large, juicy steamed manti with lamb and onion, a pride of the Fergana Valley.", "Namangan cooks are famous for folding manti with paper-thin dough and serving them with yogurt.", []string{"flour", "lamb", "onion", "black pepper"}, 41.0810, 71.2988},
		{"cccccccc-cccc-cccc-cccc-cccccccccd03", "tashkent-region", "10", "Tashkent region", "Toshkent viloyati", "Tashkent", "Tashkent Somsa", imgSomsa, "Golden flaky pastry baked in a tandoor, filled with lamb, onion and herbs.", "A everyday street food of the Tashkent region, best eaten fresh from the clay oven.", []string{"flour", "lamb", "onion", "herbs"}, 41.2414, 69.9547},
		{"cccccccc-cccc-cccc-cccc-cccccccccd04", "fergana", "11", "Fergana region", "Fargʻona viloyati", "Fergana", "Fergana Plov", imgPilaf, "A lighter, aromatic plov with chickpeas and fragrant rice from the Fergana Valley.", "Fergana plov traditions reflect the valley's Silk Road exchanges and refined home cooking.", []string{"rice", "chickpeas", "lamb", "cumin"}, 40.3259, 71.2824},
		{"cccccccc-cccc-cccc-cccc-cccccccccd05", "andijan", "1", "Andijan region", "Andijon viloyati", "Andijan", "Andijan Manti", imgManti, "Hand-folded dumplings with a rich lamb filling, steamed and served with sour cream.", "Andijan is renowned for its dumpling masters and generous family tables.", []string{"flour", "lamb", "onion", "yogurt"}, 40.7308, 72.3217},
		{"cccccccc-cccc-cccc-cccc-cccccccccd06", "syrdarya", "9", "Syrdarya region", "Sirdaryo viloyati", "Syrdarya", "Syrdarya Shashlik", imgShashlik, "Charcoal-grilled lamb skewers marinated with onion and spices.", "Along the Syr Darya river, grilled meat is the center of outdoor gatherings.", []string{"lamb", "onion", "salt", "spices"}, 40.5986, 68.6314},
		{"cccccccc-cccc-cccc-cccc-cccccccccd07", "jizzakh", "3", "Jizzakh region", "Jizzax viloyati", "Jizzakh", "Jizzakh Somsa", imgSomsa, "Famous Jizzakh somsa with a generous meat filling and a distinctive local flavor.", "Jizzakh somsa is a point of local pride, sold from tandoors in every market.", []string{"flour", "lamb", "onion", "black cumin"}, 40.3560, 67.8337},
		{"cccccccc-cccc-cccc-cccc-cccccccccd08", "navoi", "5", "Navoi region", "Navoiy viloyati", "Navoi", "Navoi Shurpa", imgShurpa, "A hearty lamb and vegetable soup slowly simmered in a cauldron.", "Shurpa warms households across the desert region and is served with flatbread.", []string{"lamb", "potatoes", "carrots", "tomatoes"}, 41.5994, 64.2370},
		{"cccccccc-cccc-cccc-cccc-cccccccccd09", "samarkand", "7", "Samarkand region", "Samarqand viloyati", "Samarkand", "Samarkand Plov", imgCauldron, "A ceremonial plov with lamb, rice and golden raisins, cooked for large gatherings.", "Samarkand's plov is part of Silk Road hospitality and is often prepared in public kazans.", []string{"rice", "lamb", "raisins", "carrots", "cumin"}, 39.9496, 66.3267},
		{"cccccccc-cccc-cccc-cccc-cccccccccd10", "kashkadarya", "4", "Kashkadarya region", "Qashqadaryo viloyati", "Kashkadarya", "Tandir Kebab", imgShashlik, "Lamb roasted in a clay tandir until tender, with a smoky, rich flavor.", "Tandir kebab is the festive dish of Kashkadarya, prepared for weddings and guests.", []string{"lamb", "salt", "spices"}, 38.7610, 66.0358},
		{"cccccccc-cccc-cccc-cccc-cccccccccd11", "surkhandarya", "8", "Surkhandarya region", "Surxondaryo viloyati", "Surkhandarya", "Surkhandarya Plov", imgPilaf, "A southern-style plov with local rice, lamb and generous spices.", "Surkhandarya shares culinary traditions with neighbouring regions while keeping its own spice blends.", []string{"rice", "lamb", "cumin", "barberries"}, 38.1086, 67.4632},
		{"cccccccc-cccc-cccc-cccc-cccccccccd12", "bukhara", "2", "Bukhara region", "Buxoro viloyati", "Bukhara", "Bukhara Plov", imgNational, "Bukhara-style plov with tender meat, rice and a distinctive blend of spices.", "Bukhara's ancient bazaars and teahouses are the heart of its culinary traditions.", []string{"rice", "lamb", "onion", "cumin"}, 40.2101, 63.7637},
		{"cccccccc-cccc-cccc-cccc-cccccccccd13", "khorezm", "12", "Khorezm region", "Xorazm viloyati", "Khorezm", "Shivit Oshi", imgLagman, "Green noodle soup with herbs, meat and vegetables — a Khorezm specialty.", "Shivit oshi is served in a carved gourde and reflects the region's green, irrigated gardens.", []string{"flour", "dill", "lamb", "potatoes", "carrots"}, 41.2733, 61.2492},
		{"cccccccc-cccc-cccc-cccc-cccccccccd14", "karakalpakstan", "13", "Republic of Karakalpakstan", "Qoraqalpogʻiston Respublikasi", "Karakalpakstan", "Karakalpak Plov", imgSumalak, "A hearty rice dish with meat and local spices from the Aral region.", "Karakalpak cuisine blends nomadic traditions with the flavors of the Amudarya delta.", []string{"rice", "lamb", "carrots", "spices"}, 43.2856, 59.2410},
	}
	for _, r := range rows {
		ing, err := json.Marshal(r.ingredients)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO region_foods(id, slug, geo_id, name, name_uz, city, dish_name, dish_image_url,
				description, ingredients, cultural_info, latitude, longitude, status, created_by)
			VALUES($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb, $11, $12, $13, 'published', $14::uuid)
			ON CONFLICT (id) DO NOTHING`,
			r.id, r.slug, r.geoID, r.name, r.nameUZ, r.city, r.dish, r.image, r.description,
			string(ing), r.cultural, r.lat, r.lon, adminID); err != nil {
			return fmt.Errorf("seed region food %s: %w", r.slug, err)
		}
	}
	return nil
}
