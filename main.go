package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

type User struct {
	gorm.Model
	// ID        string `gorm:"primarykey"`
	// CreatedAt time.Time
	// UpdatedAt time.Time
	// DeletedAt gorm.DeletedAt `gorm:"index"`
	Name      string    `json:"name" form:"name"`
	Email     string    `gorm:"unique" json:"email" form:"email"`
	Password  string    `json:"password" form:"password"`
	Phone     string    `json:"phone" form:"phone"`
	Address   string    `json:"address" form:"address"`
	StoreName string    `json:"store_name" form:"store_name"`
	Products  []Product `gorm:"foreignKey:UserID;references:ID"`
	Favorites []Favorite

}
type Product struct {
	ID          uint `gorm:"primarykey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	UserID      uint
	ProductName string
	Description string
	Price       float64
	Stock       int
	Type        string
	Favorites   []Favorite
}

type Favorite struct {
	gorm.Model
	UserID    uint 
	ProductID uint
}

func InitDB() {
	var err error
	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	// dsn := "root:qwerty123@tcp(127.0.0.1:3306)/db_be22_unit2?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", os.Getenv("DBUSER"), os.Getenv("DBPASS"), os.Getenv("DBHOST"), os.Getenv("DBPORT"), os.Getenv("DBNAME"))
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
}

func InitialMigration() {
	DB.AutoMigrate(&User{})
	DB.AutoMigrate(&Product{})
	DB.AutoMigrate(&Favorite{})
}

func main() {
	InitDB()
	fmt.Println("running")
	InitialMigration()

	// create new instance echo
	e := echo.New()
	e.GET("/users", GetAllUsersController)
	e.POST("/users", AddUserController)
	e.DELETE("/users/:id", DeleteUserByIdController)
	e.PUT("/users/:id", UpdateUserController)

	/*
		TODO:
		Update user by id
		CRUD product
		CRUD favorite
	*/

	// start server and port
	e.Logger.Fatal(e.Start(":8080"))
}

func GetAllUsersController(c echo.Context) error {
	var allUsers []User                          // var penampung data yg dibaca dari db
	tx := DB.Preload("Products").Find(&allUsers) // select * from users
	if tx.Error != nil {                         //check apakah terjadi error saat menjalankan query
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"status":  "failed",
			"message": "error read data",
		})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"status":  "success",
		"message": "success read data",
		"results": allUsers,
	})
}

func AddUserController(c echo.Context) error {
	newUser := User{}
	errBind := c.Bind(&newUser)
	if errBind != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"status":  "failed",
			"message": "error bind data: " + errBind.Error(),
		})
	}
	// data newArticle simpan ke DB
	tx := DB.Create(&newUser) //menjalankan query insert into users(...) values(....)
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"status":  "failed",
			"message": "error insert data " + tx.Error.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"status":  "success",
		"message": "success add user",
	})
}

func DeleteUserByIdController(c echo.Context) error {
	id := c.Param("id")
	idConv, errConv := strconv.Atoi(id)
	if errConv != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"status":  "failed",
			"message": "error convert id: " + errConv.Error(),
		})
	}

	tx := DB.Delete(&User{}, idConv)
	if tx.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"status":  "failed",
			"message": "error delete data " + tx.Error.Error(),
		})
	}
	if tx.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]any{
			"status":  "failed",
			"message": "data not found",
		})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"status":  "success",
		"message": "success delete user",
	})

}

func UpdateUserController(c echo.Context) error {
	// Ambil ID dari parameter URL
	id := c.Param("id")
	idConv, errConv := strconv.Atoi(id)
	if errConv != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "failed",
			"message": "error convert id: " + errConv.Error(),
		})
	}

	// Ambil data pengguna yang akan diperbarui dari body permintaan
	var updatedUser User
	if err := c.Bind(&updatedUser); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "failed",
			"message": "error bind data: " + err.Error(),
		})
	}

	// Cari pengguna berdasarkan ID
	var existingUser User
	if err := DB.First(&existingUser, idConv).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"status":  "failed",
			"message": "user not found",
		})
	}

	// Perbarui data pengguna
	existingUser.Name = updatedUser.Name
	existingUser.Email = updatedUser.Email
	existingUser.Password = updatedUser.Password
	existingUser.Phone = updatedUser.Phone
	existingUser.Address = updatedUser.Address
	existingUser.StoreName = updatedUser.StoreName

	// Simpan perubahan ke dalam database
	if err := DB.Save(&existingUser).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "failed",
			"message": "error updating user: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "user updated successfully",
	})
}

