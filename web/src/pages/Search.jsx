import { useState } from 'react'
import { useSearchParams, Link } from 'react-router-dom'
import ReleaseCard from '../components/ReleaseCard'

function Search() {
  const [searchParams, setSearchParams] = useSearchParams()
  const query = searchParams.get('q') || ''
  const [inputValue, setInputValue] = useState(query)
  const [entries, setEntries] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [searched, setSearched] = useState(false)

  const handleSearch = async (e) => {
    e.preventDefault()
    if (!inputValue.trim()) return

    setSearchParams({ q: inputValue })
    setLoading(true)
    setError(null)

    try {
      const response = await fetch(`/api/search?q=${encodeURIComponent(inputValue)}`)
      if (!response.ok) throw new Error('Search failed')
      const data = await response.json()
      setEntries(data.entries)
      setSearched(true)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <div className="mb-6">
        <Link to="/" className="text-accent hover:text-accent-hover text-sm">
          ← Back to all releases
        </Link>
      </div>

      <div className="bg-surface border border-border rounded-lg p-6 mb-6">
        <h1 className="text-2xl font-bold text-text mb-4">Search Releases</h1>
        <form onSubmit={handleSearch} className="flex gap-2">
          <input
            type="text"
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            placeholder="Search by title or content..."
            className="flex-1 px-4 py-2 border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent"
          />
          <button
            type="submit"
            disabled={loading || !inputValue.trim()}
            className="px-6 py-2 bg-accent text-white rounded-lg hover:bg-accent-hover transition-colors disabled:opacity-50"
          >
            {loading ? 'Searching...' : 'Search'}
          </button>
        </form>
      </div>

      {error && (
        <div className="text-center py-8">
          <p className="text-error">{error}</p>
        </div>
      )}

      {searched && !error && (
        <div>
          <p className="text-muted mb-4">
            {entries.length === 0
              ? `No results found for "${query}"`
              : `Found ${entries.length} result${entries.length === 1 ? '' : 's'} for "${query}"`}
          </p>

          {entries.length > 0 && (
            <div className="space-y-4">
              {entries.map(entry => (
                <ReleaseCard key={entry.id} entry={entry} />
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default Search