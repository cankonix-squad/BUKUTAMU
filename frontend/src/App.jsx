import { Navigate, NavLink, Route, Routes, useNavigate } from 'react-router-dom'
import { session } from './api'
import Login from './pages/Login'
import FormTamu from './pages/FormTamu'
import DaftarKunjungan from './pages/DaftarKunjungan'

function Protected({ children }) {
  return session.get() ? children : <Navigate to="/login" replace />
}

function Layout({ children }) {
  const nav = useNavigate()
  const s = session.get()
  return (
    <>
      <header className="topbar">
        <div className="brand-wrap">
          <img className="logo-mark" src="/logo-bandung.svg" alt="Lambang Kota Bandung" />
          <div className="brand">
            <strong>Buku Tamu Digital</strong>
            <span>Dinas Sumber Daya Air dan Bina Marga Kota Bandung</span>
          </div>
        </div>
        <div className="topnav">
          <nav className="menu">
            <NavLink to="/" end>
              <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M15 12a4 4 0 1 0-4-4 4 4 0 0 0 4 4Zm-9-2V7H4v3H1v2h3v3h2v-3h3v-2H6Zm9 4c-2.7 0-8 1.3-8 4v2h16v-2c0-2.7-5.3-4-8-4Z" /></svg>
              Tamu Baru
            </NavLink>
            <NavLink to="/kunjungan">
              <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M19 3h-4.2A3 3 0 0 0 12 1a3 3 0 0 0-2.8 2H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V5a2 2 0 0 0-2-2Zm-7 0a1 1 0 1 1 0 2 1 1 0 0 1 0-2ZM7 8h10v2H7V8Zm0 4h10v2H7v-2Zm0 4h7v2H7v-2Z" /></svg>
              Daftar Kunjungan
            </NavLink>
          </nav>
          <div className="usr">
            <span className="avatar" aria-hidden="true">{(s?.nama || '?').trim().charAt(0).toUpperCase()}</span>
            <span className="usr-txt"><b>{s?.nama}</b><small>{s?.peran === 'admin' ? 'Admin' : 'Resepsionis'}</small></span>
          </div>
          <button className="logout" onClick={() => { session.clear(); nav('/login') }} title="Keluar">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" /><path d="M16 17l5-5-5-5" /><path d="M21 12H9" /></svg>
            <span>Keluar</span>
          </button>
        </div>
      </header>
      <main>{children}</main>
    </>
  )
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/" element={<Protected><Layout><FormTamu /></Layout></Protected>} />
      <Route path="/kunjungan" element={<Protected><Layout><DaftarKunjungan /></Layout></Protected>} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
