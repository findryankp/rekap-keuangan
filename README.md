# Rekap Keuangan

Aplikasi rekap keuangan sederhana menggunakan Go (Echo Framework) dan SQLite.

## Cara Menjalankan

### Menggunakan Go

```bash
go run .
```

### Menggunakan PM2 (Windows)

Sesuai konfigurasi pengguna:

```bash
pm2 start go --run . --name="golang-rekap"
```

Atau gunakan script helper:

```bash
./start_pm2.bat
```

## Fitur

- Input Transaksi (Pemasukan/Pengeluaran)
- Kategori & Penanggung Jawab (CRUD)
- Dashboard Ringkasan & Grafik
- Export ke CSV

## Struktur Project

- `main.go`: Backend Logic
- `index.html`: Frontend Single Page Application
- `keuangan.db`: Database SQLite
