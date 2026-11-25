// Package api berisi HTTP handler REST untuk aplikasi buku tamu.
package api

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/dsdabm-bandung/buku-tamu/backend/internal/auth"
	"github.com/dsdabm-bandung/buku-tamu/backend/internal/dukcapil"
	"github.com/dsdabm-bandung/buku-tamu/backend/internal/store"
)

type Server struct {
	Store      *store.Store
	Auth       *auth.Manager
	Dukcapil   dukcapil.Client
	StorageDir string
	CORSOrigin string
}

const (
	Admin       = "admin"
	Resepsionis = "resepsionis"
)

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	staf := s.Auth.Require(Admin, Resepsionis)
	admin := s.Auth.Require(Admin)

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /api/auth/login", s.login)

	mux.Handle("GET /api/me", staf(http.HandlerFunc(s.me)))
	mux.Handle("GET /api/nik/{nik}", staf(http.HandlerFunc(s.cekNIK)))
	mux.Handle("GET /api/bidang", staf(http.HandlerFunc(s.listBidang)))
	mux.Handle("POST /api/bidang", admin(http.HandlerFunc(s.createBidang)))
	mux.Handle("GET /api/pegawai", staf(http.HandlerFunc(s.listPegawai)))
	mux.Handle("POST /api/pegawai", admin(http.HandlerFunc(s.createPegawai)))
	mux.Handle("POST /api/pengguna", admin(http.HandlerFunc(s.createPengguna)))
	mux.Handle("GET /api/kunjungan", staf(http.HandlerFunc(s.listKunjungan)))
	mux.Handle("POST /api/kunjungan", staf(http.HandlerFunc(s.createKunjungan)))
	mux.Handle("POST /api/kunjungan/{id}/keluar", staf(http.HandlerFunc(s.checkout)))
	mux.Handle("GET /api/kunjungan/{id}/foto", staf(http.HandlerFunc(s.foto)))

	return s.cors(securityHeaders(mux))
}

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.CORSOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", s.CORSOrigin)
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Cache-Control", "no-store") // data pribadi jangan di-cache
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func internalErr(w http.ResponseWriter, err error) {
	log.Printf("internal error: %v", err)
	writeErr(w, 500, "terjadi kesalahan pada server")
}

func clientIP(r *http.Request) string {
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		return strings.TrimSpace(strings.Split(xf, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func claims(r *http.Request) *auth.Claims {
	c, _ := auth.FromContext(r.Context())
	if c == nil {
		return &auth.Claims{}
	}
	return c
}

func isNotFound(err error) bool { return errors.Is(err, store.ErrNotFound) }
