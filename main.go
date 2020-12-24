package main

import (
	handlers "api/handlers"
	models "api/models"
	"log"
	"os"

	svcs "api/services"

	_ "api/docs" // docs is generated by Swag CLI, you have to import it.

	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	jwtware "github.com/gofiber/jwt/v2"
	"github.com/joho/godotenv"
)

// @title Fiber Example API
// @version 1.0
// @description This is a sample swagger for Fiber
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email fiber@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:3000
// @BasePath /api

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	secret := os.Getenv("SECRET")
	models.ConnectDataBase()
	svcs.InitCloudinary()

	app := fiber.New()
	app.Use(recover.New())
	app.Use(cors.New())

	api := app.Group("/api")

	// set before jwt mid so no auth here
	api.Get("/auth/:id", handlers.ValidateUserEmail)
	api.Post("/login", handlers.Login)
	api.Post("/signup", handlers.Signup)

	// jwt mid auth
	api.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(secret),
	}))

	// set after jwt mid so jwt token required
	api.Get("/users", handlers.FindUsers)
	api.Get("/", handlers.Home)
	api.Get("/search", handlers.SearchItems)
	api.Post("/cart", handlers.AddToCart)
	api.Post("/checkout", handlers.Checkout)
	api.Post("/cancel", handlers.Canceled)
	api.Post("/deliver/:id", handlers.Deliver)
	api.Get("/del/cart/:id", handlers.RemoveFromCart)
	api.Post("/categories", handlers.AddCategory)
	api.Post("/products", handlers.AddProduct)
	api.Post("/items", handlers.AddItem)
	api.Get("/del/categories/:id", handlers.DeleteCategories)
	api.Get("/del/products/:id", handlers.DeleteProducts)
	api.Get("/del/items/:id", handlers.DeleteItems)

	app.Get("/swagger/*", swagger.Handler) // default
	// app.Get("/swagger/*", swagger.New(swagger.Config{ // custom
	// 	URL: "http://example.com/doc.json",
	// 	DeepLinking: false,
	// }))

	log.Fatal(app.Listen(":3000"))
}
