package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/dsdabm-bandung/buku-tamu/backend/internal/api"
	"github.com/dsdabm-bandung/buku-tamu/backend/internal/auth"
	"github.com/dsdabm-bandung/buku-tamu/backend/internal/config"
	"github.com/dsdabm-bandung/buku-tamu/backend/internal/db"
	"github.com/dsdabm-bandung/buku-tamu/backend/internal/dukcapil"
	"github.com/dsdabm-bandung/buku-tamu/backend/internal/secure"
	"github.com/dsdabm-bandung/buku-tamu/backend/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("konfigurasi: %v", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()
	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrasi: %v", err)
	}

	nikCipher, err := secure.NewNIKCipher(cfg.NIKEncKey, cfg.NIKHashKey)
	if err != nil {
		log.Fatal(err)
	}
	st := &store.Store{DB: pool, NIK: nikCipher}
	if err := seedAdmin(ctx, st, cfg); err != nil {
		log.Fatalf("admin awal: %v", err)
	}

	dc, err := dukcapil.New(cfg.DukcapilMode, cfg.DukcapilURL, cfg.DukcapilUser, cfg.DukcapilPassword, cfg.DukcapilIP)
	if err != nil {
		log.Fatal(err)
	}
	if cfg.DukcapilMode == "mock" {
		log.Println("PERINGATAN: Dukcapil berjalan dalam mode MOCK (data fiktif)")
	}
	if err := os.MkdirAll(cfg.StorageDir, 0o750); err != nil {
		log.Fatal(err)
	}

	srv := &api.Server{Store: st, Auth: auth.NewManager(cfg.JWTSecret), Dukcapil: dc,
		StorageDir: cfg.StorageDir, CORSOrigin: cfg.CORSOrigin}
	httpSrv := &http.Server{Addr: cfg.Addr, Handler: srv.Routes(), ReadHeaderTimeout: 10 * time.Second}

	if cfg.RetentionDays > 0 {
		go retensi(ctx, st, cfg)
	}

	go func() {
		log.Printf("server berjalan di %s", cfg.Addr)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
}

// seedAdmin membuat akun admin pertama bila tabel pengguna masih kosong.
func seedAdmin(ctx context.Context, st *store.Store, cfg *config.Config) error {
	n, err := st.CountPengguna(ctx)
	if err != nil || n > 0 {
		return err
	}
	if cfg.AdminPassword == "" {
		return errors.New("belum ada pengguna; isi ADMIN_PASSWORD untuk membuat akun admin awal")
	}
	hash, err := auth.HashPassword(cfg.AdminPassword)
	if err != nil {
		return err
	}
	log.Printf("membuat akun admin awal: %s", cfg.AdminUsername)
	return st.CreatePengguna(ctx, cfg.AdminUsername, "Administrator", hash, "admin")
}

// retensi menghapus data kunjungan yang melewati masa simpan, sekali sehari (prinsip minimisasi UU PDP).
func retensi(ctx context.Context, st *store.Store, cfg *config.Config) {
	run := func() {
		batas := time.Now().AddDate(0, 0, -cfg.RetentionDays)
		paths, err := st.HapusKedaluwarsa(ctx, batas)
		if err != nil {
			log.Printf("retensi: %v", err)
			return
		}
		for _, p := range paths {
			os.Remove(filepath.Join(cfg.StorageDir, filepath.Clean("/"+p)))
		}
		if len(paths) > 0 {
			log.Printf("retensi: %d foto kedaluwarsa dihapus", len(paths))
		}
	}
	run()
	t := time.NewTicker(24 * time.Hour)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			run()
		}
	}
}
