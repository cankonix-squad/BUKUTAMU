import { useCallback, useEffect, useMemo, useState } from 'react'
import { api } from '../api'

const hariIni = () => new Date().toLocaleDateString('sv-SE') // YYYY-MM-DD waktu lokal
const jam = (t) => (t ? new Date(t).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' }) : '–')

export default function DaftarKunjungan() {
  const [dari, setDari] = useState(hariIni())
  const [sampai, setSampai] = useState(hariIni())
  const [bidangId, setBidangId] = useState('')
  const [aktif, setAktif] = useState(false)
  const [bidang, setBidang] = useState([])
  const [data, setData] = useState([])
  const [err, setErr] = useState('')
  const [foto, setFoto] = useState(null)
  const [cari, setCari] = useState('')

  useEffect(() => { api.bidang().then(setBidang).catch(() => {}) }, [])

  const muat = useCallback(() => {
    setErr('')
    api.kunjungan({ dari, sampai, bidang_id: bidangId || 0, aktif: aktif ? 1 : 0 })
      .then(setData).catch((e) => setErr(e.message))
  }, [dari, sampai, bidangId, aktif])
  useEffect(() => { muat() }, [muat])

  async function checkout(id) {
    try { await api.checkout(id); muat() } catch (e) { setErr(e.message) }
  }

  async function lihatFoto(id) {
    try { setFoto(URL.createObjectURL(await api.foto(id))) } catch (e) { setErr(e.message) }
  }

  function tutupFoto() { URL.revokeObjectURL(foto); setFoto(null) }
  useEffect(() => {
    if (!foto) return
    const h = (e) => { if (e.key === 'Escape') tutupFoto() }
    window.addEventListener('keydown', h)
    return () => window.removeEventListener('keydown', h)
  })

  const tampil = useMemo(() => {
    const q = cari.trim().toLowerCase()
    if (!q) return data
    return data.filter((k) => [k.nama_tamu, k.keperluan, k.nama_bidang, k.nama_pegawai, k.nik_masked].some((v) => v && String(v).toLowerCase().includes(q)))
  }, [data, cari])

  const stat = {
    total: data.length,
    aktif: data.filter((k) => !k.waktu_keluar).length,
    keluar: data.filter((k) => k.waktu_keluar).length,
    verif: data.filter((k) => k.terverifikasi).length,
  }

  return (
    <div className="dk">
      <header className="ft-head">
        <div>
          <h1>Daftar Kunjungan</h1>
          <p>Pantau tamu yang datang, lihat foto, dan lakukan check-out saat tamu pulang.</p>
        </div>
        <button type="button" className="secondary btn-ic" onClick={muat}>
          <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M17.7 6.3A8 8 0 1 0 20 12h-2a6 6 0 1 1-1.8-4.2L13 11h7V4l-2.3 2.3Z" /></svg>
          Muat Ulang
        </button>
      </header>

      <div className="dk-stats">
        <div className="dk-stat"><span className="lbl">Total Kunjungan</span><b>{stat.total}</b></div>
        <div className="dk-stat blue"><span className="lbl">Masih di Lokasi</span><b>{stat.aktif}</b></div>
        <div className="dk-stat gray"><span className="lbl">Sudah Check-out</span><b>{stat.keluar}</b></div>
        <div className="dk-stat green"><span className="lbl">Terverifikasi Dukcapil</span><b>{stat.verif}</b></div>
      </div>

      <section className="card dk-card">
        <div className="dk-filters">
          <label className="f dk-search"><span className="f-lbl">Cari</span>
            <span className="dk-search-inp">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M10 2a8 8 0 0 1 6.3 12.9l5.4 5.4-1.4 1.4-5.4-5.4A8 8 0 1 1 10 2Zm0 2a6 6 0 1 0 0 12 6 6 0 0 0 0-12Z" /></svg>
              <input value={cari} onChange={(e) => setCari(e.target.value)} placeholder="Nama, keperluan, bidang…" />
            </span>
          </label>
          <label className="f"><span className="f-lbl">Dari</span><input type="date" value={dari} onChange={(e) => setDari(e.target.value)} /></label>
          <label className="f"><span className="f-lbl">Sampai</span><input type="date" value={sampai} onChange={(e) => setSampai(e.target.value)} /></label>
          <label className="f"><span className="f-lbl">Bidang</span>
            <select value={bidangId} onChange={(e) => setBidangId(e.target.value)}>
              <option value="">Semua bidang</option>
              {bidang.map((b) => <option key={b.id} value={b.id}>{b.nama}</option>)}
            </select>
          </label>
          <label className="inline dk-toggle"><input type="checkbox" checked={aktif} onChange={(e) => setAktif(e.target.checked)} /> Masih di lokasi</label>
        </div>

        {err && <div className="alert error">{err}</div>}

        <div className="dk-table-wrap">
          <table className="dk-table">
            <thead>
              <tr><th>Tamu</th><th>NIK</th><th>Tujuan</th><th>Keperluan</th><th>Waktu</th><th>Status</th><th className="r">Aksi</th></tr>
            </thead>
            <tbody>
              {tampil.length === 0 && (
                <tr className="dk-empty"><td colSpan={7}>
                  <svg width="44" height="44" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M19 3h-4.2A3 3 0 0 0 12 1a3 3 0 0 0-2.8 2H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V5a2 2 0 0 0-2-2Zm-7 0a1 1 0 1 1 0 2 1 1 0 0 1 0-2ZM7 8h10v2H7V8Zm0 4h10v2H7v-2Zm0 4h7v2H7v-2Z" /></svg>
                  <b>{cari ? 'Tidak ada yang cocok dengan pencarian' : 'Belum ada kunjungan'}</b>
                  <span>{cari ? 'Coba kata kunci lain.' : 'Kunjungan pada rentang tanggal ini akan muncul di sini.'}</span>
                </td></tr>
              )}
              {tampil.map((k) => (
                <tr key={k.id}>
                  <td data-label="Tamu">
                    <div className="dk-guest">
                      <span className="dk-av">{(k.nama_tamu || '?').trim().charAt(0).toUpperCase()}</span>
                      <div>
                        <b>{k.nama_tamu}</b>
                        {k.terverifikasi ? <span className="pill ok">Terverifikasi</span> : <span className="pill warn">Manual</span>}
                      </div>
                    </div>
                  </td>
                  <td data-label="NIK" className="mono">{k.nik_masked}</td>
                  <td data-label="Tujuan"><b className="dk-strong">{k.nama_bidang}</b>{k.nama_pegawai ? <div className="muted">{k.nama_pegawai}</div> : null}</td>
                  <td data-label="Keperluan" className="dk-need">{k.keperluan}</td>
                  <td data-label="Waktu" className="dk-time">
                    <span><i className="in">Masuk</i>{jam(k.waktu_masuk)}</span>
                    <span><i className="out">Keluar</i>{jam(k.waktu_keluar)}</span>
                  </td>
                  <td data-label="Status">{k.waktu_keluar ? <span className="pill gray">Selesai</span> : <span className="pill blue"><i className="dot" />Di lokasi</span>}</td>
                  <td className="r dk-act"><div className="actions">
                    {k.ada_foto && <button className="secondary small" onClick={() => lihatFoto(k.id)}>Lihat Foto</button>}
                    {!k.waktu_keluar && <button className="small" onClick={() => checkout(k.id)}>Check-out</button>}
                  </div></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {tampil.length > 0 && <div className="dk-foot">Menampilkan {tampil.length} dari {data.length} kunjungan</div>}
      </section>

      {foto && (
        <div className="modal" role="dialog" aria-label="Foto tamu" onClick={tutupFoto}>
          <figure className="modal-box" onClick={(e) => e.stopPropagation()}>
            <img src={foto} alt="Foto tamu" />
            <button type="button" className="secondary" onClick={tutupFoto}>Tutup</button>
          </figure>
        </div>
      )}
    </div>
  )
}
