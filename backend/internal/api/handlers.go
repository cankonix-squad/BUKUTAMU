package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/dsdabm-bandung/buku-tamu/backend/internal/auth"
	"github.com/dsdabm-bandung/buku-tamu/backend/internal/dukcapil"
	"github.com/dsdabm-bandung/buku-tamu/backend/internal/secure"
	"github.com/dsdabm-bandung/buku-tamu/backend/internal/store"
)

// ---------- Auth ----------

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var in struct{ Username, Password string }
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, 400, "format permintaan tidak valid")
		return
	}
	p, err := s.Store.PenggunaByUsername(r.Context(), in.Username)
	if err != nil || !auth.CheckPassword(p.PasswordHash, in.Password) {
		s.Store.Audit(r.Context(), 0, "login_gagal", "pengguna", in.Username, "", clientIP(r))
		writeErr(w, 401, "username atau password salah")
		return
	}
	tok, err := s.Auth.Issue(p.ID, p.Nama, p.Peran)
	if err != nil {
		internalErr(w, err)
		return
	}
	s.Store.Audit(r.Context(), p.ID, "login", "pengguna", p.Username, "", clientIP(r))
	writeJSON(w, 200, map[string]any{"token": tok, "nama": p.Nama, "peran": p.Peran})
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	c := claims(r)
	writeJSON(w, 200, map[string]any{"id": c.UserID, "nama": c.Nama, "peran": c.Peran})
}

// createPengguna (khusus admin) menambah akun admin/resepsionis.
func (s *Server) createPengguna(w http.ResponseWriter, r *http.Request) {
	var in struct{ Username, Nama, Password, Peran string }
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.Username == "" || in.Nama == "" {
		writeErr(w, 400, "username dan nama wajib diisi")
		return
	}
	if len(in.Password) < 8 {
		writeErr(w, 400, "password minimal 8 karakter")
		return
	}
	if in.Peran != Admin && in.Peran != Resepsionis {
		writeErr(w, 400, "peran harus admin atau resepsionis")
		return
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		internalErr(w, err)
		return
	}
	if err := s.Store.CreatePengguna(r.Context(), in.Username, in.Nama, hash, in.Peran); err != nil {
		writeErr(w, 409, "username sudah dipakai")
		return
	}
	s.Store.Audit(r.Context(), claims(r).UserID, "tambah_pengguna", "pengguna", in.Username, in.Peran, clientIP(r))
	writeJSON(w, 201, map[string]bool{"ok": true})
}

// ---------- NIK / Dukcapil ----------

// cekNIK memverifikasi NIK ke Dukcapil, lalu melengkapi No. HP dari kunjungan sebelumnya bila ada.
func (s *Server) cekNIK(w http.ResponseWriter, r *http.Request) {
	nik := r.PathValue("nik")
	if !secure.ValidNIK(nik) {
		writeErr(w, 400, "NIK harus 16 digit angka")
		return
	}
	c := claims(r)
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	p, err := s.Dukcapil.CekNIK(ctx, nik)
	s.Store.Audit(r.Context(), c.UserID, "cek_nik", "tamu", secure.MaskNIK(nik), errString(err), clientIP(r))

	var noHP string
	if lama, e := s.Store.TamuByNIK(r.Context(), nik); e == nil {
		noHP = lama.NoHP
	}
	switch {
	case err == nil:
		writeJSON(w, 200, map[string]any{"terverifikasi": true, "sumber": "dukcapil", "data": p, "no_hp": noHP})
	case errors.Is(err, dukcapil.ErrNotFound):
		writeJSON(w, 404, map[string]any{"terverifikasi": false, "error": err.Error()})
	default:
		// Dukcapil tidak bisa dihubungi: tawarkan data kunjungan sebelumnya bila ada
		if lama, e := s.Store.TamuByNIK(r.Context(), nik); e == nil {
			writeJSON(w, 200, map[string]any{"terverifikasi": lama.Terverifikasi, "sumber": "riwayat", "data": lama, "no_hp": lama.NoHP,
				"peringatan": "Layanan Dukcapil tidak tersedia, data diambil dari kunjungan sebelumnya"})
			return
		}
		writeJSON(w, 503, map[string]any{"terverifikasi": false, "error": "Layanan Dukcapil tidak tersedia, silakan isi data secara manual"})
	}
}

func errString(err error) string {
	if err == nil {
		return "ok"
	}
	return err.Error()
}

// ---------- Master data ----------

func (s *Server) listBidang(w http.ResponseWriter, r *http.Request) {
	bs, err := s.Store.ListBidang(r.Context(), r.URL.Query().Get("semua") != "1")
	if err != nil {
		internalErr(w, err)
		return
	}
	writeJSON(w, 200, bs)
}

func (s *Server) createBidang(w http.ResponseWriter, r *http.Request) {
	var in struct{ Nama string }
	if json.NewDecoder(r.Body).Decode(&in) != nil || strings.TrimSpace(in.Nama) == "" {
		writeErr(w, 400, "nama bidang wajib diisi")
		return
	}
	id, err := s.Store.CreateBidang(r.Context(), strings.TrimSpace(in.Nama))
	if err != nil {
		internalErr(w, err)
		return
	}
	writeJSON(w, 201, map[string]int{"id": id})
}

func (s *Server) listPegawai(w http.ResponseWriter, r *http.Request) {
	bid, _ := strconv.Atoi(r.URL.Query().Get("bidang_id"))
	ps, err := s.Store.ListPegawai(r.Context(), bid)
	if err != nil {
		internalErr(w, err)
		return
	}
	writeJSON(w, 200, ps)
}

func (s *Server) createPegawai(w http.ResponseWriter, r *http.Request) {
	var p store.Pegawai
	if json.NewDecoder(r.Body).Decode(&p) != nil || p.BidangID == 0 || strings.TrimSpace(p.Nama) == "" {
		writeErr(w, 400, "bidang_id dan nama wajib diisi")
		return
	}
	id, err := s.Store.CreatePegawai(r.Context(), p)
	if err != nil {
		internalErr(w, err)
		return
	}
	writeJSON(w, 201, map[string]int{"id": id})
}

// ---------- Kunjungan ----------

const maxFoto = 5 << 20 // 5 MB

// createKunjungan menerima multipart/form-data:
//   - field "data": JSON {tamu:{...data KTP...}, bidang_id, pegawai_id, keperluan, instansi_asal, persetujuan_pdp}
//   - field "foto": file gambar JPEG/PNG
func (s *Server) createKunjungan(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxFoto+1<<20)
	if err := r.ParseMultipartForm(maxFoto); err != nil {
		writeErr(w, 400, "form tidak valid atau foto terlalu besar (maks 5 MB)")
		return
	}
	var in struct {
		Tamu           store.Tamu `json:"tamu"`
		BidangID       int        `json:"bidang_id"`
		PegawaiID      *int       `json:"pegawai_id"`
		Keperluan      string     `json:"keperluan"`
		InstansiAsal   *string    `json:"instansi_asal"`
		PersetujuanPDP bool       `json:"persetujuan_pdp"`
	}
	if err := json.Unmarshal([]byte(r.FormValue("data")), &in); err != nil {
		writeErr(w, 400, "data kunjungan tidak valid")
		return
	}
	t := &in.Tamu
	switch {
	case !secure.ValidNIK(t.NIK):
		writeErr(w, 400, "NIK harus 16 digit angka")
		return
	case strings.TrimSpace(t.Nama) == "":
		writeErr(w, 400, "nama wajib diisi")
		return
	case in.BidangID == 0:
		writeErr(w, 400, "bidang tujuan wajib dipilih")
		return
	case strings.TrimSpace(in.Keperluan) == "":
		writeErr(w, 400, "keperluan wajib diisi")
		return
	case !in.PersetujuanPDP:
		writeErr(w, 400, "tamu harus menyetujui pemrosesan data pribadi")
		return
	}

	// Status verifikasi ditentukan server, bukan klien: cek ulang ke Dukcapil.
	t.Terverifikasi, t.SumberData = false, "manual"
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	if p, err := s.Dukcapil.CekNIK(ctx, t.NIK); err == nil {
		applyDukcapil(t, p)
	}
	cancel()

	var fotoPath *string
	if f, hdr, err := r.FormFile("foto"); err == nil {
		defer f.Close()
		p, err := s.simpanFoto(f, hdr.Header.Get("Content-Type"))
		if err != nil {
			writeErr(w, 400, err.Error())
			return
		}
		fotoPath = &p
	}

	c := claims(r)
	id, err := s.Store.CreateKunjungan(r.Context(), t, store.KunjunganBaru{
		BidangID: in.BidangID, PegawaiID: in.PegawaiID, Keperluan: strings.TrimSpace(in.Keperluan),
		InstansiAsal: in.InstansiAsal, FotoPath: fotoPath, PersetujuanPDP: true, DicatatOleh: c.UserID,
	})
	if err != nil {
		if fotoPath != nil {
			os.Remove(filepath.Join(s.StorageDir, *fotoPath))
		}
		internalErr(w, err)
		return
	}
	s.Store.Audit(r.Context(), c.UserID, "catat_kunjungan", "kunjungan", fmt.Sprint(id), secure.MaskNIK(t.NIK), clientIP(r))
	writeJSON(w, 201, map[string]any{"id": id, "terverifikasi": t.Terverifikasi})
}

// applyDukcapil menimpa data KTP dengan data resmi Dukcapil agar tidak bisa dimanipulasi.
func applyDukcapil(t *store.Tamu, p *dukcapil.Penduduk) {
	t.Nama, t.TempatLahir, t.TanggalLahir = p.Nama, p.TempatLahir, p.TanggalLahir
	t.JenisKelamin, t.GolonganDarah, t.Alamat = p.JenisKelamin, p.GolonganDarah, p.Alamat
	t.RT, t.RW, t.Kelurahan, t.Kecamatan = p.RT, p.RW, p.Kelurahan, p.Kecamatan
	t.KabupatenKota, t.Provinsi, t.Agama = p.KabupatenKota, p.Provinsi, p.Agama
	t.StatusPerkawinan, t.Pekerjaan, t.Kewarganegaraan = p.StatusPerkawinan, p.Pekerjaan, p.Kewarganegaraan
	t.Terverifikasi, t.SumberData = true, "dukcapil"
}

func (s *Server) simpanFoto(f io.Reader, ctype string) (string, error) {
	head := make([]byte, 512)
	n, _ := io.ReadFull(f, head)
	head = head[:n]
	var ext string
	switch http.DetectContentType(head) {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	default:
		return "", errors.New("foto harus berformat JPEG atau PNG")
	}
	sub := time.Now().Format("2006/01")
	dir := filepath.Join(s.StorageDir, sub)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	rel := filepath.Join(sub, uuid.NewString()+ext)
	out, err := os.OpenFile(filepath.Join(s.StorageDir, rel), os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o640)
	if err != nil {
		return "", err
	}
	defer out.Close()
	if _, err := out.Write(head); err != nil {
		return "", err
	}
	if _, err := io.Copy(out, io.LimitReader(f, maxFoto)); err != nil {
		return "", err
	}
	return rel, nil
}

func (s *Server) listKunjungan(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	loc := time.Local
	today := time.Now().In(loc)
	dari := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, loc)
	sampai := dari.AddDate(0, 0, 1)
	if v := q.Get("dari"); v != "" {
		if d, err := time.ParseInLocation("2006-01-02", v, loc); err == nil {
			dari = d
		}
	}
	if v := q.Get("sampai"); v != "" {
		if d, err := time.ParseInLocation("2006-01-02", v, loc); err == nil {
			sampai = d.AddDate(0, 0, 1)
		}
	}
	bid, _ := strconv.Atoi(q.Get("bidang_id"))
	ks, err := s.Store.ListKunjungan(r.Context(), store.FilterKunjungan{
		Dari: dari, Sampai: sampai, BidangID: bid, HanyaAktif: q.Get("aktif") == "1",
	})
	if err != nil {
		internalErr(w, err)
		return
	}
	if ks == nil {
		ks = []store.Kunjungan{}
	}
	writeJSON(w, 200, ks)
}

func (s *Server) checkout(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, 400, "id tidak valid")
		return
	}
	if err := s.Store.CheckoutKunjungan(r.Context(), id); err != nil {
		if isNotFound(err) {
			writeErr(w, 404, "kunjungan tidak ditemukan atau sudah check-out")
			return
		}
		internalErr(w, err)
		return
	}
	s.Store.Audit(r.Context(), claims(r).UserID, "checkout", "kunjungan", fmt.Sprint(id), "", clientIP(r))
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (s *Server) foto(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, 400, "id tidak valid")
		return
	}
	rel, err := s.Store.FotoPath(r.Context(), id)
	if err != nil {
		if isNotFound(err) {
			writeErr(w, 404, "foto tidak ada")
			return
		}
		internalErr(w, err)
		return
	}
	full := filepath.Join(s.StorageDir, filepath.Clean("/"+rel))
	s.Store.Audit(r.Context(), claims(r).UserID, "lihat_foto", "kunjungan", fmt.Sprint(id), "", clientIP(r))
	w.Header().Set("Cache-Control", "private, no-store")
	http.ServeFile(w, r, full)
}
