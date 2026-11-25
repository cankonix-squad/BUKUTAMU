import { useEffect, useState } from 'react'
import { api } from '../api'
import Kamera from '../components/Kamera'
import ScanKTP from '../components/ScanKTP'
import { agenKTP } from '../ktp'

const kosong = {
  nik: '', nama: '', tempat_lahir: '', tanggal_lahir: '', jenis_kelamin: '', golongan_darah: '',
  alamat: '', rt: '', rw: '', kelurahan: '', kecamatan: '', kabupaten_kota: '', provinsi: '',
  agama: '', status_perkawinan: '', pekerjaan: '', kewarganegaraan: 'WNI', no_hp: '',
}

// Field KTP: [key, label, tipe/opsi, lebar kolom dari 12]
const identitas = [
  ['nama', 'Nama Lengkap', null, 6],
  ['tempat_lahir', 'Tempat Lahir', null, 3],
  ['tanggal_lahir', 'Tanggal Lahir', 'date', 3],
  ['jenis_kelamin', 'Jenis Kelamin', ['LAKI-LAKI', 'PEREMPUAN'], 3],
  ['golongan_darah', 'Golongan Darah', ['A', 'B', 'AB', 'O', '-'], 3],
  ['agama', 'Agama', ['ISLAM', 'KRISTEN', 'KATHOLIK', 'HINDU', 'BUDHA', 'KHONGHUCU', 'KEPERCAYAAN TERHADAP TUHAN YME'], 3],
  ['status_perkawinan', 'Status Perkawinan', ['BELUM KAWIN', 'KAWIN', 'CERAI HIDUP', 'CERAI MATI'], 3],
  ['pekerjaan', 'Pekerjaan', null, 6],
  ['kewarganegaraan', 'Kewarganegaraan', ['WNI', 'WNA'], 3],
]
const alamatKTP = [
  ['alamat', 'Alamat', null, 8],
  ['rt', 'RT', null, 2],
  ['rw', 'RW', null, 2],
  ['kelurahan', 'Kelurahan/Desa', null, 3],
  ['kecamatan', 'Kecamatan', null, 3],
  ['kabupaten_kota', 'Kabupaten/Kota', null, 3],
  ['provinsi', 'Provinsi', null, 3],
]

const I = {
  search: 'M10 2a8 8 0 0 1 6.3 12.9l5.4 5.4-1.4 1.4-5.4-5.4A8 8 0 1 1 10 2Zm0 2a6 6 0 1 0 0 12 6 6 0 0 0 0-12Z',
  id: 'M3 4h18a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2Zm5 4a2.5 2.5 0 1 0 0 5 2.5 2.5 0 0 0 0-5Zm-4 9h8c0-2-1.8-3-4-3s-4 1-4 3Zm10-8v2h6V9h-6Zm0 4v2h4v-2h-4Z',
  pin: 'M12 2a7 7 0 0 1 7 7c0 5.2-7 13-7 13S5 14.2 5 9a7 7 0 0 1 7-7Zm0 4.5a2.5 2.5 0 1 0 0 5 2.5 2.5 0 0 0 0-5Z',
  cam: 'M9 3h6l1.8 2H20a2 2 0 0 1 2 2v11a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V7a2 2 0 0 1 2-2h3.2L9 3Zm3 5a4.5 4.5 0 1 0 0 9 4.5 4.5 0 0 0 0-9Z',
  check: 'M9 16.2 4.8 12l-1.4 1.4L9 19 21 7l-1.4-1.4L9 16.2Z',
  save: 'M17 3H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V7l-4-4Zm-5 16a3 3 0 1 1 0-6 3 3 0 0 1 0 6Zm3-10H5V5h10v4Z',
  reset: 'M12 5V2L7 6l5 4V7a6 6 0 1 1-6 6H4a8 8 0 1 0 8-8Z',
  tap: 'M2 7a2 2 0 0 1 2-2h11a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V7Zm3 2v2h4V9H5Zm14.5-1.5 1.4-1.4a9 9 0 0 1 0 11.8l-1.4-1.4a7 7 0 0 0 0-9Zm-2.1 2.1 1.4-1.4a5 5 0 0 1 0 7.6l-1.4-1.4a3 3 0 0 0 0-4.8Z',
  info: 'M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20Zm1 15h-2v-6h2v6Zm0-8h-2V7h2v2Z',
}
const Ico = ({ d, size = 20 }) => <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d={d} /></svg>

export default function FormTamu() {
  const [tamu, setTamu] = useState(kosong)
  const [status, setStatus] = useState(null) // {tipe, pesan}
  const [terverifikasi, setTerverifikasi] = useState(false)
  const [mengecek, setMengecek] = useState(false)
  const [bidang, setBidang] = useState([])
  const [pegawai, setPegawai] = useState([])
  const [bidangId, setBidangId] = useState('')
  const [pegawaiId, setPegawaiId] = useState('')
  const [keperluan, setKeperluan] = useState('')
  const [instansi, setInstansi] = useState('')
  const [foto, setFoto] = useState(null)
  const [setuju, setSetuju] = useState(false)
  const [menyimpan, setMenyimpan] = useState(false)
  const [kameraKey, setKameraKey] = useState(0)
  const [scan, setScan] = useState(false)
  const [tapping, setTapping] = useState(false)
  const [agen, setAgen] = useState(null)

  useEffect(() => { agenKTP.status().then(setAgen) }, [])

  useEffect(() => { api.bidang().then(setBidang).catch((e) => setStatus({ tipe: 'error', pesan: e.message })) }, [])
  useEffect(() => {
    setPegawaiId('')
    if (!bidangId) { setPegawai([]); return }
    api.pegawai(bidangId).then(setPegawai).catch(() => setPegawai([]))
  }, [bidangId])

  const set = (k) => (e) => setTamu({ ...tamu, [k]: e.target.value })

  // nikArg dan cadangan dipakai saat NIK berasal dari tap/foto KTP: cadangan mengisi form
  // bila NIK tidak ditemukan di Dukcapil, agar petugas tidak mengetik ulang.
  async function cekNIK(nikArg, cadangan) {
    const nik = typeof nikArg === 'string' ? nikArg : tamu.nik
    if (!/^\d{16}$/.test(nik)) {
      setStatus({ tipe: 'error', pesan: 'NIK harus 16 digit angka' })
      return
    }
    setMengecek(true); setStatus(null)
    try {
      const r = await api.cekNIK(nik)
      if (r.ok) {
        const d = r.body.data
        setTamu({ ...kosong, ...Object.fromEntries(Object.entries(d).filter(([k]) => k in kosong)), nik, no_hp: r.body.no_hp || '' })
        setTerverifikasi(r.body.terverifikasi)
        setStatus(r.body.peringatan
          ? { tipe: 'warn', pesan: r.body.peringatan }
          : { tipe: 'ok', pesan: 'NIK terverifikasi Dukcapil. Data KTP terisi otomatis.' })
      } else {
        setTerverifikasi(false)
        if (cadangan) setTamu({ ...kosong, ...Object.fromEntries(Object.entries(cadangan).filter(([k]) => k in kosong)), nik })
        setStatus({ tipe: 'warn', pesan: cadangan
          ? `${r.body.error || 'NIK tidak ditemukan'}. Data diisi dari pembacaan KTP; periksa dan lengkapi sesuai KTP asli.`
          : `${r.body.error || 'NIK tidak ditemukan'}. Silakan isi data sesuai KTP secara manual.` })
      }
    } catch (e) {
      setStatus({ tipe: 'error', pesan: e.message })
    } finally {
      setMengecek(false)
    }
  }

  function pakaiKTP(d) {
    setScan(false)
    setTamu({ ...tamu, nik: d.nik }); setTerverifikasi(false)
    cekNIK(d.nik, d)
  }

  async function tapKTP() {
    setTapping(true); setStatus({ tipe: 'info', pesan: 'Tempelkan KTP elektronik ke card reader…' })
    try {
      pakaiKTP(await agenKTP.baca())
    } catch (e) {
      setStatus({ tipe: 'error', pesan: e.message })
    } finally {
      setTapping(false)
      agenKTP.status().then(setAgen)
    }
  }

  function reset() {
    setTamu(kosong); setTerverifikasi(false); setBidangId(''); setKeperluan(''); setInstansi('')
    setFoto(null); setSetuju(false); setKameraKey((k) => k + 1)
  }

  async function simpan(e) {
    e.preventDefault()
    if (!foto) { setStatus({ tipe: 'error', pesan: 'Foto tamu wajib diambil' }); return }
    setMenyimpan(true); setStatus(null)
    try {
      const r = await api.catatKunjungan({
        tamu,
        bidang_id: Number(bidangId),
        pegawai_id: pegawaiId ? Number(pegawaiId) : null,
        keperluan,
        instansi_asal: instansi || null,
        persetujuan_pdp: setuju,
      }, foto)
      reset()
      setStatus({ tipe: 'ok', pesan: `Kunjungan tercatat (No. ${r.id})${r.terverifikasi ? '' : ' — data belum terverifikasi Dukcapil'}.` })
      window.scrollTo(0, 0)
    } catch (e) {
      setStatus({ tipe: 'error', pesan: e.message })
    } finally {
      setMenyimpan(false)
    }
  }

  const field = ([k, label, tipe, span]) => (
    <label key={k} className={`f s${span}`}>
      <span className="f-lbl">{label}{k === 'nama' && <i className="req">*</i>}</span>
      {Array.isArray(tipe) ? (
        <select value={tamu[k]} onChange={set(k)} disabled={terverifikasi}>
          <option value="">Pilih…</option>
          {tipe.map((o) => <option key={o}>{o}</option>)}
        </select>
      ) : (
        <input type={tipe === 'date' ? 'date' : 'text'} value={tamu[k]} onChange={set(k)}
          readOnly={terverifikasi} required={k === 'nama'} />
      )}
    </label>
  )

  return (
    <form className="ft" onSubmit={simpan}>
      <header className="ft-head">
        <div>
          <h1>Pencatatan Tamu Baru</h1>
          <p>Cek NIK tamu ke Dukcapil, lengkapi tujuan kunjungan, lalu ambil foto tamu.</p>
        </div>
        <span className={`ft-chip ${terverifikasi ? 'ok' : ''}`}>
          <Ico size={16} d={terverifikasi ? I.check : I.info} />
          {terverifikasi ? 'Data terverifikasi Dukcapil' : 'Belum diverifikasi'}
        </span>
      </header>

      {status && <div className={`alert ${status.tipe}`}>{status.pesan}</div>}

      <div className="ft-layout">
        <div className="ft-main">
          <section className="card ft-card">
            <div className="ft-card-head"><span className="num">1</span><div><h3>Nomor Induk Kependudukan</h3><p>Masukkan 16 digit NIK sesuai KTP, lalu tekan Cek Dukcapil.</p></div></div>
            <div className="nik-bar">
              <div className="nik-inp">
                <Ico d={I.search} />
                <input className="nik" inputMode="numeric" maxLength={16} placeholder="Masukkan 16 digit NIK"
                  value={tamu.nik} onChange={(e) => { setTamu({ ...tamu, nik: e.target.value.replace(/\D/g, '') }); setTerverifikasi(false) }}
                  onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); cekNIK() } }} required />
                <span className="nik-count">{tamu.nik.length}/16</span>
              </div>
              <button type="button" className="btn-ic" onClick={cekNIK} disabled={mengecek}><Ico d={I.id} />{mengecek ? 'Mengecek…' : 'Cek Dukcapil'}</button>
            </div>
            <div className="ktp-alt">
              <span>Atau baca langsung dari KTP:</span>
              <button type="button" className="secondary btn-ic" onClick={tapKTP} disabled={tapping || mengecek || !agen?.siap}
                title={agen?.siap ? 'Baca chip e-KTP dengan card reader' : agen?.pesan}><Ico size={18} d={I.tap} />{tapping ? 'Menunggu KTP…' : 'Tap KTP'}</button>
              <button type="button" className="secondary btn-ic" onClick={() => setScan(true)} disabled={mengecek}><Ico size={18} d={I.cam} />Foto KTP</button>
              {agen && <small className={`agen ${agen.siap ? 'on' : ''}`}>{agen.siap ? `Card reader siap${agen.mode === 'mock' ? ' (mode uji)' : ''}` : agen.pesan}</small>}
            </div>
          </section>

          <section className="card ft-card">
            <div className="ft-card-head"><span className="num">2</span><div><h3>Data KTP</h3><p>{terverifikasi ? 'Terisi otomatis dari Dukcapil dan dikunci.' : 'Isi sesuai KTP bila NIK tidak ditemukan.'}</p></div></div>
            <div className="ft-sub"><Ico size={16} d={I.id} />Identitas</div>
            <div className="g12">
              {identitas.map(field)}
              <label className="f s3"><span className="f-lbl">No. HP</span><input type="tel" value={tamu.no_hp} onChange={set('no_hp')} placeholder="08xx" /></label>
            </div>
            <div className="ft-sub"><Ico size={16} d={I.pin} />Alamat sesuai KTP</div>
            <div className="g12">{alamatKTP.map(field)}</div>
          </section>

          <section className="card ft-card">
            <div className="ft-card-head"><span className="num">3</span><div><h3>Tujuan Kunjungan</h3><p>Bidang dan pegawai yang akan ditemui.</p></div></div>
            <div className="g12">
              <label className="f s6"><span className="f-lbl">Bidang Tujuan<i className="req">*</i></span>
                <select value={bidangId} onChange={(e) => setBidangId(e.target.value)} required>
                  <option value="">Pilih bidang…</option>
                  {bidang.map((b) => <option key={b.id} value={b.id}>{b.nama}</option>)}
                </select>
              </label>
              <label className="f s6"><span className="f-lbl">Pegawai yang Dituju</span>
                <select value={pegawaiId} onChange={(e) => setPegawaiId(e.target.value)} disabled={!pegawai.length}>
                  <option value="">{pegawai.length ? 'Pilih pegawai (opsional)…' : 'Belum ada data pegawai'}</option>
                  {pegawai.map((p) => <option key={p.id} value={p.id}>{p.nama}{p.jabatan ? ` — ${p.jabatan}` : ''}</option>)}
                </select>
              </label>
              <label className="f s12"><span className="f-lbl">Instansi/Asal <small>(opsional)</small></span><input value={instansi} onChange={(e) => setInstansi(e.target.value)} placeholder="Contoh: PT Maju Jaya, Kecamatan Coblong, atau pribadi" /></label>
              <label className="f s12"><span className="f-lbl">Keperluan<i className="req">*</i></span><textarea value={keperluan} onChange={(e) => setKeperluan(e.target.value)} rows={3} placeholder="Jelaskan keperluan kunjungan secara singkat" required /></label>
            </div>
          </section>
        </div>

        <aside className="ft-side">
          <section className="card ft-card">
            <div className="ft-card-head"><span className="num">4</span><div><h3>Foto Tamu</h3><p>Pastikan wajah terlihat jelas.</p></div></div>
            <Kamera key={kameraKey} onCapture={setFoto} />
          </section>

          <section className="card ft-card">
            <div className="ft-card-head"><span className="num">5</span><div><h3>Persetujuan &amp; Simpan</h3></div></div>
            <label className="consent">
              <input type="checkbox" checked={setuju} onChange={(e) => setSetuju(e.target.checked)} required />
              <span>
                Saya menyetujui pengumpulan dan pemrosesan data pribadi saya (NIK, data KTP, dan foto) oleh
                Dinas Sumber Daya Air dan Bina Marga Kota Bandung untuk keperluan pencatatan dan keamanan kunjungan,
                sesuai UU No. 27 Tahun 2022 tentang Pelindungan Data Pribadi.
              </span>
            </label>
            <div className="ft-actions">
              <button className="btn-ic primary-lg" disabled={menyimpan}><Ico d={I.save} />{menyimpan ? 'Menyimpan…' : 'Simpan Kunjungan'}</button>
              <button type="button" className="secondary btn-ic" onClick={reset}><Ico d={I.reset} />Bersihkan Form</button>
            </div>
          </section>
        </aside>
      </div>
      {scan && <ScanKTP onHasil={pakaiKTP} onTutup={() => setScan(false)} />}
    </form>
  )
}
