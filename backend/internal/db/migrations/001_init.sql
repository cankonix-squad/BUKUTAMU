-- Skema awal buku tamu DSDABM Kota Bandung

CREATE TABLE IF NOT EXISTS bidang (
    id         SERIAL PRIMARY KEY,
    nama       TEXT NOT NULL UNIQUE,
    aktif      BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS pegawai (
    id         SERIAL PRIMARY KEY,
    bidang_id  INT NOT NULL REFERENCES bidang(id),
    nip        TEXT,
    nama       TEXT NOT NULL,
    jabatan    TEXT,
    no_hp      TEXT,
    aktif      BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX IF NOT EXISTS idx_pegawai_bidang ON pegawai(bidang_id);

-- Pengguna aplikasi (admin / resepsionis)
CREATE TABLE IF NOT EXISTS pengguna (
    id            SERIAL PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    nama          TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    peran         TEXT NOT NULL CHECK (peran IN ('admin','resepsionis')),
    aktif         BOOLEAN NOT NULL DEFAULT TRUE,
    dibuat_pada   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Data tamu sesuai elemen KTP-el.
-- NIK disimpan terenkripsi (AES-256-GCM); nik_hash (HMAC-SHA256) dipakai untuk pencarian.
CREATE TABLE IF NOT EXISTS tamu (
    id                 BIGSERIAL PRIMARY KEY,
    nik_hash           BYTEA NOT NULL UNIQUE,
    nik_enc            BYTEA NOT NULL,
    nama               TEXT NOT NULL,
    tempat_lahir       TEXT,
    tanggal_lahir      DATE,
    jenis_kelamin      TEXT,
    golongan_darah     TEXT,
    alamat             TEXT,
    rt                 TEXT,
    rw                 TEXT,
    kelurahan          TEXT,
    kecamatan          TEXT,
    kabupaten_kota     TEXT,
    provinsi           TEXT,
    agama              TEXT,
    status_perkawinan  TEXT,
    pekerjaan          TEXT,
    kewarganegaraan    TEXT,
    no_hp              TEXT,
    terverifikasi      BOOLEAN NOT NULL DEFAULT FALSE,
    sumber_data        TEXT NOT NULL DEFAULT 'manual', -- 'dukcapil' | 'manual'
    diverifikasi_pada  TIMESTAMPTZ,
    dibuat_pada        TIMESTAMPTZ NOT NULL DEFAULT now(),
    diperbarui_pada    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS kunjungan (
    id                 BIGSERIAL PRIMARY KEY,
    tamu_id            BIGINT NOT NULL REFERENCES tamu(id) ON DELETE CASCADE,
    bidang_id          INT NOT NULL REFERENCES bidang(id),
    pegawai_id         INT REFERENCES pegawai(id),
    keperluan          TEXT NOT NULL,
    instansi_asal      TEXT,
    foto_path          TEXT,
    persetujuan_pdp    BOOLEAN NOT NULL,
    waktu_masuk        TIMESTAMPTZ NOT NULL DEFAULT now(),
    waktu_keluar       TIMESTAMPTZ,
    dicatat_oleh       INT REFERENCES pengguna(id)
);
CREATE INDEX IF NOT EXISTS idx_kunjungan_masuk ON kunjungan(waktu_masuk);
CREATE INDEX IF NOT EXISTS idx_kunjungan_tamu ON kunjungan(tamu_id);

-- Log audit akses data pribadi (UU PDP)
CREATE TABLE IF NOT EXISTS audit_log (
    id          BIGSERIAL PRIMARY KEY,
    pengguna_id INT REFERENCES pengguna(id),
    aksi        TEXT NOT NULL,
    objek       TEXT,
    objek_id    TEXT,
    keterangan  TEXT,
    ip          TEXT,
    waktu       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_audit_waktu ON audit_log(waktu);

CREATE TABLE IF NOT EXISTS schema_migrations (
    versi TEXT PRIMARY KEY,
    waktu TIMESTAMPTZ NOT NULL DEFAULT now()
);
