import { useState, useEffect } from "react";
import { useLocation, Link, useParams, useSearchParams } from "react-router-dom";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { withBase } from "../lib/paths";

function Entry() {
  const { id } = useParams();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const [entry, setEntry] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const currentTool = searchParams.get("tool") || "";
  const backTo = location.state?.backTo || (currentTool ? `/?tool=${currentTool}` : "/");

  useEffect(() => {
    const fetchEntry = async () => {
      try {
        setLoading(true);
        const response = await fetch(withBase(`/api/entries/${id}`));
        if (!response.ok) {
          if (response.status === 404) {
            throw new Error("Release not found");
          }
          throw new Error("Failed to fetch entry");
        }
        const data = await response.json();
        setEntry(data);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchEntry();
  }, [id]);

  if (loading) {
    return (
      <div className="animate-pulse">
        <div className="h-8 w-48 bg-gray-200 rounded mb-4"></div>
        <div className="h-6 w-32 bg-gray-200 rounded mb-6"></div>
        <div className="bg-surface border border-border rounded-lg p-6">
          <div className="h-4 w-full bg-gray-200 rounded mb-3"></div>
          <div className="h-4 w-3/4 bg-gray-200 rounded mb-3"></div>
          <div className="h-4 w-full bg-gray-200 rounded"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-12">
        <p className="text-error mb-4">{error}</p>
        <Link to={backTo} className="text-accent hover:text-accent-hover">
          Back to home
        </Link>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6">
        <Link to={backTo} className="text-accent hover:text-accent-hover text-sm">
          ← Back to releases
        </Link>
      </div>

      <div className="bg-surface border border-border rounded-lg p-6 mb-6">
        <div className="flex items-start justify-between gap-4 mb-4">
          <div>
            <div className="flex items-center gap-2 mb-2">
              <Link
                to={`/tools/${entry.tool_slug}`}
                className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-accent/10 text-accent hover:bg-accent/20"
              >
                {entry.tool_name}
              </Link>
              {entry.is_prerelease === 1 && (
                <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-yellow-100 text-yellow-800">
                  pre-release
                </span>
              )}
            </div>
            <h1 className="text-2xl font-bold text-text">
              {entry.version ? (
                <span className="font-mono">{entry.version}</span>
              ) : (
                <span>{entry.title}</span>
              )}
            </h1>
            {entry.version && entry.title && entry.title !== entry.version && (
              <p className="text-lg text-muted mt-1">{entry.title}</p>
            )}
          </div>
        </div>

        <div className="flex items-center gap-4 text-sm text-muted">
          <span>
            Published{" "}
            {new Date(entry.published_at).toLocaleDateString("en-US", {
              year: "numeric",
              month: "long",
              day: "numeric",
            })}
          </span>
          <a
            href={entry.url}
            target="_blank"
            rel="noopener noreferrer"
            className="text-accent hover:text-accent-hover"
          >
            View original source →
          </a>
        </div>
      </div>

      <div className="bg-surface border border-border rounded-lg p-6">
        <article className="markdown-body">
          <ReactMarkdown remarkPlugins={[remarkGfm]}>
            {entry.body_md || "No release notes available."}
          </ReactMarkdown>
        </article>
      </div>
    </div>
  );
}

export default Entry;
