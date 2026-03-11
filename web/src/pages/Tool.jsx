import { useState, useEffect, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import ReleaseCard from '../components/ReleaseCard'
import LoadingSkeleton from '../components/LoadingSkeleton'

function Tool() {
  const { slug } = useParams()
  const [tool, setTool] = useState(null)
  const [entries, setEntries] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [hasMore, setHasMore] = useState(false)
  const [cursor, setCursor] = useState(null)
  const [loadingMore, setLoadingMore] = useState(false)

  const fetchEntries = useCallback(async (cursorParam = null, append = false) => {
    try {
      if (append) {
        setLoadingMore(true)
      } else {
        setLoading(true)
        setEntries([])
      }
      setError(null)

      let url = `/api/tools/${slug}/entries?limit=20`
      if (cursorParam) url += `&cursor=${cursorParam}`

      const response = await fetch(url)
      if (!response.ok) {
        if (response.status === 404) {
          throw new Error('Tool not found')
        }
        throw new Error('Failed to fetch tool')
      }

      const data = await response.json()
      setTool(data.tool)

      if (append) {
        setEntries(prev => [...prev, ...data.entries])
      } else {
        setEntries(data.entries)
      }
      setHasMore(data.hasMore)
      if (data.nextCursor) {
        setCursor(data.nextCursor)
      }
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
      setLoadingMore(false)
    }
  }, [slug])

  useEffect(() => {
    setCursor(null)
    fetchEntries()
  }, [slug, fetchEntries])

  const loadMore = () => {
    if (cursor && hasMore && !loadingMore) {
      fetchEntries(cursor, true)
    }
  }

  if (loading) {
    return <LoadingSkeleton />
  }

  if (error) {
    return (
      <div className="text-center py-12">
        <p className="text-error mb-4">{error}</p>
        <Link to="/" className="text-accent hover:text-accent-hover">
          Back to home
        </Link>
      </div>
    )
  }

  return (
    <div>
      <div className="mb-6">
        <Link to="/" className="text-accent hover:text-accent-hover text-sm">
          ← Back to all releases
        </Link>
      </div>

      <div className="bg-surface border border-border rounded-lg p-6 mb-6">
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-2xl font-bold text-text mb-2">{tool.name}</h1>
            <div className="flex items-center gap-4 text-sm">
              <a
                href={`https://github.com/${tool.source_repo}`}
                target="_blank"
                rel="noopener noreferrer"
                className="text-accent hover:text-accent-hover"
              >
                GitHub Repository →
              </a>
              {tool.homepage && (
                <a
                  href={tool.homepage}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-accent hover:text-accent-hover"
                >
                  Homepage →
                </a>
              )}
            </div>
          </div>
          <a
            href={`/rss/${slug}`}
            target="_blank"
            rel="noopener noreferrer"
            className="px-3 py-1.5 text-sm bg-gray-100 text-muted hover:bg-gray-200 rounded-lg transition-colors flex items-center gap-1"
          >
            <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
              <path d="M5 3a1 1 0 011 1v12a1 1 0 11-2 0V4a1 1 0 011-1zm3.293 4.293a1 1 0 011.414 0L12 9.586l2.293-2.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z"/>
            </svg>
            RSS Feed
          </a>
        </div>
      </div>

      <h2 className="text-lg font-semibold text-text mb-4">Release History</h2>

      {entries.length === 0 ? (
        <div className="text-center py-8">
          <p className="text-muted">No releases found for this tool.</p>
        </div>
      ) : (
        <>
          <div className="space-y-4">
            {entries.map(entry => (
              <ReleaseCard key={entry.id} entry={entry} />
            ))}
          </div>

          {hasMore && (
            <div className="mt-6 text-center">
              <button
                onClick={loadMore}
                disabled={loadingMore}
                className="px-6 py-2 bg-gray-100 text-text rounded-lg hover:bg-gray-200 transition-colors disabled:opacity-50"
              >
                {loadingMore ? 'Loading...' : 'Load More'}
              </button>
            </div>
          )}
        </>
      )}
    </div>
  )
}

export default Tool