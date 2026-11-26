package main

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func newApp() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:               "TourChef API",
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			message := "internal server error"
			var fe *fiber.Error
			if errors.As(err, &fe) {
				code = fe.Code
				message = fe.Message
			}
			return c.Status(code).JSON(fiber.Map{"error": message})
		},
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigin,
		AllowMethods:     "GET, POST, PATCH, PUT, DELETE, OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: false,
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Static("/uploads", cfg.UploadDir, fiber.Static{
		MaxAge:       3600,
		Browse:       false,
		Index:        "",
		Compress:     false,
	})

	api := app.Group("/api")

	// Auth (public)
	api.Post("/auth/register", handleRegister)
	api.Post("/auth/login", handleLogin)
	api.Post("/auth/forgot-password", handleForgotPassword)
	api.Post("/upload", requireAuth, requireRole("admin", "chef"), handleUpload)
	api.Get("/auth/me", requireAuth, handleMe)
	api.Patch("/auth/me", requireAuth, handleUpdateMe)

	// Public catalog
	api.Get("/chefs", handleChefs)
	api.Get("/families", handleChefs) // chef -> family (UI nomi)
	api.Get("/dishes", handleDishes)
	api.Get("/dishes/:slug", handleDishDetail)
	api.Get("/masterclasses", handleMasterclasses)
	api.Get("/masterclasses/:slug/sessions", handleMasterclassSessions)
	api.Get("/masterclasses/:slug", handleMasterclassDetail)
	api.Get("/restaurants", handleRestaurants)
	api.Get("/restaurants/:slug", handleRestaurantDetail)
	api.Get("/reels", handleReels)
	api.Get("/reviews", handleReviews)

	// Taste of Uzbekistan — interaktiv mintaqa-taom xaritasi
	api.Get("/v1/regions", handleRegions)
	api.Get("/v1/regions/:slug", handleRegionDetail)
	api.Get("/v1/regions/:slug/featured-dish", handleRegionDish)
	api.Get("/v1/regions/:slug/masterclasses", handleRegionMasterclasses)
	api.Get("/v1/regions/:slug/restaurants", handleRegionRestaurants)
	// qisqa alias (frontend /api/regions ni ham ishlata oladi)
	api.Get("/regions", handleRegions)
	api.Get("/regions/:slug", handleRegionDetail)

	// Dishes — admin writes
	api.Post("/dishes", requireAuth, requireRole("admin"), handleCreateDish)
	api.Patch("/dishes/:id", requireAuth, requireRole("admin"), handleUpdateDish)
	api.Delete("/dishes/:id", requireAuth, requireRole("admin"), handleDeleteDish)

	// Masterclasses — chef or admin create
	api.Post("/masterclasses", requireAuth, requireRole("chef", "admin"), handleCreateMasterclass)

	// Chef self-service
	chef := api.Group("/chef", requireAuth, requireRole("chef", "admin"))
	chef.Get("/masterclasses", handleChefMasterclasses)
	chef.Post("/sessions", handleCreateSession)
	chef.Get("/sessions", handleChefSessions)

	// Family Cabinet — xuddi shu funksiyalar, yangi nom bilan.
	family := api.Group("/family", requireAuth, requireRole("chef", "admin"))
	family.Get("/masterclasses", handleChefMasterclasses)
	family.Post("/sessions", handleCreateSession)
	family.Get("/sessions", handleChefSessions)

	// Client cabinet
	client := api.Group("/client", requireAuth, requireRole("client"))
	client.Post("/bookings", handleCreateBooking)
	client.Get("/bookings", handleClientBookings)
	client.Get("/bookings/:id/qr", handleBookingQR)
	client.Patch("/bookings/:id/cancel", handleCancelBooking)
	client.Get("/favorites", handleClientFavorites)
	client.Post("/favorites", handleAddFavorite)
	client.Delete("/favorites/:id", handleDeleteFavorite)
	client.Post("/reviews", handleCreateReview)
	client.Get("/notifications", handleClientNotifications)

	// Bookings — client only (public POST is intentionally protected)
	api.Post("/bookings", requireAuth, requireRole("client"), handleCreateBooking)

	// Admin
	admin := api.Group("/admin", requireAuth, requireRole("admin"))
	admin.Get("/chefs", handleAdminChefs)
	admin.Patch("/chefs/:id", handleAdminUpdateChef)
	admin.Get("/families", handleAdminChefs)
	admin.Patch("/families/:id", handleAdminUpdateChef)
	admin.Get("/users", handleAdminUsers)
	admin.Patch("/users/:id", handleAdminUpdateUser)
	admin.Get("/masterclasses", handleAdminMasterclasses)
	admin.Get("/masterclasses/:slug", handleAdminMasterclassDetail)
	admin.Patch("/masterclasses/:id", handleAdminUpdateMasterclass)
	admin.Get("/sessions", handleAdminSessions)
	admin.Get("/bookings", handleAdminBookings)
	admin.Patch("/bookings/:id", handleAdminUpdateBooking)
	admin.Get("/stats", handleAdminStats)
	admin.Get("/reels", handleAdminReels)
	admin.Patch("/reels/:id", handleAdminUpdateReel)
	admin.Get("/regions", handleAdminRegions)
	admin.Post("/regions", handleAdminCreateRegion)
	admin.Patch("/regions/:id", handleAdminUpdateRegion)
	admin.Delete("/regions/:id", handleAdminDeleteRegion)

	app.Use(func(c *fiber.Ctx) error {
		return fail(c, fiber.StatusNotFound, "endpoint not found")
	})
	return app
}
