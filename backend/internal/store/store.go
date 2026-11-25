// Package store berisi akses database untuk tamu, kunjungan, master data, dan audit.
package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dsdabm-bandung/buku-tamu/backend/internal/secure"
)

type Store struct {
	DB  *pgxpool.Pool
	NIK *secure.NIKCipher
}

type Bidang struct {
	ID    int    `json:"id"`
	Nama  string `json:"nama"`
	Aktif bool   `json:"aktif"`
}

type Pegawai struct {
	ID       int     `json:"id"`
	BidangID int     `json:"bidang_id"`
	NIP      *string `json:"nip"`
	Nama     string  `json:"nama"`
	Jabatan  *string `json:"jabatan"`
	Aktif    bool    `json:"aktif"`
}

type Pengguna struct {
	ID           int
	Username     string
	Nama         string
	PasswordHash string
	Peran        string
}

// Tamu = data KTP + kontak. NIK berisi nilai asli (hanya di memori).
type Tamu struct {
	ID               int64  `json:"id"`
	NIK              string `json:"nik,omitempty"`
	NIKMasked        string `json:"nik_masked"`
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
	NoHP             string `json:"no_hp"`
	Terverifikasi    bool   `json:"terverifikasi"`
	SumberData       string `json:"sumber_data"`
}

type Kunjungan struct {
	ID            int64      `json:"id"`
	TamuID        int64      `json:"tamu_id"`
	NamaTamu      string     `json:"nama_tamu"`
	NIKMasked     string     `json:"nik_masked"`
	Terverifikasi bool       `json:"terverifikasi"`
	BidangID      int        `json:"bidang_id"`
	NamaBidang    string     `json:"nama_bidang"`
	PegawaiID     *int       `json:"pegawai_id"`
	NamaPegawai   *string    `json:"nama_pegawai"`
	Keperluan     string     `json:"keperluan"`
	InstansiAsal  *string    `json:"instansi_asal"`
	AdaFoto       bool       `json:"ada_foto"`
	WaktuMasuk    time.Time  `json:"waktu_masuk"`
	WaktuKeluar   *time.Time `json:"waktu_keluar"`
}

var ErrNotFound = errors.New("data tidak ditemukan")

// ---------- Pengguna ----------

func (s *Store) PenggunaByUsername(ctx context.Context, username string) (*Pengguna, error) {
	p := &Pengguna{}
	err := s.DB.QueryRow(ctx, `SELECT id, username, nama, password_hash, peran FROM pengguna WHERE username=$1 AND aktif`, username).
		Scan(&p.ID, &p.Username, &p.Nama, &p.PasswordHash, &p.Peran)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return p, err
}

func (s *Store) CountPengguna(ctx context.Context) (int, error) {
	var n int
	err := s.DB.QueryRow(ctx, `SELECT count(*) FROM pengguna`).Scan(&n)
	return n, err
}

func (s *Store) CreatePengguna(ctx context.Context, username, nama, hash, peran string) error {
	_, err := s.DB.Exec(ctx, `INSERT INTO pengguna (username, nama, password_hash, peran) VALUES ($1,$2,$3,$4)`, username, nama, hash, peran)
	return err
}

// ---------- Master data ----------

func (s *Store) ListBidang(ctx context.Context, hanyaAktif bool) ([]Bidang, error) {
	rows, err := s.DB.Query(ctx, `SELECT id, nama, aktif FROM bidang WHERE ($1 = false OR aktif) ORDER BY nama`, hanyaAktif)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Bidang, error) {
		var b Bidang
		return b, r.Scan(&b.ID, &b.Nama, &b.Aktif)
	})
}

func (s *Store) CreateBidang(ctx context.Context, nama string) (int, error) {
	var id int
	err := s.DB.QueryRow(ctx, `INSERT INTO bidang (nama) VALUES ($1) RETURNING id`, nama).Scan(&id)
	return id, err
}

func (s *Store) ListPegawai(ctx context.Context, bidangID int) ([]Pegawai, error) {
	rows, err := s.DB.Query(ctx, `SELECT id, bidang_id, nip, nama, jabatan, aktif FROM pegawai
		WHERE aktif AND ($1 = 0 OR bidang_id = $1) ORDER BY nama`, bidangID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Pegawai, error) {
		var p Pegawai
		return p, r.Scan(&p.ID, &p.BidangID, &p.NIP, &p.Nama, &p.Jabatan, &p.Aktif)
	})
}

func (s *Store) CreatePegawai(ctx context.Context, p Pegawai) (int, error) {
	var id int
	err := s.DB.QueryRow(ctx, `INSERT INTO pegawai (bidang_id, nip, nama, jabatan) VALUES ($1,$2,$3,$4) RETURNING id`,
		p.BidangID, p.NIP, p.Nama, p.Jabatan).Scan(&id)
	return id, err
}

// ---------- Tamu ----------

func nullDate(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// UpsertTamu menyimpan/memperbarui data tamu berdasarkan NIK (via hash).
func (s *Store) UpsertTamu(ctx context.Context, tx pgx.Tx, t *Tamu) (int64, error) {
	enc, err := s.NIK.Encrypt(t.NIK)
	if err != nil {
		return 0, err
	}
	var verifiedAt any
	if t.Terverifikasi {
		verifiedAt = time.Now()
	}
	var id int64
	err = tx.QueryRow(ctx, `
		INSERT INTO tamu (nik_hash, nik_enc, nama, tempat_lahir, tanggal_lahir, jenis_kelamin, golongan_darah,
			alamat, rt, rw, kelurahan, kecamatan, kabupaten_kota, provinsi, agama, status_perkawinan,
			pekerjaan, kewarganegaraan, no_hp, terverifikasi, sumber_data, diverifikasi_pada)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
		ON CONFLICT (nik_hash) DO UPDATE SET
			nik_enc=EXCLUDED.nik_enc, nama=EXCLUDED.nama, tempat_lahir=EXCLUDED.tempat_lahir,
			tanggal_lahir=EXCLUDED.tanggal_lahir, jenis_kelamin=EXCLUDED.jenis_kelamin,
			golongan_darah=EXCLUDED.golongan_darah, alamat=EXCLUDED.alamat, rt=EXCLUDED.rt, rw=EXCLUDED.rw,
			kelurahan=EXCLUDED.kelurahan, kecamatan=EXCLUDED.kecamatan, kabupaten_kota=EXCLUDED.kabupaten_kota,
			provinsi=EXCLUDED.provinsi, agama=EXCLUDED.agama, status_perkawinan=EXCLUDED.status_perkawinan,
			pekerjaan=EXCLUDED.pekerjaan, kewarganegaraan=EXCLUDED.kewarganegaraan, no_hp=EXCLUDED.no_hp,
			terverifikasi=EXCLUDED.terverifikasi, sumber_data=EXCLUDED.sumber_data,
			diverifikasi_pada=COALESCE(EXCLUDED.diverifikasi_pada, tamu.diverifikasi_pada),
			diperbarui_pada=now()
		RETURNING id`,
		s.NIK.Hash(t.NIK), enc, t.Nama, t.TempatLahir, nullDate(t.TanggalLahir), t.JenisKelamin, t.GolonganDarah,
		t.Alamat, t.RT, t.RW, t.Kelurahan, t.Kecamatan, t.KabupatenKota, t.Provinsi, t.Agama, t.StatusPerkawinan,
		t.Pekerjaan, t.Kewarganegaraan, t.NoHP, t.Terverifikasi, t.SumberData, verifiedAt,
	).Scan(&id)
	return id, err
}

// TamuByNIK mencari tamu yang pernah berkunjung (untuk mengisi otomatis No. HP dsb).
func (s *Store) TamuByNIK(ctx context.Context, nik string) (*Tamu, error) {
	t := &Tamu{}
	var tgl *time.Time
	err := s.DB.QueryRow(ctx, `SELECT id, nama, COALESCE(tempat_lahir,''), tanggal_lahir, COALESCE(jenis_kelamin,''),
		COALESCE(golongan_darah,''), COALESCE(alamat,''), COALESCE(rt,''), COALESCE(rw,''), COALESCE(kelurahan,''),
		COALESCE(kecamatan,''), COALESCE(kabupaten_kota,''), COALESCE(provinsi,''), COALESCE(agama,''),
		COALESCE(status_perkawinan,''), COALESCE(pekerjaan,''), COALESCE(kewarganegaraan,''), COALESCE(no_hp,''),
		terverifikasi, sumber_data FROM tamu WHERE nik_hash=$1`, s.NIK.Hash(nik)).
		Scan(&t.ID, &t.Nama, &t.TempatLahir, &tgl, &t.JenisKelamin, &t.GolonganDarah, &t.Alamat, &t.RT, &t.RW,
			&t.Kelurahan, &t.Kecamatan, &t.KabupatenKota, &t.Provinsi, &t.Agama, &t.StatusPerkawinan,
			&t.Pekerjaan, &t.Kewarganegaraan, &t.NoHP, &t.Terverifikasi, &t.SumberData)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if tgl != nil {
		t.TanggalLahir = tgl.Format("2006-01-02")
	}
	t.NIK = nik
	t.NIKMasked = secure.MaskNIK(nik)
	return t, nil
}

// ---------- Kunjungan ----------

type KunjunganBaru struct {
	BidangID       int
	PegawaiID      *int
	Keperluan      string
	InstansiAsal   *string
	FotoPath       *string
	PersetujuanPDP bool
	DicatatOleh    int
}

func (s *Store) CreateKunjungan(ctx context.Context, t *Tamu, k KunjunganBaru) (int64, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	tamuID, err := s.UpsertTamu(ctx, tx, t)
	if err != nil {
		return 0, err
	}
	var id int64
	err = tx.QueryRow(ctx, `INSERT INTO kunjungan (tamu_id, bidang_id, pegawai_id, keperluan, instansi_asal, foto_path, persetujuan_pdp, dicatat_oleh)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		tamuID, k.BidangID, k.PegawaiID, k.Keperluan, k.InstansiAsal, k.FotoPath, k.PersetujuanPDP, k.DicatatOleh).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, tx.Commit(ctx)
}

type FilterKunjungan struct {
	Dari, Sampai time.Time
	BidangID     int
	HanyaAktif   bool // belum check-out
}

func (s *Store) ListKunjungan(ctx context.Context, f FilterKunjungan) ([]Kunjungan, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT k.id, k.tamu_id, t.nama, t.nik_enc, t.terverifikasi, k.bidang_id, b.nama, k.pegawai_id, p.nama,
		       k.keperluan, k.instansi_asal, k.foto_path IS NOT NULL, k.waktu_masuk, k.waktu_keluar
		FROM kunjungan k
		JOIN tamu t ON t.id = k.tamu_id
		JOIN bidang b ON b.id = k.bidang_id
		LEFT JOIN pegawai p ON p.id = k.pegawai_id
		WHERE k.waktu_masuk >= $1 AND k.waktu_masuk < $2
		  AND ($3 = 0 OR k.bidang_id = $3)
		  AND ($4 = false OR k.waktu_keluar IS NULL)
		ORDER BY k.waktu_masuk DESC`, f.Dari, f.Sampai, f.BidangID, f.HanyaAktif)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Kunjungan
	for rows.Next() {
		var k Kunjungan
		var enc []byte
		if err := rows.Scan(&k.ID, &k.TamuID, &k.NamaTamu, &enc, &k.Terverifikasi, &k.BidangID, &k.NamaBidang,
			&k.PegawaiID, &k.NamaPegawai, &k.Keperluan, &k.InstansiAsal, &k.AdaFoto, &k.WaktuMasuk, &k.WaktuKeluar); err != nil {
			return nil, err
		}
		if nik, err := s.NIK.Decrypt(enc); err == nil {
			k.NIKMasked = secure.MaskNIK(nik)
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

func (s *Store) CheckoutKunjungan(ctx context.Context, id int64) error {
	ct, err := s.DB.Exec(ctx, `UPDATE kunjungan SET waktu_keluar = now() WHERE id=$1 AND waktu_keluar IS NULL`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) FotoPath(ctx context.Context, id int64) (string, error) {
	var p *string
	err := s.DB.QueryRow(ctx, `SELECT foto_path FROM kunjungan WHERE id=$1`, id).Scan(&p)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && p == nil) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return *p, nil
}

// HapusKedaluwarsa menghapus kunjungan yang melewati masa retensi beserta tamu yang tidak lagi
// memiliki kunjungan. Mengembalikan path foto yang perlu dihapus dari disk.
func (s *Store) HapusKedaluwarsa(ctx context.Context, sebelum time.Time) ([]string, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, `DELETE FROM kunjungan WHERE waktu_masuk < $1 RETURNING foto_path`, sebelum)
	if err != nil {
		return nil, err
	}
	var paths []string
	for rows.Next() {
		var p *string
		if err := rows.Scan(&p); err != nil {
			rows.Close()
			return nil, err
		}
		if p != nil {
			paths = append(paths, *p)
		}
	}
	rows.Close()
	if _, err := tx.Exec(ctx, `DELETE FROM tamu t WHERE NOT EXISTS (SELECT 1 FROM kunjungan k WHERE k.tamu_id = t.id)`); err != nil {
		return nil, err
	}
	return paths, tx.Commit(ctx)
}

// ---------- Audit ----------

func (s *Store) Audit(ctx context.Context, penggunaID int, aksi, objek, objekID, ket, ip string) {
	var uid any
	if penggunaID > 0 {
		uid = penggunaID
	}
	_, _ = s.DB.Exec(ctx, `INSERT INTO audit_log (pengguna_id, aksi, objek, objek_id, keterangan, ip) VALUES ($1,$2,$3,$4,$5,$6)`,
		uid, aksi, objek, objekID, ket, ip)
}
