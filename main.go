package main

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Model
type Pengeluaran struct {
	gorm.Model
	Nama      string  `json:"nama"`
	Keperluan string  `json:"keperluan"`
	Kategori  string  `json:"kategori"`
	Amount    float64 `json:"amount"`
}

var db *gorm.DB

func main() {

	// Init DB
	var err error
	db, err = gorm.Open(sqlite.Open("pengeluaran.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto migrate
	db.AutoMigrate(&Pengeluaran{})

	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"}, // izinkan semua origin
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))

	// Routes
	e.File("/", "index.html")
	e.POST("/pengeluaran", createPengeluaran)
	e.GET("/pengeluaran", getPengeluaran)
	e.GET("/pengeluaran/:id", getPengeluaranByID)
	e.PUT("/pengeluaran/:id", updatePengeluaran)
	e.DELETE("/pengeluaran/:id", deletePengeluaran)
	e.GET("/pengeluaran/filter", filterByDate)

	e.Logger.Fatal(e.Start(":8080"))
}

// CREATE
func createPengeluaran(c echo.Context) error {
	pengeluaran := new(Pengeluaran)
	if err := c.Bind(pengeluaran); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	db.Create(&pengeluaran)
	return c.JSON(http.StatusOK, pengeluaran)
}

// READ ALL
func getPengeluaran(c echo.Context) error {
	var list []Pengeluaran
	db.Find(&list)
	return c.JSON(http.StatusOK, list)
}

// READ BY ID
func getPengeluaranByID(c echo.Context) error {
	id := c.Param("id")
	var p Pengeluaran

	if err := db.First(&p, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{
			"message": "Data tidak ditemukan",
		})
	}

	return c.JSON(http.StatusOK, p)
}

// UPDATE
func updatePengeluaran(c echo.Context) error {
	id := c.Param("id")
	var p Pengeluaran

	if err := db.First(&p, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"message": "Data tidak ditemukan"})
	}

	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	db.Save(&p)
	return c.JSON(http.StatusOK, p)
}

// DELETE
func deletePengeluaran(c echo.Context) error {
	id := c.Param("id")
	var p Pengeluaran

	if err := db.First(&p, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"message": "Data tidak ditemukan"})
	}

	db.Delete(&p)
	return c.JSON(http.StatusOK, echo.Map{"message": "Data berhasil dihapus"})
}

func filterByDate(c echo.Context) error {
	start := c.QueryParam("start")
	end := c.QueryParam("end")

	// Validasi input
	if start == "" || end == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "start dan end date harus diisi (format: YYYY-MM-DD)",
		})
	}

	// Parse ke time
	startDate, err1 := time.Parse("2006-01-02", start)
	endDate, err2 := time.Parse("2006-01-02", end)

	if err1 != nil || err2 != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "Format date salah. Gunakan YYYY-MM-DD",
		})
	}

	var result []Pengeluaran

	db.Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Order("created_at asc").
		Find(&result)

	return c.JSON(http.StatusOK, result)
}
