// App.jsx
import { BrowserRouter, Routes, Route, Outlet } from 'react-router-dom'
import './App.css'
import Nav from './components/Nav/Nav.tsx'
import Home from './components/Home/Home.tsx'
import Settings from './components/Settings/Settings.tsx'
import Stats from './components/Statistics/Statistics.tsx'
import { ModelProvider } from './context/modelContext.tsx'
import Login from './components/Login.tsx'
import Register from './components/Register.tsx'
import { AuthProvider } from './context/authContext.tsx'

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
    <AuthProvider>
    <ModelProvider>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Home />} /> 
          <Route path="settings" element={<Settings />} />
          <Route path="stats" element={<Stats />} />
          <Route path="login" element={<Login />} />
          <Route path="register" element={<Register />} />
        </Route>
      </Routes>
    </BrowserRouter>
   </ModelProvider>
  </AuthProvider>
  )
}

export default App

