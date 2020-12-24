package models

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" //pg drivers
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

//User schema
type User struct {
	gorm.Model
	Username  string // make it unique
	Email     string `gorm:"unique"`
	Password  string
	Phoneno   string
	Validated bool `gorm:"default:false"`
	Admin     bool `gorm:"default:false"`
	Banned    bool `gorm:"default:false"`
	Orders    []Order
}

// Category schema
type Category struct {
	gorm.Model
	Name        string
	Description string
}

// Product schema
type Product struct {
	gorm.Model
	Name        string
	Description string
	ImgURL      string
	ItemID      uint
	CategoryID  uint     // create ref onself beacuse it's a one to many rel and other solutions would result in overwritten rels
	Category    Category `gorm:"foreignkey:CategoryID"`
}

// Item schema
type Item struct {
	gorm.Model
	Product      Product
	Qty          int
	PricePerItem float32
	Disc         bool    `gorm:"default:false"`
	Top          bool    `gorm:"default:false"` // maens will show always on homepage
	DiscAmount   float32 // %
}

// OrderItem schema
type OrderItem struct {
	gorm.Model
	OrderID uint
	ItemID  uint // create ref onself beacuse it's a one to many rel and other solutions would result in overwritten rels
	Item    Item `gorm:"foreignkey:ItemID"`
	Qty     int
}

// Order schema
type Order struct {
	gorm.Model
	// Name         string
	UserID uint
	// Description  string
	Confirmed    bool `gorm:"default:false"` // user set
	Delivered    bool `gorm:"default:false"` // admin set
	Canceled     bool `gorm:"default:false"` // user set
	LocationDesc string
	// Price        float32 // calculate on front
	Orders []OrderItem
}

//DB var
var DB *gorm.DB

//ConnectDataBase globaly
func ConnectDataBase() {
	env := os.Getenv("DB")
	var db = "test.db"
	var dbName = "sqlite3"
	if env == "ps" {
		dbName = "postgres"
		// 30432
		db = "host=" + os.Getenv("DB_HOST") + " port=5432 user=" + os.Getenv("DB_USER") + " password=" + os.Getenv("DB_PASSWORD") + " dbname=" + os.Getenv("DB_NAME") + " sslmode=disable"
		//db = "host=localhost port=5432 user=postgres password=1234 dbname=postgres sslmode=disable"
	}
	database, err := gorm.Open(dbName, db)

	if err != nil {
		fmt.Println(err.Error())
		panic("Failed to connect to database!")
	}

	//database.DropTableIfExists(&User{}, &Post{}, &Comment{})

	database.AutoMigrate(&User{}, &Category{}, &Product{}, &Item{}, &OrderItem{}, &Order{})

	DB = database

	// seed with admin acc
	user := User{Username: "admin", Password: "admin", Phoneno: "555-123", Email: "", Validated: true, Admin: true}
	if err := DB.Create(&user); err.Error != nil {
		fmt.Println(err.Error)
	}
}
