// Klien API sederhana. Token disimpan di sessionStorage (hilang saat tab ditutup).
const BASE = import.meta.env.VITE_API_URL || ''

export const session = {
  get: () => JSON.parse(sessionStorage.getItem('sesi') || 'null'),
  set: (s) => sessionStorage.setItem('sesi', JSON.stringify(s)),
  clear: () => sessionStorage.removeItem('sesi'),
}

async function request(path, opts = {}) {
  const s = session.get()
  const headers = { ...(opts.headers || {}) }
  if (s?.token) headers.Authorization = `Bearer ${s.token}`
  const res = await fetch(BASE + path, { ...opts, headers })
  if (res.status === 401 && path !== '/api/auth/login') {
    session.clear()
    window.location.href = '/login'
    throw new Error('Sesi berakhir, silakan login kembali')
  }
  const ctype = res.headers.get('content-type') || ''
  const body = ctype.includes('application/json') ? await res.json() : await res.blob()
  return { ok: res.ok, status: res.status, body }
}

function json(method, path, data) {
  return request(path, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: data ? JSON.stringify(data) : undefined,
  })
}

async function must(p) {
  const r = await p
  if (!r.ok) throw new Error(r.body?.error || `Gagal (${r.status})`)
  return r.body
}

export const api = {
  login: (username, password) => must(json('POST', '/api/auth/login', { username, password })),
  cekNIK: (nik) => request(`/api/nik/${nik}`),
  bidang: () => must(request('/api/bidang')),
  pegawai: (bidangId) => must(request(`/api/pegawai?bidang_id=${bidangId || 0}`)),
  kunjungan: (params) => must(request('/api/kunjungan?' + new URLSearchParams(params))),
  checkout: (id) => must(json('POST', `/api/kunjungan/${id}/keluar`)),
  foto: (id) => must(request(`/api/kunjungan/${id}/foto`)),
  catatKunjungan: (data, fotoBlob) => {
    const fd = new FormData()
    fd.append('data', JSON.stringify(data))
    if (fotoBlob) fd.append('foto', fotoBlob, 'foto.jpg')
    return must(request('/api/kunjungan', { method: 'POST', body: fd }))
  },
}
