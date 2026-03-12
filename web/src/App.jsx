import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import Header from './components/Header'
import Home from './pages/Home'
import Entry from './pages/Entry'
import Tool from './pages/Tool'
import Search from './pages/Search'
import Logs from './pages/Logs'
import NotFound from './pages/NotFound'
import { basePath } from './lib/paths'

function App() {
  return (
    <Router basename={basePath || "/"}>
      <div className="min-h-screen bg-background">
        <Header />
        <main className="max-w-4xl mx-auto px-4 py-8">
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/entries/:id" element={<Entry />} />
            <Route path="/tools/:slug" element={<Tool />} />
            <Route path="/search" element={<Search />} />
            <Route path="/logs" element={<Logs />} />
            <Route path="*" element={<NotFound />} />
          </Routes>
        </main>
      </div>
    </Router>
  )
}

export default App
