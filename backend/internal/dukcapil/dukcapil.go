// Package dukcapil adalah adapter untuk layanan verifikasi data kependudukan.
//
// Akses resmi ke data Dukcapil memerlukan Perjanjian Kerja Sama (PKS) pemanfaatan
// data kependudukan dengan Ditjen Dukcapil Kemendagri. Selama PKS diproses,
// gunakan DUKCAPIL_MODE=mock. Setelah kredensial resmi terbit, set DUKCAPIL_MODE=http
// dan sesuaikan httpClient dengan spesifikasi teknis yang diberikan Dukcapil.
package dukcapil

import (
	"context"
	"errors"
)

var (
	ErrNotFound    = errors.New("NIK tidak ditemukan di database kependudukan")
	ErrUnavailable = errors.New("layanan Dukcapil tidak dapat dihubungi")
)

// Penduduk memuat elemen data yang tercantum pada KTP-el.
type Penduduk struct {
	NIK              string `json:"nik"`
	Nama             string `json:"nama"`
	TempatLahir      string `json:"tempat_lahir"`
	TanggalLahir     string `json:"tanggal_lahir"` // YYYY-MM-DD
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

// Client adalah kontrak yang harus dipenuhi setiap implementasi (mock maupun resmi).
type Client interface {
	CekNIK(ctx context.Context, nik string) (*Penduduk, error)
}
