import { useEffect, useRef, useState } from 'react'
import { ocrKTP } from '../ktp'

// Modal untuk memotret KTP (webcam) atau mengunggah foto KTP, lalu membacanya dengan OCR.
// onHasil(data) dipanggil setelah petugas memeriksa dan menyetujui NIK hasil bacaan.
export default function ScanKTP({ onHasil, onTutup }) {
  const videoRef = useRef(null)
  const streamRef = useRef(null)
  const [err, setErr] = useState('')
  const [proses, setProses] = useState(null) // 0..1 saat OCR berjalan
  const [hasil, setHasil] = useState(null)
  const [nik, setNik] = useState('')
  const [pratinjau, setPratinjau] = useState(null)
  const [kameraOn, setKameraOn] = useState(false)

  useEffect(() => {
    let batal = false
    navigator.mediaDevices?.getUserMedia({ video: { width: { ideal: 1920 }, height: { ideal: 1080 }, facingMode: 'environment' } })
      .then((s) => {
        if (batal) { s.getTracks().forEach((t) => t.stop()); return }
        streamRef.current = s
        setKameraOn(true)
        if (videoRef.current) { videoRef.current.srcObject = s; videoRef.current.play().catch(() => {}) }
      })
      .catch(() => setErr('Kamera tidak dapat diakses. Anda tetap bisa mengunggah foto KTP.'))
    const esc = (e) => e.key === 'Escape' && onTutup()
    window.addEventListener('keydown', esc)
    return () => { batal = true; stopKamera(); window.removeEventListener('keydown', esc) }
  }, [])

  useEffect(() => () => pratinjau && URL.revokeObjectURL(pratinjau), [pratinjau])

  function stopKamera() { streamRef.current?.getTracks().forEach((t) => t.stop()); streamRef.current = null; setKameraOn(false) }

  async function baca(sumber, url) {
    setPratinjau(url); setErr(''); setHasil(null); setProses(0)
    try {
      const r = await ocrKTP(sumber, setProses)
      setHasil(r); setNik(r.nik || '')
      if (!r.nik) setErr('NIK tidak terbaca. Ketik NIK secara manual, atau foto ulang dengan cahaya lebih terang dan KTP memenuhi bingkai.')
    } catch {
      setErr('Gagal membaca KTP. Periksa koneksi internet (data bahasa OCR diunduh saat pertama dipakai), lalu coba lagi.')
    } finally {
      setProses(null)
    }
  }

  function potret() {
    const v = videoRef.current
    if (!v?.videoWidth) return
    const c = document.createElement('canvas')
    c.width = v.videoWidth; c.height = v.videoHeight
    c.getContext('2d').drawImage(v, 0, 0)
    stopKamera()
    c.toBlob((b) => baca(c, URL.createObjectURL(b)), 'image/jpeg', 0.9)
  }

  function unggah(e) {
    const f = e.target.files?.[0]
    if (!f) return
    stopKamera()
    const url = URL.createObjectURL(f)
    const img = new Image()
    img.onload = () => baca(img, url)
    img.src = url
  }

  const nikValid = /^\d{16}$/.test(nik)
  const terbaca = hasil ? Object.keys(hasil).filter((k) => !['teks', 'nik'].includes(k)).length : 0

  return (
    <div className="modal" onClick={onTutup}>
      <div className="modal-box scan" role="dialog" aria-modal="true" aria-label="Foto KTP" onClick={(e) => e.stopPropagation()}>
        <div className="scan-head">
          <h3>Baca KTP dari Foto</h3>
          <button type="button" className="scan-x" onClick={onTutup} aria-label="Tutup">×</button>
        </div>

        <div className="scan-frame">
          {pratinjau ? <img src={pratinjau} alt="Foto KTP" /> : <video ref={videoRef} autoPlay playsInline muted />}
          {!pratinjau && <span className="scan-guide" aria-hidden="true" />}
          {proses !== null && <div className="scan-busy"><span>Membaca KTP… {Math.round(proses * 100)}%</span><i style={{ width: `${proses * 100}%` }} /></div>}
        </div>
        {!pratinjau && <p className="scan-tip">Letakkan KTP di dalam bingkai, pastikan tidak silau dan tulisan tajam.</p>}

        {err && <div className="alert warn">{err}</div>}

        {hasil && (
          <div className="scan-hasil">
            <label className="f"><span className="f-lbl">NIK hasil bacaan — periksa dengan KTP asli</span>
              <input className="nik" inputMode="numeric" maxLength={16} value={nik} onChange={(e) => setNik(e.target.value.replace(/\D/g, ''))} onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); if (nikValid) onHasil({ ...hasil, nik }) } }} />
            </label>
            <p className="scan-info">{hasil.nama ? <>Nama terbaca: <b>{hasil.nama}</b>. </> : null}{terbaca} kolom lain terbaca dan dipakai bila NIK tidak ditemukan di Dukcapil.</p>
          </div>
        )}

        <div className="scan-act">
          {!pratinjau && <button type="button" onClick={potret} disabled={!kameraOn}>Potret KTP</button>}
          {pratinjau && proses === null && <button type="button" className="secondary" onClick={onTutup}>Batal</button>}
          {hasil && <button type="button" disabled={!nikValid} onClick={() => onHasil({ ...hasil, nik })}>Pakai &amp; Cek Dukcapil</button>}
          {!pratinjau && <label className="secondary scan-up">Unggah foto<input type="file" accept="image/*" onChange={unggah} hidden /></label>}
        </div>
        <p className="scan-note">Foto KTP hanya diproses di browser ini dan tidak disimpan maupun dikirim ke server.</p>
      </div>
    </div>
  )
}
