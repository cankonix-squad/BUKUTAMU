import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, session } from '../api'

const Ic = ({ d, size = 24, fill = 'currentColor' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill={fill} aria-hidden="true">{d}</svg>
)
const IcLogin = ({ size }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
    <path d="M14 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4" /><path d="M10 17l5-5-5-5" /><path d="M15 12H3" />
  </svg>
)
const IcUser = <path d="M12 12a4.5 4.5 0 1 0 0-9 4.5 4.5 0 0 0 0 9Zm0 2c-4.4 0-8 2.2-8 5v2h16v-2c0-2.8-3.6-5-8-5Z" />
const IcLock = <path d="M17 9V7a5 5 0 0 0-10 0v2a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2v-9a2 2 0 0 0-2-2Zm-8-2a3 3 0 0 1 6 0v2H9V7Zm4 9.7V18h-2v-1.3a2 2 0 1 1 2 0Z" />
const IcShield = <path d="M12 2 4 5v6c0 5 3.4 9.7 8 11 4.6-1.3 8-6 8-11V5l-8-3Zm-1.2 14.2-3.5-3.5 1.4-1.4 2.1 2.1 4.9-4.9 1.4 1.4-6.3 6.3Z" />
const IcBolt = <path d="M13 2 4 14h7l-1 8 9-12h-7l1-8Z" />
const IcChart = <path d="M4 20h16v2H4v-2Zm1-6h3v5H5v-5Zm5-4h3v9h-3v-9Zm5-5h3v14h-3V5Zm-9.3 3.3L10 4l3 3 4.3-4.3 1.4 1.4L13 9.8l-3-3-3.3 3.3-1-1.8Z" />
const IcEye = (open) => open
  ? <path d="M12 5C6.5 5 2.7 9.4 1.5 12c1.2 2.6 5 7 10.5 7s9.3-4.4 10.5-7C21.3 9.4 17.5 5 12 5Zm0 11a4 4 0 1 1 0-8 4 4 0 0 1 0 8Z" />
  : <path d="M2.3 3.7 3.7 2.3l18 18-1.4 1.4-3.3-3.3A11 11 0 0 1 12 19C6.5 19 2.7 14.6 1.5 12a14 14 0 0 1 4-4.9L2.3 3.7ZM12 5c5.5 0 9.3 4.4 10.5 7a14 14 0 0 1-2.9 3.9l-3.7-3.7A4 4 0 0 0 11.8 8L9.2 5.4C10.1 5.1 11 5 12 5Zm-3.9 5.5a4 4 0 0 0 5.4 5.4l-5.4-5.4Z" />

function Ilustrasi() {
  return (
    <svg className="lg-illu" viewBox="0 0 520 460" aria-hidden="true">
      <defs>
        <linearGradient id="gd" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stopColor="#f7fbff" /><stop offset="1" stopColor="#cfe0f5" /></linearGradient>
        <linearGradient id="gl" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stopColor="#8fbdea" /><stop offset="1" stopColor="#4f8fd0" /></linearGradient>
      </defs>
      <ellipse cx="430" cy="40" rx="60" ry="26" fill="#fff" opacity=".9" />
      <ellipse cx="390" cy="50" rx="40" ry="20" fill="#fff" opacity=".9" />
      {/* gedung */}
      <rect x="150" y="40" width="300" height="340" fill="url(#gd)" stroke="#b9cde6" />
      <rect x="140" y="30" width="320" height="16" fill="#e9f1fb" stroke="#b9cde6" />
      {Array.from({ length: 6 }).map((_, r) => Array.from({ length: 6 }).map((_, c) => (
        <rect key={r + '-' + c} x={168 + c * 46} y={62 + r * 50} width="34" height="36" fill="url(#gl)" opacity=".85" />
      )))}
      {/* pohon */}
      {[[120, 300, 50], [175, 330, 38], [470, 300, 45]].map(([x, y, r], i) => (
        <g key={i}><rect x={x - 4} y={y} width="8" height="70" fill="#7a5a3a" /><circle cx={x} cy={y - 8} r={r} fill="#5fae5a" /><circle cx={x - r / 3} cy={y - r / 2} r={r * .6} fill="#78c26e" /></g>
      ))}
      <rect x="60" y="370" width="460" height="14" rx="7" fill="#dbe7f5" />
      {/* tablet */}
      <g transform="rotate(8 260 300)">
        <rect x="150" y="170" width="200" height="270" rx="18" fill="#2d58a8" />
        <rect x="164" y="184" width="172" height="242" rx="10" fill="#f3f8ff" />
        <circle cx="250" cy="240" r="30" fill="#1561d6" /><circle cx="250" cy="232" r="10" fill="#fff" /><path d="M232 256a18 14 0 0 1 36 0Z" fill="#fff" />
        <rect x="200" y="290" width="100" height="8" rx="4" fill="#b8d0ef" /><rect x="210" y="306" width="80" height="8" rx="4" fill="#b8d0ef" />
        <circle cx="240" cy="370" r="30" fill="#0a4bb0" /><path d="M226 370l10 10 18-20" stroke="#fff" strokeWidth="7" fill="none" strokeLinecap="round" strokeLinejoin="round" />
      </g>
      {/* kartu ID */}
      <g transform="rotate(-4 380 380)">
        <rect x="372" y="262" width="12" height="40" rx="3" fill="#9bb4d4" />
        <rect x="335" y="300" width="96" height="140" rx="10" fill="#fff" stroke="#c9d9ee" />
        <circle cx="383" cy="345" r="16" fill="#1d4f9c" /><path d="M362 380a21 15 0 0 1 42 0Z" fill="#1d4f9c" />
        <rect x="355" y="395" width="56" height="6" rx="3" fill="#b8d0ef" /><rect x="360" y="408" width="46" height="6" rx="3" fill="#b8d0ef" />
      </g>
    </svg>
  )
}

export default function Login() {
  const nav = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [show, setShow] = useState(false)
  const [err, setErr] = useState('')
  const [loading, setLoading] = useState(false)

  async function submit(e) {
    e.preventDefault()
    setErr(''); setLoading(true)
    try {
      session.set(await api.login(username, password))
      nav('/')
    } catch (e) {
      setErr(e.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="lg">
      <section className="lg-hero">
        <div className="lg-brand">
          <img src="/logo-bandung.svg" alt="Lambang Kota Bandung" />
          <div className="sep" />
          <div><b>Dinas Sumber Daya Air dan Bina Marga</b><span>Kota Bandung</span></div>
        </div>
        <div className="lg-copy">
          <h1>Buku Tamu <em>Digital</em></h1>
          <div className="lg-bar"><i /><i /></div>
          <p>Pencatatan kunjungan tamu yang terverifikasi, cepat, dan menjaga perlindungan data pribadi.</p>
          <div className="lg-feat">
            <div><span className="ic"><Ic size={34} d={IcShield} /></span>Aman<br />&amp; Terpercaya</div>
            <div><span className="ic"><Ic size={34} d={IcBolt} /></span>Cepat<br />&amp; Mudah</div>
            <div><span className="ic"><Ic size={34} d={IcChart} /></span>Data<br />Terkelola</div>
          </div>
        </div>
        <Ilustrasi />
        <svg className="lg-wave" viewBox="0 0 1000 200" preserveAspectRatio="none" aria-hidden="true">
          <path d="M0 90 C 300 20, 600 30, 1000 140 L1000 200 L0 200Z" fill="#1561d6" opacity=".35" />
          <path d="M0 120 C 320 60, 620 70, 1000 170 L1000 200 L0 200Z" fill="#0a3f8f" />
          <path d="M620 200 C 760 170, 880 160, 1000 150 L1000 200Z" fill="var(--accent)" />
        </svg>
      </section>

      <div className="lg-side">
        <div className="lg-dots" />
        <form className="lg-card" onSubmit={submit}>
          <div className="head-ic"><IcLogin size={58} /></div>
          <h2>Masuk</h2>
          <p className="sub">Silakan login menggunakan akun resepsionis.</p>
          <label className="lg-field">
            <span className="lbl"><Ic d={IcUser} />Username</span>
            <span className="inp"><input value={username} onChange={(e) => setUsername(e.target.value)} placeholder="Masukkan username Anda..." autoComplete="username" autoFocus required /></span>
          </label>
          <label className="lg-field">
            <span className="lbl"><Ic d={IcLock} />Password</span>
            <span className="inp">
              <input type={show ? 'text' : 'password'} value={password} onChange={(e) => setPassword(e.target.value)} placeholder="Masukkan password Anda..." autoComplete="current-password" required />
              <button type="button" className="lg-eye" onClick={() => setShow(!show)} aria-label={show ? 'Sembunyikan password' : 'Tampilkan password'}><Ic d={IcEye(show)} /></button>
            </span>
          </label>
          {err && <div className="alert error">{err}</div>}
          <button className="lg-submit" disabled={loading}><IcLogin size={28} />{loading ? 'Memproses…' : 'Masuk'}</button>
          <div className="lg-secure"><Ic size={22} d={IcShield} />Keamanan Data Menjadi Prioritas Kami</div>
        </form>
      </div>
    </div>
  )
}
