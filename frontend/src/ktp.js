// Pembacaan identitas KTP: lewat agen card reader (tap) atau OCR foto KTP.

const AGENT = import.meta.env?.VITE_KTP_AGENT_URL || 'http://127.0.0.1:7788'

export const agenKTP = {
  async status() {
    try {
      const r = await fetch(`${AGENT}/status`, { signal: AbortSignal.timeout(1500) })
      return r.ok ? await r.json() : { siap: false, pesan: `Agen KTP error (${r.status})` }
    } catch {
      return { siap: false, pesan: 'Agen card reader tidak berjalan di komputer ini' }
    }
  },
  async baca() {
    let r
    try {
      r = await fetch(`${AGENT}/baca`, { method: 'POST', signal: AbortSignal.timeout(25000) })
    } catch {
      throw new Error('Agen card reader tidak berjalan di komputer ini. Jalankan ktp-agent lalu coba lagi.')
    }
    const body = await r.json().catch(() => ({}))
    if (!r.ok) throw new Error(body.error || `Gagal membaca KTP (${r.status})`)
    return body
  },
}

// ---------- OCR ----------

let workerP = null
async function worker(onProgress) {
  if (!workerP) {
    workerP = import('tesseract.js').then(({ createWorker }) =>
      createWorker('ind', 1, { logger: (m) => m.status === 'recognizing text' && onProgress?.(m.progress) }))
  }
  return workerP
}

// Perbesar, ubah ke abu-abu, dan naikkan kontras agar huruf KTP lebih terbaca.
function siapkan(src) {
  const w = 1600, h = Math.round((src.height / src.width) * w)
  const c = document.createElement('canvas')
  c.width = w; c.height = h
  const g = c.getContext('2d')
  g.drawImage(src, 0, 0, w, h)
  const img = g.getImageData(0, 0, w, h), p = img.data
  for (let i = 0; i < p.length; i += 4) {
    const y = 0.299 * p[i] + 0.587 * p[i + 1] + 0.114 * p[i + 2]
    const v = Math.max(0, Math.min(255, (y - 128) * 1.8 + 140))
    p[i] = p[i + 1] = p[i + 2] = v
  }
  g.putImageData(img, 0, 0)
  return c
}

export async function ocrKTP(sumber, onProgress) {
  const w = await worker(onProgress)
  const { data } = await w.recognize(siapkan(sumber))
  return { teks: data.text, ...uraiKTP(data.text) }
}

const digit = (s) => s.replace(/[Oo]/g, '0').replace(/[Il|!]/g, '1').replace(/[S]/g, '5').replace(/[B]/g, '8').replace(/\D/g, '')
const bersih = (s) => (s || '').replace(/^[\s:;.\-=]+/, '').replace(/\s+/g, ' ').trim().toUpperCase()

const PILIHAN = {
  agama: ['ISLAM', 'KRISTEN', 'KATHOLIK', 'HINDU', 'BUDHA', 'KHONGHUCU'],
  status_perkawinan: ['BELUM KAWIN', 'CERAI HIDUP', 'CERAI MATI', 'KAWIN'],
}
const cocok = (v, opsi) => opsi.find((o) => v.replace(/\s/g, '').includes(o.replace(/\s/g, ''))) || ''

// Mengurai teks hasil OCR KTP menjadi field form. Field yang tidak terbaca dikosongkan.
export function uraiKTP(teks) {
  const baris = teks.split('\n').map((b) => b.trim()).filter(Boolean)
  const ambil = (re) => { for (const b of baris) { const m = b.match(re); if (m) return bersih(m[1]) } return '' }
  const d = {}

  // NIK: cari 16 digit setelah label NIK, atau deret 16 digit mana pun.
  const lNik = baris.find((b) => /N\s*[I1l]\s*K/i.test(b))
  const kand = [lNik && digit(lNik.replace(/^.*?N\s*[I1l]\s*K/i, '')), ...baris.map(digit)]
  d.nik = kand.find((k) => k && k.length === 16) || ''

  const prov = ambil(/PROVINSI\s+(.+)/i); if (prov) d.provinsi = prov
  const kota = ambil(/^((?:KOTA|KABUPATEN)\s+.+)/i); if (kota) d.kabupaten_kota = kota
  d.nama = ambil(/^Nama\s*[:;]?\s*(.+)/i)

  const ttl = ambil(/Tempat.*?Lahir\s*[:;]?\s*(.+)/i)
  if (ttl) {
    const m = ttl.match(/^(.*?)[,.\s]+(\d{2})[-\s./]?(\d{2})[-\s./]?(\d{4})/)
    if (m) { d.tempat_lahir = m[1].replace(/,$/, '').trim(); d.tanggal_lahir = `${m[4]}-${m[3]}-${m[2]}` }
    else d.tempat_lahir = ttl.split(',')[0]
  }
  const jk = ambil(/Jenis\s*Kelamin\s*[:;]?\s*(.+)/i)
  if (/PEREMPUAN/.test(jk)) d.jenis_kelamin = 'PEREMPUAN'
  else if (/LAKI/.test(jk)) d.jenis_kelamin = 'LAKI-LAKI'
  const gd = jk.match(/DARAH\s*[:;]?\s*(AB|A|B|O|-)/)
  if (gd) d.golongan_darah = gd[1]

  d.alamat = ambil(/^Alamat\s*[:;]?\s*(.+)/i)
  const rtrw = ambil(/RT\s*\/?\s*RW\s*[:;]?\s*(.+)/i).match(/(\d{1,3})\s*\/\s*(\d{1,3})/)
  if (rtrw) { d.rt = rtrw[1].padStart(3, '0'); d.rw = rtrw[2].padStart(3, '0') }
  d.kelurahan = ambil(/Kel\s*\/?\s*Desa\s*[:;]?\s*(.+)/i)
  d.kecamatan = ambil(/Kecamatan\s*[:;]?\s*(.+)/i)
  d.agama = cocok(ambil(/Agama\s*[:;]?\s*(.+)/i), PILIHAN.agama)
  d.status_perkawinan = cocok(ambil(/Status\s*Perkawinan\s*[:;]?\s*(.+)/i), PILIHAN.status_perkawinan)
  d.pekerjaan = ambil(/Pekerjaan\s*[:;]?\s*(.+)/i)
  const wn = ambil(/Kewarganegaraan\s*[:;]?\s*(.+)/i)
  if (wn) d.kewarganegaraan = /WNA/.test(wn) ? 'WNA' : 'WNI'

  return Object.fromEntries(Object.entries(d).filter(([, v]) => v))
}
