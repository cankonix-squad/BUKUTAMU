// ktp-agent berjalan di komputer resepsionis dan menjembatani card reader e-KTP
// dengan aplikasi web Buku Tamu. Agen hanya mendengarkan di 127.0.0.1, sehingga
// tidak bisa diakses dari komputer lain.
//
// Mode:
//   mock  : mensimulasikan tap KTP dengan data fiktif (untuk uji coba).
//   pcsc  : membaca card reader e-KTP bersertifikat. Membutuhkan SAM dari
//           Ditjen Dukcapil dan SDK reader; belum diimplementasikan.
//
// Variabel lingkungan:
//   KTP_AGENT_ADDR   alamat dengar (bawaan 127.0.0.1:7788)
//   KTP_AGENT_MODE   mock | pcsc (bawaan mock)
//   KTP_AGENT_ORIGIN origin aplikasi web yang diizinkan, pisahkan dengan koma
//                    (bawaan http://localhost:5173,http://localhost:8000)
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// DataKTP memakai nama field yang sama dengan form tamu di frontend.
type DataKTP struct {
	NIK              string `json:"nik"`
	Nama             string `json:"nama"`
	TempatLahir      string `json:"tempat_lahir"`
	TanggalLahir     string `json:"tanggal_lahir"`
	JenisKelamin     string `json:"jenis_kelamin"`
	GolonganDarah    string `json:"golongan_darah"`
	Alamat           string `json:"alamat"`
	RT               string `json:"rt"`
	RW               string `json:"rw"`
	Kelurahan        string `json:"kelurahan"`
	Kecamatan        string `json:"kecamatan"`
	KabupatenKota    string `json:"kabupaten_kota"`
	Provinsi         string `json:"provinsi"`
	Agama            string `json:"agama"`
	StatusPerkawinan string `json:"status_perkawinan"`
	Pekerjaan        string `json:"pekerjaan"`
	Kewarganegaraan  string `json:"kewarganegaraan"`
}

type Reader interface {
	Nama() string
	Siap() error
	// Baca menunggu kartu ditempel sampai ctx habis.
	Baca(ctx context.Context) (*DataKTP, error)
}

var errBelumSiap = errors.New("reader belum dikonfigurasi")

// mockReader bergantian mengembalikan dua penduduk fiktif yang juga ada di
// adapter Dukcapil mock, sehingga alur tap -> cek Dukcapil bisa diuji penuh.
type mockReader struct {
	mu sync.Mutex
	n  int
}

var contoh = []DataKTP{
	{NIK: "3273010101900001", Nama: "BUDI SANTOSO", TempatLahir: "BANDUNG", TanggalLahir: "1990-01-01",
		JenisKelamin: "LAKI-LAKI", GolonganDarah: "O", Alamat: "JL. CONTOH NO. 1", RT: "001", RW: "002",
		Kelurahan: "CITARUM", Kecamatan: "BANDUNG WETAN", KabupatenKota: "KOTA BANDUNG", Provinsi: "JAWA BARAT",
		Agama: "ISLAM", StatusPerkawinan: "KAWIN", Pekerjaan: "KARYAWAN SWASTA", Kewarganegaraan: "WNI"},
	{NIK: "3273024505950002", Nama: "SITI RAHAYU", TempatLahir: "GARUT", TanggalLahir: "1995-05-05",
		JenisKelamin: "PEREMPUAN", GolonganDarah: "A", Alamat: "JL. PERCOBAAN NO. 22", RT: "003", RW: "005",
		Kelurahan: "LEBAKSIUH", Kecamatan: "COBLONG", KabupatenKota: "KOTA BANDUNG", Provinsi: "JAWA BARAT",
		Agama: "ISLAM", StatusPerkawinan: "BELUM KAWIN", Pekerjaan: "PELAJAR/MAHASISWA", Kewarganegaraan: "WNI"},
}

func (*mockReader) Nama() string { return "mock" }
func (*mockReader) Siap() error  { return nil }
func (m *mockReader) Baca(ctx context.Context) (*DataKTP, error) {
	select {
	case <-time.After(1500 * time.Millisecond): // simulasi menunggu kartu ditempel
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	m.mu.Lock()
	d := contoh[m.n%len(contoh)]
	m.n++
	m.mu.Unlock()
	return &d, nil
}

// pcscReader adalah kerangka untuk card reader e-KTP resmi. Setelah SAM dan SDK
// dari vendor reader tersedia, implementasikan Siap() (cek reader + SAM) dan
// Baca() (tunggu kartu, autentikasi dengan SAM, baca data demografi).
type pcscReader struct{}

func (pcscReader) Nama() string { return "pcsc" }
func (pcscReader) Siap() error {
	return errors.New("reader e-KTP belum dikonfigurasi: butuh SAM dari Ditjen Dukcapil dan SDK reader")
}
func (r pcscReader) Baca(context.Context) (*DataKTP, error) { return nil, r.Siap() }

func main() {
	addr := env("KTP_AGENT_ADDR", "127.0.0.1:7788")
	if host, _, err := net.SplitHostPort(addr); err != nil || !(host == "127.0.0.1" || host == "localhost" || host == "::1") {
		log.Fatalf("KTP_AGENT_ADDR harus alamat loopback (127.0.0.1), bukan %q", addr)
	}
	origins := map[string]bool{}
	for _, o := range strings.Split(env("KTP_AGENT_ORIGIN", "http://localhost:5173,http://localhost:8000"), ",") {
		origins[strings.TrimSpace(o)] = true
	}

	var rd Reader
	switch mode := env("KTP_AGENT_MODE", "mock"); mode {
	case "mock":
		rd = &mockReader{}
	case "pcsc":
		rd = pcscReader{}
	default:
		log.Fatalf("KTP_AGENT_MODE tidak dikenal: %s", mode)
	}

	var sibuk sync.Mutex // satu reader hanya bisa melayani satu pembacaan
	mux := http.NewServeMux()
	mux.HandleFunc("GET /status", func(w http.ResponseWriter, r *http.Request) {
		st := map[string]any{"mode": rd.Nama(), "siap": true}
		if err := rd.Siap(); err != nil {
			st["siap"] = false
			st["pesan"] = err.Error()
		}
		tulis(w, http.StatusOK, st)
	})
	mux.HandleFunc("POST /baca", func(w http.ResponseWriter, r *http.Request) {
		if !sibuk.TryLock() {
			tulis(w, http.StatusConflict, map[string]string{"error": "reader sedang dipakai"})
			return
		}
		defer sibuk.Unlock()
		ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
		defer cancel()
		d, err := rd.Baca(ctx)
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			tulis(w, http.StatusRequestTimeout, map[string]string{"error": "tidak ada KTP yang ditempel"})
		case err != nil:
			tulis(w, http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
		default:
			tulis(w, http.StatusOK, d)
		}
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		o := r.Header.Get("Origin")
		if !origins[o] {
			http.Error(w, "origin tidak diizinkan", http.StatusForbidden)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", o)
		w.Header().Set("Vary", "Origin")
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			// Chrome meminta izin khusus agar halaman web boleh memanggil localhost.
			w.Header().Set("Access-Control-Allow-Private-Network", "true")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		mux.ServeHTTP(w, r)
	})

	log.Printf("ktp-agent mode=%s mendengarkan di http://%s", rd.Nama(), addr)
	srv := &http.Server{Addr: addr, Handler: h, ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(srv.ListenAndServe())
}

func tulis(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
