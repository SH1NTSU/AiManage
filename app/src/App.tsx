// App.jsx
import { BrowserRouter, Routes, Route, Outlet } from 'react-router-dom'
import './App.css'
import Nav from './components/Nav/Nav.tsx'
import Home from './components/Home/Home.tsx'
import Settings from './components/Settings/Settings.tsx'
import Stats from './components/Statistics/Statistics.tsx'

function Layout() {
  return (
    <>
      <Nav />
      <div className="main-content">
        <Outlet /> {/* This will render the current route */}
      </div>
    </>
  )
}

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Home />} /> 
          <Route path="settings" element={<Settings />} />
          <Route path="stats" element={<Stats />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App

