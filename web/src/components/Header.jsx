import { Link, useSearchParams, useNavigate } from 'react-router-dom'
import { useState } from 'react'

const TOOLS = [
  { slug: 'claude-code', name: 'Claude Code' },
  { slug: 'codex', name: 'Codex' },
  { slug: 'gemini-cli', name: 'Gemini CLI' },
  { slug: 'opencode', name: 'OpenCode' },
]

function Header() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const currentTool = searchParams.get('tool') || ''
  const [searchQuery, setSearchQuery] = useState('')

  const handleSearch = (e) => {
    e.preventDefault()
    if (searchQuery.trim()) {
      navigate(`/search?q=${encodeURIComponent(searchQuery.trim())}`)
      setSearchQuery('')
    }
  }

  return (
    <header className="bg-surface border-b border-border sticky top-0 z-10">
      <div className="max-w-4xl mx-auto px-4 py-4">
        <div className="flex items-center justify-between mb-4">
          <Link to="/" className="text-xl font-bold text-text hover:text-accent transition-colors">
            Agent Tracker
          </Link>
          <div className="flex items-center gap-3">
            <form onSubmit={handleSearch} className="flex items-center">
              <div className="relative">
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Search releases..."
                  className="w-48 px-3 py-1.5 pl-8 text-sm border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent"
                />
                <svg
                  className="w-4 h-4 text-muted absolute left-2.5 top-1/2 -translate-y-1/2"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
              </div>
            </form>
            <a
              href="/rss/all"
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm text-muted hover:text-accent transition-colors flex items-center gap-1"
              title="RSS Feed"
            >
              <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                <path d="M5 3a1 1 0 011 1v12a1 1 0 11-2 0V4a1 1 0 011-1zm3.293 4.293a1 1 0 011.414 0L12 9.586l2.293-2.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z"/>
              </svg>
              RSS
            </a>
          </div>
        </div>

        <div className="flex flex-wrap gap-2 items-center">
          <Link
            to="/"
            className={`px-3 py-1.5 rounded-full text-sm font-medium transition-colors ${
              currentTool === '' ? 'bg-accent text-white' : 'bg-gray-100 text-muted hover:bg-gray-200'
            }`}
          >
            All
          </Link>
          {TOOLS.map(tool => (
            <Link
              key={tool.slug}
              to={`/?tool=${tool.slug}`}
              className={`px-3 py-1.5 rounded-full text-sm font-medium transition-colors ${
                currentTool === tool.slug
                  ? 'bg-accent text-white'
                  : 'bg-gray-100 text-muted hover:bg-gray-200'
              }`}
            >
              {tool.name}
            </Link>
          ))}
        </div>
      </div>
    </header>
  )
}

export default Header