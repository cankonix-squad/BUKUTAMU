package dukcapil

import (
	"context"
	"time"
)

// MockClient mensimulasikan layanan Dukcapil dengan data fiktif untuk pengembangan.
type MockClient struct{}

var mockData = map[string]Penduduk{
	"3273010101900001": {
		NIK: "3273010101900001", Nama: "BUDI SANTOSO", TempatLahir: "BANDUNG", TanggalLahir: "1990-01-01",
		JenisKelamin: "LAKI-LAKI", GolonganDarah: "O", Alamat: "JL. CONTOH NO. 1", RT: "001", RW: "002",
		Kelurahan: "CITARUM", Kecamatan: "BANDUNG WETAN", KabupatenKota: "KOTA BANDUNG", Provinsi: "JAWA BARAT",
		Agama: "ISLAM", StatusPerkawinan: "KAWIN", Pekerjaan: "KARYAWAN SWASTA", Kewarganegaraan: "WNI",
	},
	"3273024505950002": {
		NIK: "3273024505950002", Nama: "SITI RAHAYU", TempatLahir: "GARUT", TanggalLahir: "1995-05-05",
		JenisKelamin: "PEREMPUAN", GolonganDarah: "A", Alamat: "JL. PERCOBAAN NO. 22", RT: "003", RW: "005",
		Kelurahan: "LEBAKSIUH", Kecamatan: "COBLONG", KabupatenKota: "KOTA BANDUNG", Provinsi: "JAWA BARAT",
		Agama: "ISLAM", StatusPerkawinan: "BELUM KAWIN", Pekerjaan: "PELAJAR/MAHASISWA", Kewarganegaraan: "WNI",
	},
}

func (MockClient) CekNIK(ctx context.Context, nik string) (*Penduduk, error) {
	select {
	case <-time.After(300 * time.Millisecond): // simulasi latensi jaringan
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	p, ok := mockData[nik]
	if !ok {
		return nil, ErrNotFound
	}
	return &p, nil
}
