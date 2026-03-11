import { Link } from 'react-router-dom'

function NotFound() {
  return (
    <div className="text-center py-16">
      <h1 className="text-4xl font-bold text-text mb-4">404</h1>
      <p className="text-muted mb-6">The page you're looking for doesn't exist.</p>
      <Link
        to="/"
        className="px-6 py-2 bg-accent text-white rounded-lg hover:bg-accent-hover transition-colors"
      >
        Back to Home
      </Link>
    </div>
  )
}

export default NotFound