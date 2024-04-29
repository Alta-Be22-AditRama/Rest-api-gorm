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

	e.GET("/products", GetAllProductsController)
	e.POST("/products", AddProductController)
	e.DELETE("/products/:id", DeleteProductController)
	e.PUT("/products/:id", UpdateProductController)

	e.GET("/favorites", GetAllFavoritesController)
	e.POST("/favorites", AddFavoriteController)
	e.DELETE("/favorites/:id", DeleteFavoriteController)
	e.PUT("/favorites/:id", UpdateFavoriteController)

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

func AddProductController(c echo.Context) error {
    newProduct := Product{}
    errBind := c.Bind(&newProduct)
    if errBind != nil {
        return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "status":  "failed",
            "message": "error bind data: " + errBind.Error(),
        })
    }
    // Simpan data produk ke DB
    tx := DB.Create(&newProduct)
    if tx.Error != nil {
        return c.JSON(http.StatusInternalServerError, map[string]interface{}{
            "status":  "failed",
            "message": "error insert data " + tx.Error.Error(),
        })
    }

    return c.JSON(http.StatusCreated, map[string]interface{}{
        "status":  "success",
        "message": "success add product",
    })
}

func GetAllProductsController(c echo.Context) error {
    var allProducts []Product
    tx := DB.Find(&allProducts)
    if tx.Error != nil {
        return c.JSON(http.StatusInternalServerError, map[string]interface{}{
            "status":  "failed",
            "message": "error read data",
        })
    }
    return c.JSON(http.StatusOK, map[string]interface{}{
        "status":  "success",
        "message": "success read data",
        "results": allProducts,
    })
}

func UpdateProductController(c echo.Context) error {
    // Ambil ID dari parameter URL
    id := c.Param("id")
    idConv, errConv := strconv.Atoi(id)
    if errConv != nil {
        return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "status":  "failed",
            "message": "error convert id: " + errConv.Error(),
        })
    }

    // Ambil data produk yang akan diperbarui dari body permintaan
    var updatedProduct Product
    if err := c.Bind(&updatedProduct); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "status":  "failed",
            "message": "error bind data: " + err.Error(),
        })
    }

    // Cari produk berdasarkan ID
    var existingProduct Product
    if err := DB.First(&existingProduct, idConv).Error; err != nil {
        return c.JSON(http.StatusNotFound, map[string]interface{}{
            "status":  "failed",
            "message": "product not found",
        })
    }

    // Perbarui data produk
    existingProduct.ProductName = updatedProduct.ProductName
    existingProduct.Description = updatedProduct.Description
    existingProduct.Price = updatedProduct.Price
    existingProduct.Stock = updatedProduct.Stock
    existingProduct.Type = updatedProduct.Type

    // Simpan perubahan ke dalam database
    if err := DB.Save(&existingProduct).Error; err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]interface{}{
            "status":  "failed",
            "message": "error updating product: " + err.Error(),
        })
    }

    return c.JSON(http.StatusOK, map[string]interface{}{
        "status":  "success",
        "message": "product updated successfully",
    })
}

func DeleteProductController(c echo.Context) error {
    // Ambil ID dari parameter URL
    id := c.Param("id")
    idConv, errConv := strconv.Atoi(id)
    if errConv != nil {
        return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "status":  "failed",
            "message": "error convert id: " + errConv.Error(),
        })
    }

    // Hapus produk dari database berdasarkan ID
    if err := DB.Delete(&Product{}, idConv).Error; err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]interface{}{
            "status":  "failed",
            "message": "error deleting product: " + err.Error(),
        })
    }

    return c.JSON(http.StatusOK, map[string]interface{}{
        "status":  "success",
        "message": "product deleted successfully",
    })
}

func AddFavoriteController(c echo.Context) error {
    newFavorite := Favorite{}
    errBind := c.Bind(&newFavorite)
    if errBind != nil {
        return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "status":  "failed",
            "message": "error bind data: " + errBind.Error(),
        })
    }
    // Simpan data favorite ke DB
    tx := DB.Create(&newFavorite)
    if tx.Error != nil {
        return c.JSON(http.StatusInternalServerError, map[string]interface{}{
            "status":  "failed",
            "message": "error insert data " + tx.Error.Error(),
        })
    }

    return c.JSON(http.StatusCreated, map[string]interface{}{
        "status":  "success",
        "message": "success add favorite",
    })
}

func GetAllFavoritesController(c echo.Context) error {
    var allFavorites []Favorite
    tx := DB.Find(&allFavorites)
    if tx.Error != nil {
        return c.JSON(http.StatusInternalServerError, map[string]interface{}{
            "status":  "failed",
            "message": "error read data",
        })
    }
    return c.JSON(http.StatusOK, map[string]interface{}{
        "status":  "success",
        "message": "success read data",
        "results": allFavorites,
    })
}

func UpdateFavoriteController(c echo.Context) error {
    // Ambil ID dari parameter URL
    id := c.Param("id")
    idConv, errConv := strconv.Atoi(id)
    if errConv != nil {
        return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "status":  "failed",
            "message": "error convert id: " + errConv.Error(),
        })
    }

    // Ambil data favorite yang akan diperbarui dari body permintaan
    var updatedFavorite Favorite
    if err := c.Bind(&updatedFavorite); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "status":  "failed",
            "message": "error bind data: " + err.Error(),
        })
    }

    // Cari favorite berdasarkan ID
    var existingFavorite Favorite
    if err := DB.First(&existingFavorite, idConv).Error; err != nil {
        return c.JSON(http.StatusNotFound, map[string]interface{}{
            "status":  "failed",
            "message": "favorite not found",
        })
    }

    // Perbarui data favorite
    existingFavorite.UserID = updatedFavorite.UserID
    existingFavorite.ProductID = updatedFavorite.ProductID

    // Simpan perubahan ke dalam database
    if err := DB.Save(&existingFavorite).Error; err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]interface{}{
            "status":  "failed",
            "message": "error updating favorite: " + err.Error(),
        })
    }

    return c.JSON(http.StatusOK, map[string]interface{}{
        "status":  "success",
        "message": "favorite updated successfully",
    })
}

func DeleteFavoriteController(c echo.Context) error {
    // Ambil ID dari parameter URL
    id := c.Param("id")
    idConv, errConv := strconv.Atoi(id)
    if errConv != nil {
        return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "status":  "failed",
            "message": "error convert id: " + errConv.Error(),
        })
    }

    // Hapus favorite dari database berdasarkan ID
    if err := DB.Delete(&Favorite{}, idConv).Error; err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]interface{}{
            "status":  "failed",
            "message": "error deleting favorite: " + err.Error(),
        })
    }

    return c.JSON(http.StatusOK, map[string]interface{}{
        "status":  "success",
        "message": "favorite deleted successfully",
    })
}
