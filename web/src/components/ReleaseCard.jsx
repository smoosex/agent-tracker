import { Link } from "react-router-dom";

function formatDate(dateString) {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now - date;
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

  if (diffHours < 1) return "Just now";
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;

  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: date.getFullYear() !== now.getFullYear() ? "numeric" : undefined,
  });
}

function ReleaseCard({ entry }) {
  return (
    <Link
      to={`/entries/${entry.id}`}
      className="block bg-surface border border-border rounded-lg p-4 hover:border-accent hover:shadow-sm transition-all"
    >
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-2">
            <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-accent/10 text-accent">
              {entry.tool_name}
            </span>
            {entry.is_prerelease === 1 && (
              <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-yellow-100 text-yellow-800">
                pre-release
              </span>
            )}
          </div>
          <h3 className="font-semibold text-text mb-1 truncate">
            {entry.version ? (
              <span className="font-mono text-sm">{entry.version}</span>
            ) : (
              <span>{entry.title}</span>
            )}
            {entry.version && entry.title && entry.title !== entry.version && (
              <span className="ml-2 text-muted font-normal">{entry.title}</span>
            )}
          </h3>
          {entry.excerpt && (
            <p className="text-sm text-muted line-clamp-2">{entry.excerpt}</p>
          )}
        </div>
        <span className="text-xs text-muted whitespace-nowrap">
          {formatDate(entry.published_at)}
        </span>
      </div>
    </Link>
  );
}

export default ReleaseCard;
