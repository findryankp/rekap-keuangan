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
type Transaction struct {
	gorm.Model
	Nama       string    `json:"nama"`
	Keperluan  string    `json:"keperluan"`
	Kategori   string    `json:"kategori"`
	Amount     float64   `json:"amount"`
	Tipe       string    `json:"tipe"` // "pemasukan" atau "pengeluaran"
	Tanggal    string    `json:"tanggal"`
	ParsedDate time.Time `json:"parsed_date"`
}

var db *gorm.DB

func main() {
	// Init DB
	var err error
	db, err = gorm.Open(sqlite.Open("keuangan.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto migrate
	db.AutoMigrate(&Transaction{})

	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))

	// Routes
	e.File("/", "index.html")

	// Transaction routes (untuk pemasukan dan pengeluaran)
	e.POST("/transactions", createTransaction)
	e.GET("/transactions", getTransaction)
	e.GET("/transactions/:id", getTransactionByID)
	e.PUT("/transactions/:id", updateTransaction)
	e.DELETE("/transactions/:id", deleteTransaction)
	e.GET("/transactions/filter", filterByDate)

	// Rute khusus untuk laporan
	e.GET("/transactions/resume", getResume)
	e.GET("/transactions/resume/monthly", getResumeMonthly)

	e.Logger.Fatal(e.Start(":8080"))
}

// CREATE transaksi (pemasukan atau pengeluaran)
func createTransaction(c echo.Context) error {
	transaksi := new(Transaction)

	if err := c.Bind(transaksi); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	layout := "2006-01-02"
	parsedDate, err := time.Parse(layout, transaksi.Tanggal)
	if err != nil {
		panic(err)
	}

	transaksi.ParsedDate = parsedDate

	// Validasi tipe
	if transaksi.Tipe != "pemasukan" && transaksi.Tipe != "pengeluaran" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "Tipe harus 'pemasukan' atau 'pengeluaran'",
		})
	}

	// Set tanggal jika tidak diisi
	if transaksi.ParsedDate.IsZero() {
		transaksi.ParsedDate = time.Now()
	}

	db.Create(&transaksi)
	return c.JSON(http.StatusOK, transaksi)
}

// READ ALL transaksi
func getTransaction(c echo.Context) error {
	var list []Transaction
	db.Order("tanggal desc").Find(&list)
	return c.JSON(http.StatusOK, list)
}

// READ BY ID
func getTransactionByID(c echo.Context) error {
	id := c.Param("id")
	var t Transaction

	if err := db.First(&t, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{
			"message": "Data tidak ditemukan",
		})
	}

	return c.JSON(http.StatusOK, t)
}

// UPDATE
func updateTransaction(c echo.Context) error {
	id := c.Param("id")
	var t Transaction

	if err := db.First(&t, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"message": "Data tidak ditemukan"})
	}

	if err := c.Bind(&t); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	db.Save(&t)
	return c.JSON(http.StatusOK, t)
}

// DELETE
func deleteTransaction(c echo.Context) error {
	id := c.Param("id")
	var t Transaction

	if err := db.First(&t, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"message": "Data tidak ditemukan"})
	}

	db.Delete(&t)
	return c.JSON(http.StatusOK, echo.Map{"message": "Data berhasil dihapus"})
}

// FILTER BY DATE
func filterByDate(c echo.Context) error {
	start := c.QueryParam("start")
	end := c.QueryParam("end")
	tipe := c.QueryParam("tipe") // opsional: filter by type

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

	var result []Transaction
	query := db.Where("parsed_date BETWEEN ? AND ?", startDate, endDate.AddDate(0, 0, 1))

	// Filter by tipe jika ada
	if tipe != "" {
		query = query.Where("tipe = ?", tipe)
	}

	query.Order("parsed_date desc").Find(&result)

	return c.JSON(http.StatusOK, result)
}

// GET RINGKASAN (total pemasukan, pengeluaran, saldo)
func getResume(c echo.Context) error {
	var totalPemasukan, totalPengeluaran float64

	// Hitung total pemasukan
	db.Model(&Transaction{}).Where("tipe = ?", "pemasukan").Select("COALESCE(SUM(amount), 0)").Scan(&totalPemasukan)

	// Hitung total pengeluaran
	db.Model(&Transaction{}).Where("tipe = ?", "pengeluaran").Select("COALESCE(SUM(amount), 0)").Scan(&totalPengeluaran)

	saldo := totalPemasukan - totalPengeluaran

	return c.JSON(http.StatusOK, echo.Map{
		"total_pemasukan":   totalPemasukan,
		"total_pengeluaran": totalPengeluaran,
		"saldo":             saldo,
	})
}

// GET RINGKASAN BULANAN
func getResumeMonthly(c echo.Context) error {
	bulan := c.QueryParam("bulan") // format: YYYY-MM
	tahun := c.QueryParam("tahun") // format: YYYY

	var startDate, endDate time.Time
	var err error

	if bulan != "" {
		startDate, err = time.Parse("2006-01", bulan)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"message": "Format bulan salah. Gunakan YYYY-MM",
			})
		}
		endDate = startDate.AddDate(0, 1, 0)
	} else if tahun != "" {
		startDate, err = time.Parse("2006", tahun)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"message": "Format tahun salah. Gunakan YYYY",
			})
		}
		endDate = startDate.AddDate(1, 0, 0)
	} else {
		// Default bulan ini
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 1, 0)
	}

	var totalPemasukan, totalPengeluaran float64

	db.Model(&Transaction{}).
		Where("tipe = ? AND parsed_date BETWEEN ? AND ?", "pemasukan", startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").Scan(&totalPemasukan)

	db.Model(&Transaction{}).
		Where("tipe = ? AND parsed_date BETWEEN ? AND ?", "pengeluaran", startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").Scan(&totalPengeluaran)

	saldo := totalPemasukan - totalPengeluaran

	return c.JSON(http.StatusOK, echo.Map{
		"periode":           startDate.Format("2006-01"),
		"total_pemasukan":   totalPemasukan,
		"total_pengeluaran": totalPengeluaran,
		"saldo":             saldo,
		"start_date":        startDate.Format("2006-01-02"),
		"end_date":          endDate.Format("2006-01-02"),
	})
}
