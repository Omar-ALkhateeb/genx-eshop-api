package handlers

import (
	models "api/models"
	svcs "api/services"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	jwt "github.com/form3tech-oss/jwt-go"
	"github.com/gofiber/fiber/v2"
)

// FindUsers all
// @Summary find users
// @Accept json
// @Produce json
// @Router /users [get]
func FindUsers(c *fiber.Ctx) error {

	_, _, isadmin := parseJWT(c)

	if isadmin {
		var users []models.User
		if err := models.DB.Set("gorm:auto_preload", true).Find(&users).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "no records found!"})
		}

		return c.JSON(fiber.Map{"data": users})
	}
	return c.Status(304).JSON(fiber.Map{"status": "not authorized!"})

}

// Login func (generates jwt that has to be placed in Header Authorization: Bearer token)
// @Summary login
// @Accept json
// @Produce json
// @Router /login [post]
func Login(c *fiber.Ctx) error {
	var input = new(models.User)
	var user models.User
	if err := c.BodyParser(input); err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "wrong data!"})
	}

	log.Println(input) // log user data

	// find the user
	if err := models.DB.First(&user, "Username = ? AND Password = ?", input.Username, input.Password).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "no such user"})
	}

	// check if validated
	if !user.Validated {
		return c.Status(304).JSON(fiber.Map{"error": "email not validated"})
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = user.Username
	claims["admin"] = user.Admin
	claims["id"] = user.ID
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	secret := os.Getenv("SECRET")

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"token": t})
}

// Signup (no validation!!!)
// @Summary signup
// @Accept json
// @Produce json
// @Router /signup [post]
func Signup(c *fiber.Ctx) error {
	// Validate input
	var input = new(models.User)
	if err := c.BodyParser(input); err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "wrong data!"})
	}

	log.Println(input) // log user data

	// Create a user
	user := models.User{Username: input.Username, Password: input.Password, Phoneno: input.Phoneno, Email: input.Email}
	if err := models.DB.Create(&user); err.Error != nil {
		fmt.Println(err.Error)
		return c.Status(401).JSON(fiber.Map{"error": "email already used!"})
	}

	// uncomment on prod

	if err := svcs.SendMail(user.Email, user.ID); err != nil {
		fmt.Println(err.Error())
	}

	return c.Status(201).JSON(fiber.Map{"status": "created!"})

}

// ValidateUserEmail using id through email
// @Summary validate email
// @Param id path string true "user id"
// @Produce json
// @Router /auth/{id} [get]
func ValidateUserEmail(c *fiber.Ctx) error {
	var user models.User
	//models.DB.Set("gorm:auto_preload", true).Find(&user, "id = ?", c.Param("id"))
	if err := models.DB.Where("id = ?", c.Params("id")).First(&user).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "no such user"})
	}

	models.DB.Model(&user).Update("Validated", true)

	return c.Status(201).JSON(fiber.Map{"status": "created!"})

}

// helper function to retrieve data from jwt (admin and name)
func parseJWT(c *fiber.Ctx) (name string, id float64, admin bool) {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	name = claims["name"].(string)
	id = claims["id"].(float64)
	admin = claims["admin"].(bool)
	return name, id, admin
}

// AddProduct through form
// @Summary add products
// @Accept multipart/form-data
// @Produce json
// @Router /products [post]
func AddProduct(c *fiber.Ctx) error {
	_, _, isadmin := parseJWT(c)

	if isadmin {
		name := c.FormValue("name")
		desc := c.FormValue("description")
		catg := c.FormValue("category")

		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "no such file!"})
		}
		url := svcs.UploadFile(file)

		// find our category
		var category models.Category
		//models.DB.Set("gorm:auto_preload", true).Find(&user, "id = ?", c.Param("id"))
		if err := models.DB.Where("id = ?", catg).First(&category).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "no such catg!"})
		}

		// Create a prod
		prod := models.Product{Name: name, Description: desc, Category: category, CategoryID: category.ID, ImgURL: url}
		if err := models.DB.Create(&prod); err.Error != nil {
			fmt.Println(err.Error)
			return c.Status(401).JSON(fiber.Map{"error": "invalid data!"})
		}

		return c.Status(201).JSON(fiber.Map{"status": "created!"})
	}

	return c.Status(304).JSON(fiber.Map{"status": "not authorized!"})

}

// AddItem func to add  items from product ids
// @Summary add items
// @Accept json
// @Produce json
// @Router /items [post]
func AddItem(c *fiber.Ctx) error {
	_, _, isadmin := parseJWT(c)

	if isadmin {
		// Validate input
		type inputstruct struct {
			models.Item
			PID int
		}
		var input = new(inputstruct)
		if err := c.BodyParser(input); err != nil {
			// log.Panicln(err.Error())
			return c.Status(403).JSON(fiber.Map{"error": "wrong data!"})
		}

		log.Println(input.PID) // log item data

		// find our product
		var product models.Product
		//models.DB.Set("gorm:auto_preload", true).Find(&user, "id = ?", c.Param("id"))
		if err := models.DB.Where("id = ?", input.PID).First(&product).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "no such prod!"})
		}

		// Create an item
		item := models.Item{Product: product, Qty: input.Qty, PricePerItem: input.PricePerItem}
		if err := models.DB.Create(&item); err.Error != nil {
			fmt.Println(err.Error)
			return c.Status(401).JSON(fiber.Map{"error": "invalid data!"})
		}

		return c.Status(201).JSON(fiber.Map{"status": "created!"})
	}

	return c.Status(304).JSON(fiber.Map{"status": "not authorized!"})

}

// AddCategory func
// @Summary add categories
// @Accept json
// @Produce json
// @Router /categories [post]
func AddCategory(c *fiber.Ctx) error {

	_, _, isadmin := parseJWT(c)

	if isadmin {
		// Validate input
		var input = new(models.Category)
		if err := c.BodyParser(input); err != nil {
			return c.Status(403).JSON(fiber.Map{"error": "wrong data!"})
		}

		log.Println(input) // log user data

		// Create a user
		catg := models.Category{Name: input.Name, Description: input.Description}
		if err := models.DB.Create(&catg); err.Error != nil {
			fmt.Println(err.Error)
			return c.Status(401).JSON(fiber.Map{"error": "wrong data!"})
		}

		return c.Status(201).JSON(fiber.Map{"status": "created!"})
	}
	return c.Status(304).JSON(fiber.Map{"status": "not authorized!"})

}

// SearchItems by name
// @Summary search
// @Param page query int true "page num"
// @Param term query string true "search term"
// @Produce json
// @Router /search [get]
func SearchItems(c *fiber.Ctx) error {
	var items []models.Item
	var page = 1
	var pageSize = 5
	if c.Query("page") != "" {
		page, _ = strconv.Atoi(c.Query("page"))
	}
	var offset = (page - 1) * pageSize
	if err := models.DB.Set("gorm:auto_preload", true).Limit(pageSize).Offset(offset).Where("Description @@ to_tsquery(?)", c.Query("term")).Find(&items).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found!"})
	}

	return c.JSON(fiber.Map{"items": items})
}

// DeleteProducts func
// @Summary del products
// @Param id path string true "prod id"
// @Produce json
// @Router /del/products/{id} [get]
func DeleteProducts(c *fiber.Ctx) error {
	_, _, isadmin := parseJWT(c)

	if isadmin {
		// Get model if exist
		var product models.Product
		if err := models.DB.Where("id = ?", c.Params("id")).First(&product).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "not found!"})
		}

		models.DB.Delete(&product)

		return c.Status(201).JSON(fiber.Map{"status": "deleted!"})
	}

	return c.Status(304).JSON(fiber.Map{"status": "not authorized!"})
}

// DeleteItems func
// @Summary del items
// @Param id path string true "item id"
// @Produce json
// @Router /del/items/{id} [get]
func DeleteItems(c *fiber.Ctx) error {
	_, _, isadmin := parseJWT(c)

	if isadmin {
		// Get model if exist
		var item models.Item
		if err := models.DB.Where("id = ?", c.Params("id")).First(&item).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "not found!"})
		}

		models.DB.Delete(&item)

		return c.Status(201).JSON(fiber.Map{"status": "deleted!"})
	}

	return c.Status(304).JSON(fiber.Map{"status": "not authorized!"})
}

// DeleteCategories func
// @Summary del categories
// @Param id path string true "catg id""
// @Produce json
// @Router /del/categories/{id} [get]
func DeleteCategories(c *fiber.Ctx) error {
	_, _, isadmin := parseJWT(c)

	if isadmin {
		// Get model if exist
		var category models.Category
		if err := models.DB.Where("id = ?", c.Params("id")).First(&category).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "not found!"})
		}

		models.DB.Delete(&category)

		return c.Status(201).JSON(fiber.Map{"status": "deleted!"})
	}

	return c.Status(304).JSON(fiber.Map{"status": "not authorized!"})
}

// Home func
// @Summary homepage
// @Param page query int true "page num"
// @Produce json
// @Router / [get]
func Home(c *fiber.Ctx) error {
	var items []models.Item
	var page = 1
	var pageSize = 5
	if c.Query("page") != "" {
		page, _ = strconv.Atoi(c.Query("page"))
	}
	var offset = (page - 1) * pageSize
	if err := models.DB.Set("gorm:auto_preload", true).Limit(pageSize).Offset(offset).Find(&items).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found!"})
	}

	return c.JSON(fiber.Map{"items": items})
}

// GetCart func
// @Summary get cart
// @Produce json
// @Router /cart [get]
func GetCart(c *fiber.Ctx) error {
	var cart models.Order
	_, id, _ := parseJWT(c)
	if err := models.DB.Set("gorm:auto_preload", true).Where("Delivered = ? AND Canceled = ? AND Confirmed = ? AND id = ?", false, false, false, uint(id)).First(&cart).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found!"})
	}

	return c.JSON(fiber.Map{"cart": cart})
}

// AddToCart func
// @Summary add to cart
// @Accept json
// @Produce json
// @Router /cart [post]
func AddToCart(c *fiber.Ctx) error {
	var cart models.Order
	// create orderItem
	type inputstruct struct {
		models.OrderItem
		IID int // item id
	}
	var input = new(inputstruct)
	if err := c.BodyParser(input); err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "wrong data!"})
	}

	log.Println(input.IID) // log item data

	_, id, _ := parseJWT(c)
	if err := models.DB.Set("gorm:auto_preload", true).Where("Delivered = ? AND Canceled = ? AND Confirmed = ? AND id ?", false, false, false, uint(id)).First(&cart).Error; err != nil {
		// if no cart is open create one
		order := models.Order{UserID: uint(id)}
		if err := models.DB.Create(&order); err.Error != nil {
			fmt.Println(err.Error)
			return c.Status(401).JSON(fiber.Map{"error": "wrong data!"})
		}

		// find our item
		var item models.Item
		if err := models.DB.Where("id = ?", input.IID).First(&item).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "no such item!"})
		}

		// Create an order item
		orderitem := models.OrderItem{Item: item, Qty: input.Qty, OrderID: order.ID, ItemID: item.ID}
		if err := models.DB.Create(&orderitem); err.Error != nil {
			fmt.Println(err.Error)
			return c.Status(401).JSON(fiber.Map{"error": "invalid data!"})
		}

		return c.Status(201).JSON(fiber.Map{"status": "created!"})

	}
	// else

	// find our item
	var item models.Item
	if err := models.DB.Where("id = ?", input.IID).First(&item).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "no such item!"})
	}

	// Create an order item
	orderitem := models.OrderItem{Item: item, Qty: input.Qty, OrderID: cart.ID, ItemID: item.ID}
	if err := models.DB.Create(&orderitem); err.Error != nil {
		fmt.Println(err.Error)
		return c.Status(401).JSON(fiber.Map{"error": "invalid data!"})
	}

	return c.Status(201).JSON(fiber.Map{"status": "created!"})

}

// RemoveFromCart func
// @Summary remove from cart
// @Accept json
// @Produce json
// @Param id path string true "order item id"
// @Router /del/cart/{id} [get]
func RemoveFromCart(c *fiber.Ctx) error {
	_, id, _ := parseJWT(c)
	var cart models.Order

	if err := models.DB.Set("gorm:auto_preload", true).Where("Delivered = ? AND Canceled = ? AND Confirmed = ? AND id = ?", false, false, false, uint(id)).First(&cart).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "no such cart!"})
	}
	// else

	// find our order item
	var orderitem models.OrderItem
	if err := models.DB.Where("id = ?", c.Params("id")).First(&orderitem).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "no such order item!"})
	}

	models.DB.Delete(&orderitem)

	return c.Status(201).JSON(fiber.Map{"status": "Deleted!"})

}

// Checkout func
// @Summary checkout
// @Accept json
// @Produce json
// @Router /checkout [post]
func Checkout(c *fiber.Ctx) error {
	_, id, _ := parseJWT(c)
	var cart models.Order
	//models.DB.Set("gorm:auto_preload", true).Find(&user, "id = ?", c.Param("id"))
	if err := models.DB.Set("gorm:auto_preload", true).Where("Delivered = ? AND Canceled = ? AND Confirmed = ? AND id = ?", false, false, false, uint(id)).First(&cart).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "no such cart!"})
	}

	for _, order := range cart.Orders {
		// find an item
		var item models.Item
		if err := models.DB.Where("id = ?", order.ItemID).First(&item).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "no such item!"})
		}
		// remove ordered items from shop
		models.DB.Model(&item).Update("Qty", item.Qty-order.Qty)
	}

	models.DB.Model(&cart).Update("Confirmed", true)

	return c.Status(201).JSON(fiber.Map{"status": "created!"})

}

// Canceled func
// @Summary cancel cart
// @Accept json
// @Produce json
// @Router /cancel [post]
func Canceled(c *fiber.Ctx) error {
	_, id, _ := parseJWT(c)
	var cart models.Order
	//models.DB.Set("gorm:auto_preload", true).Find(&user, "id = ?", c.Param("id"))
	if err := models.DB.Set("gorm:auto_preload", true).Where("Delivered = ? AND Canceled = ? AND Confirmed = ? AND id = ?", false, false, false, uint(id)).First(&cart).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "no such cart!"})
	}

	models.DB.Model(&cart).Update("Confirmed", true)

	return c.Status(201).JSON(fiber.Map{"status": "created!"})

}

// Deliver func
// @Summary deliver
// @Accept json
// @Produce json
// @Param id path string true "cart id"
// @Router /deliver/{id} [post]
func Deliver(c *fiber.Ctx) error {
	_, _, isadmin := parseJWT(c)

	if isadmin {
		var cart models.Order
		//models.DB.Set("gorm:auto_preload", true).Find(&user, "id = ?", c.Param("id"))
		if err := models.DB.Set("gorm:auto_preload", true).Where("Delivered = ? AND Canceled = ? AND Confirmed = ? AND id = ?", false, false, false, c.Params("id")).First(&cart).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "no such cart!"})
		}

		models.DB.Model(&cart).Update("Delivered", true)

		return c.Status(201).JSON(fiber.Map{"status": "created!"})
	}

	return c.Status(304).JSON(fiber.Map{"status": "not authorized!"})

}
