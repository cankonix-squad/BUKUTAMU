# Buku Tamu DSDABM Kota Bandung

Aplikasi buku tamu berbasis web untuk Dinas Sumber Daya Air dan Bina Marga Kota Bandung.

- **Backend**: Go (net/http), PostgreSQL (pgx)
- **Frontend**: React + Vite
- **Fitur**: cek NIK ke Dukcapil (adapter, saat ini mode mock), pencatatan seluruh elemen data KTP,
  tujuan kunjungan (bidang dan pegawai), foto tamu dari webcam, check-out, daftar kunjungan, log audit.

## Alur tamu

1. Resepsionis login, lalu mengetik NIK tamu dan menekan **Cek Dukcapil**.
2. Bila NIK ditemukan, data KTP terisi otomatis dan dikunci (terverifikasi).
   Bila tidak ditemukan atau layanan Dukcapil mati, data diisi manual dan ditandai "manual".
3. Pilih bidang dan pegawai tujuan, isi keperluan.
4. Ambil foto tamu dari webcam.
5. Tamu menyetujui pernyataan pemrosesan data pribadi, lalu **Simpan Kunjungan**.
6. Saat tamu pulang, tekan **Check-out** di menu Daftar Kunjungan.

## Menjalankan dengan Docker (paling mudah)

```bash
cp .env.example .env
# isi JWT_SECRET, NIK_ENC_KEY, NIK_HASH_KEY (openssl rand -hex 32) dan ADMIN_PASSWORD
docker compose up -d --build
```

Buka http://localhost:8000 dan login dengan `admin` / ADMIN_PASSWORD.

## Menjalankan untuk pengembangan

Butuh Go 1.26+, Node.js 20+, dan PostgreSQL 14+.

```bash
# database (atau pakai: docker compose up -d db)
createuser -P bukutamu       # password: bukutamu
createdb -O bukutamu bukutamu

# backend
cd backend
set -a; source ../.env; set +a
go run ./cmd/server          # http://localhost:8080, migrasi berjalan otomatis

# frontend (terminal lain)
cd frontend
npm install
npm run dev                  # http://localhost:5173, /api diteruskan ke :8080
```

Kamera browser hanya berjalan di `localhost` atau HTTPS. Di server produksi wajib pakai HTTPS.

### NIK contoh (mode mock)

| NIK | Nama |
|---|---|
| 3273010101900001 | BUDI SANTOSO |
| 3273024505950002 | SITI RAHAYU |

NIK lain akan dianggap "tidak ditemukan" sehingga bisa dipakai menguji input manual.

## Struktur

```
backend/
  cmd/server/            entry point, admin awal, job retensi
  internal/api/          HTTP handler REST
  internal/auth/         JWT + bcrypt, pembatasan per peran
  internal/db/           koneksi + migrasi SQL (migrations/*.sql)
  internal/dukcapil/     adapter Dukcapil: mock.go (fiktif), http.go (kerangka layanan resmi)
  internal/secure/       enkripsi NIK (AES-256-GCM), blind index (HMAC), penyamaran
  internal/store/        query database
frontend/
  src/pages/             Login, FormTamu, DaftarKunjungan
  src/components/        Kamera (webcam)
```

## API ringkas

| Metode | Path | Peran |
|---|---|---|
| POST | /api/auth/login | publik |
| GET | /api/nik/{nik} | admin, resepsionis |
| GET | /api/bidang, /api/pegawai?bidang_id= | admin, resepsionis |
| POST | /api/bidang, /api/pegawai, /api/pengguna | admin |
| GET | /api/kunjungan?dari=&sampai=&bidang_id=&aktif=1 | admin, resepsionis |
| POST | /api/kunjungan (multipart: `data` JSON + `foto`) | admin, resepsionis |
| POST | /api/kunjungan/{id}/keluar | admin, resepsionis |
| GET | /api/kunjungan/{id}/foto | admin, resepsionis |

## Integrasi Dukcapil

Akses data kependudukan memerlukan **Perjanjian Kerja Sama (PKS)** pemanfaatan data kependudukan
dengan Ditjen Dukcapil Kemendagri (pengajuan biasanya melalui Disdukcapil Kota Bandung).
Setelah PKS terbit dan spesifikasi teknis diterima:

1. Sesuaikan `backend/internal/dukcapil/http.go` (format request/response saat ini masih placeholder).
2. Set `DUKCAPIL_MODE=http` dan isi `DUKCAPIL_URL`, `DUKCAPIL_USER`, `DUKCAPIL_PASSWORD`, `DUKCAPIL_IP`.

Server selalu memverifikasi ulang NIK ke Dukcapil saat menyimpan, sehingga status "terverifikasi"
dan data KTP tidak dapat dipalsukan dari browser.

## Baca KTP: tap e-KTP dan foto KTP

Selain mengetik NIK, petugas bisa membaca identitas langsung dari KTP. NIK hasil bacaan tetap dicek ke Dukcapil;
kolom lain hasil bacaan hanya dipakai bila NIK tidak ditemukan.

**Foto KTP (OCR).** Tombol *Foto KTP* memotret KTP dengan webcam atau menerima unggahan foto, lalu membacanya
dengan tesseract.js di browser. Foto KTP tidak dikirim ke server. Data bahasa OCR (`ind`) diunduh dari CDN saat
pertama dipakai, jadi komputer resepsionis butuh internet untuk penggunaan pertama. Petugas wajib mencocokkan NIK
hasil bacaan dengan KTP asli.

**Tap KTP (card reader).** Browser tidak bisa membaca chip e-KTP, jadi tombol *Tap KTP* memanggil program
`ktp-agent` yang berjalan di komputer resepsionis dan hanya mendengarkan di `127.0.0.1:7788`.

```bash
cd backend
go run ./cmd/ktp-agent        # mode mock: setiap tap mengembalikan data fiktif bergantian
```

| Variabel | Bawaan | Keterangan |
|---|---|---|
| `KTP_AGENT_MODE` | `mock` | `mock` atau `pcsc` (reader resmi) |
| `KTP_AGENT_ADDR` | `127.0.0.1:7788` | Harus alamat loopback |
| `KTP_AGENT_ORIGIN` | `http://localhost:5173,http://localhost:8000` | Origin aplikasi yang diizinkan; isi domain produksi |
| `VITE_KTP_AGENT_URL` (frontend) | `http://127.0.0.1:7788` | Alamat agen |

Membaca chip e-KTP sungguhan membutuhkan card reader e-KTP bersertifikat dan **SAM** dari Ditjen Dukcapil
(ajukan bersama PKS). Setelah tersedia, implementasikan `pcscReader` di `backend/cmd/ktp-agent/main.go`
memakai SDK dari vendor reader.

## Pelindungan data pribadi (UU No. 27/2022)

- NIK disimpan terenkripsi AES-256-GCM; pencarian memakai HMAC sehingga NIK asli tidak tersimpan polos.
- NIK ditampilkan tersamar (`3273********0001`) di daftar kunjungan.
- Persetujuan tamu wajib dicentang sebelum data disimpan.
- Akses hanya untuk pengguna login, dibatasi per peran; setiap cek NIK, pencatatan, lihat foto, dan login dicatat di `audit_log`.
- Foto disimpan di server sendiri (volume `foto`), tidak di layanan cloud publik.
- Data kunjungan dan foto dihapus otomatis setelah `RETENTION_DAYS` hari.
- Simpan cadangan `NIK_ENC_KEY` dan `NIK_HASH_KEY` di tempat aman; jika hilang, data NIK tidak bisa dipulihkan.

## Rencana berikutnya

- Halaman admin untuk kelola bidang, pegawai, dan pengguna (API sudah ada)
- Notifikasi ke pegawai yang dituju (WhatsApp/email)
- Laporan/rekap kunjungan dan ekspor Excel
- Implementasi reader e-KTP resmi (`pcscReader`) setelah SAM tersedia
