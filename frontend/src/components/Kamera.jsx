import { useEffect, useRef, useState } from 'react'

// Mengambil foto tamu dari webcam. onCapture(blob|null)
export default function Kamera({ onCapture }) {
  const videoRef = useRef(null)
  const streamRef = useRef(null)
  const [preview, setPreview] = useState(null)
  const [err, setErr] = useState('')

  const aktifRef = useRef(false)

  async function start() {
    setErr('')
    aktifRef.current = true
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ video: { width: 640, height: 480, facingMode: 'user' } })
      if (!aktifRef.current) { // komponen sudah dilepas sebelum kamera siap
        stream.getTracks().forEach((t) => t.stop())
        return
      }
      streamRef.current = stream
      const v = videoRef.current
      if (v) {
        v.srcObject = stream
        await v.play().catch(() => {})
      }
    } catch {
      setErr('Kamera tidak dapat diakses. Pastikan izin kamera diberikan dan aplikasi dibuka lewat HTTPS atau localhost.')
    }
  }

  function stop() {
    aktifRef.current = false
    streamRef.current?.getTracks().forEach((t) => t.stop())
    streamRef.current = null
  }

  useEffect(() => { start(); return stop }, [])

  function capture() {
    const v = videoRef.current
    if (!v?.videoWidth) return
    const c = document.createElement('canvas')
    c.width = v.videoWidth; c.height = v.videoHeight
    c.getContext('2d').drawImage(v, 0, 0)
    c.toBlob((blob) => {
      setPreview(URL.createObjectURL(blob))
      onCapture(blob)
      stop()
    }, 'image/jpeg', 0.85)
  }

  function ulangi() {
    if (preview) URL.revokeObjectURL(preview)
    setPreview(null)
    onCapture(null)
    start()
  }

  return (
    <div className="kamera">
      <div className={`kam-frame ${preview ? 'done' : ''}`}>
        {preview ? <img src={preview} alt="Foto tamu" /> : <video ref={videoRef} autoPlay playsInline muted />}
        {!preview && !err && <span className="kam-guide" aria-hidden="true" />}
        {preview && <span className="kam-ok">✓ Foto tersimpan</span>}
        {err && <div className="kam-empty">
          <svg width="40" height="40" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M9 3h6l1.8 2H20a2 2 0 0 1 2 2v11a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V7a2 2 0 0 1 2-2h3.2L9 3Zm3 5a4.5 4.5 0 1 0 0 9 4.5 4.5 0 0 0 0-9Z" /></svg>
          Kamera tidak aktif</div>}
      </div>
      {err && <div className="alert error">{err}</div>}
      {preview
        ? <button type="button" className="secondary kam-btn" onClick={ulangi}>Ambil Ulang</button>
        : <button type="button" className="kam-btn" onClick={capture} disabled={!!err}>Ambil Foto</button>}
    </div>
  )
}
