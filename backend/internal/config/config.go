package config

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Addr          string
	DatabaseURL   string
	JWTSecret     []byte
	NIKEncKey     []byte // 32 byte, AES-256
	NIKHashKey    []byte // kunci HMAC untuk indeks pencarian NIK
	StorageDir    string // folder penyimpanan foto
	RetentionDays int    // masa retensi data kunjungan (0 = tidak dihapus otomatis)
	CORSOrigin    string

	DukcapilMode     string // "mock" | "http"
	DukcapilURL      string
	DukcapilUser     string
	DukcapilPassword string
	DukcapilIP       string // IP terdaftar di PKS (beberapa layanan mewajibkan)

	AdminUsername string // akun admin awal, dibuat jika belum ada pengguna
	AdminPassword string
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Load() (*Config, error) {
	c := &Config{
		Addr:             env("ADDR", ":8080"),
		DatabaseURL:      env("DATABASE_URL", "postgres://bukutamu:bukutamu@localhost:5432/bukutamu?sslmode=disable"),
		JWTSecret:        []byte(os.Getenv("JWT_SECRET")),
		StorageDir:       env("STORAGE_DIR", "./data/foto"),
		CORSOrigin:       env("CORS_ORIGIN", "http://localhost:5173"),
		DukcapilMode:     env("DUKCAPIL_MODE", "mock"),
		DukcapilURL:      os.Getenv("DUKCAPIL_URL"),
		DukcapilUser:     os.Getenv("DUKCAPIL_USER"),
		DukcapilPassword: os.Getenv("DUKCAPIL_PASSWORD"),
		DukcapilIP:       os.Getenv("DUKCAPIL_IP"),
		AdminUsername:    env("ADMIN_USERNAME", "admin"),
		AdminPassword:    os.Getenv("ADMIN_PASSWORD"),
	}
	var err error
	if c.RetentionDays, err = strconv.Atoi(env("RETENTION_DAYS", "365")); err != nil {
		return nil, fmt.Errorf("RETENTION_DAYS: %w", err)
	}
	if len(c.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET wajib diisi, minimal 32 karakter")
	}
	if c.NIKEncKey, err = hexKey("NIK_ENC_KEY"); err != nil {
		return nil, err
	}
	if c.NIKHashKey, err = hexKey("NIK_HASH_KEY"); err != nil {
		return nil, err
	}
	return c, nil
}

func hexKey(name string) ([]byte, error) {
	b, err := hex.DecodeString(os.Getenv(name))
	if err != nil || len(b) != 32 {
		return nil, fmt.Errorf("%s wajib berupa 64 karakter hex (32 byte). Buat dengan: openssl rand -hex 32", name)
	}
	return b, nil
}
