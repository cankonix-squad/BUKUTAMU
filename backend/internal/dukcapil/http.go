package dukcapil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HTTPClient adalah kerangka klien untuk web service resmi Dukcapil.
//
// PENTING: format request/response di bawah ini adalah placeholder. Spesifikasi
// sebenarnya (endpoint, nama field, autentikasi, whitelist IP) diberikan oleh
// Ditjen Dukcapil setelah PKS ditandatangani. Sesuaikan fungsi CekNIK dan
// struct response di bawah dengan dokumen teknis tersebut.
type HTTPClient struct {
	URL      string
	User     string
	Password string
	IP       string
	HTTP     *http.Client
}

func NewHTTPClient(url, user, password, ip string) *HTTPClient {
	return &HTTPClient{URL: url, User: user, Password: password, IP: ip,
		HTTP: &http.Client{Timeout: 15 * time.Second}}
}

type httpRequest struct {
	NIK      string `json:"NIK"`
	UserID   string `json:"user_id"`
	Password string `json:"password"`
	IPUser   string `json:"ip_user"`
}

type httpResponse struct {
	Content []map[string]string `json:"content"`
}

func (c *HTTPClient) CekNIK(ctx context.Context, nik string) (*Penduduk, error) {
	body, _ := json.Marshal(httpRequest{NIK: nik, UserID: c.User, Password: c.Password, IPUser: c.IP})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", ErrUnavailable, resp.StatusCode)
	}
	var r httpResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("%w: respons tidak valid", ErrUnavailable)
	}
	if len(r.Content) == 0 {
		return nil, ErrNotFound
	}
	d := r.Content[0]
	if d["RESPON"] != "" { // contoh: layanan mengembalikan pesan bila NIK tidak ditemukan
		return nil, ErrNotFound
	}
	return &Penduduk{
		NIK: nik, Nama: d["NAMA_LGKP"], TempatLahir: d["TMPT_LHR"], TanggalLahir: d["TGL_LHR"],
		JenisKelamin: d["JENIS_KLMIN"], GolonganDarah: d["GOL_DARAH"], Alamat: d["ALAMAT"],
		RT: d["NO_RT"], RW: d["NO_RW"], Kelurahan: d["KEL_NAME"], Kecamatan: d["KEC_NAME"],
		KabupatenKota: d["KAB_NAME"], Provinsi: d["PROP_NAME"], Agama: d["AGAMA"],
		StatusPerkawinan: d["STATUS_KAWIN"], Pekerjaan: d["JENIS_PKRJN"], Kewarganegaraan: "WNI",
	}, nil
}

// New memilih implementasi sesuai konfigurasi.
func New(mode, url, user, password, ip string) (Client, error) {
	switch mode {
	case "mock":
		return MockClient{}, nil
	case "http":
		if url == "" {
			return nil, fmt.Errorf("DUKCAPIL_URL wajib diisi untuk mode http")
		}
		return NewHTTPClient(url, user, password, ip), nil
	default:
		return nil, fmt.Errorf("DUKCAPIL_MODE tidak dikenal: %q", mode)
	}
}
