import {
  Link,
  matchPath,
  useLocation,
  useNavigate,
  useSearchParams,
} from "react-router-dom";
import { useEffect, useState } from "react";
import { withBase } from "../lib/paths";

function Header() {
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const toolMatch = matchPath("/tools/:slug", location.pathname);
  const currentTool = searchParams.get("tool") || toolMatch?.params.slug || "";
  const [searchQuery, setSearchQuery] = useState("");
  const [tools, setTools] = useState([]);
  const [syncing, setSyncing] = useState(false);
  const [syncRunning, setSyncRunning] = useState(false);
  const [syncError, setSyncError] = useState("");
  const [syncMessage, setSyncMessage] = useState("");

  useEffect(() => {
    let cancelled = false;

    const fetchTools = async () => {
      try {
        const response = await fetch(withBase("/api/tools"));
        if (!response.ok) return;
        const data = await response.json();
        if (!cancelled) {
          setTools(data);
        }
      } catch {}
    };

    fetchTools();

    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    const eventSource = new EventSource(withBase("/api/sync/events"));

    const handleSyncStatus = (event) => {
      try {
        const data = JSON.parse(event.data);
        setSyncRunning(Boolean(data.running));
      } catch {}
    };

    eventSource.addEventListener("sync-status", handleSyncStatus);
    return () => {
      eventSource.removeEventListener("sync-status", handleSyncStatus);
      eventSource.close();
    };
  }, []);

  useEffect(() => {
    if (!syncError && !syncMessage) return;
    const timer = window.setTimeout(() => {
      setSyncError("");
      setSyncMessage("");
    }, 3000);
    return () => window.clearTimeout(timer);
  }, [syncError, syncMessage]);

  const handleSearch = (e) => {
    e.preventDefault();
    if (searchQuery.trim()) {
      navigate(`/search?q=${encodeURIComponent(searchQuery.trim())}`);
      setSearchQuery("");
    }
  };

  const handleSync = async () => {
    if (syncRunning || syncing) return;

    try {
      setSyncing(true);
      setSyncRunning(true);
      setSyncError("");
      setSyncMessage("");

      const response = await fetch(withBase("/api/sync"), { method: "POST" });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) {
        if (response.status === 409) {
          setSyncMessage("Sync already in progress");
          return;
        }
        throw new Error(data.error || "Failed to sync data");
      }

      setSyncMessage(data.message || "Data refreshed");
      window.dispatchEvent(new Event("tracker:sync"));
    } catch (err) {
      setSyncError(err.message);
    } finally {
      setSyncing(false);
    }
  };

  return (
    <header className="bg-surface border-b border-border sticky top-0 z-10">
      <div className="max-w-4xl mx-auto px-4 py-4">
        <div className="flex items-center justify-between mb-4">
          <Link
            to="/"
            className="text-xl font-bold text-text hover:text-accent transition-colors"
          >
            Agent Tracker
          </Link>
          <div className="flex items-center gap-3">
            <form onSubmit={handleSearch} className="flex items-center">
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={handleSync}
                  disabled={syncing || syncRunning}
                  className="h-8 w-8 shrink-0 rounded-lg border border-border bg-white text-muted transition-colors hover:text-accent hover:border-accent disabled:cursor-not-allowed disabled:opacity-50"
                  title={syncing || syncRunning ? "Syncing..." : "Refresh data"}
                  aria-label={syncing || syncRunning ? "Syncing data" : "Refresh data"}
                >
                  <svg
                    className={`mx-auto h-4 w-4 ${syncing || syncRunning ? "animate-spin" : ""}`}
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={1.8}
                      d="M21 12a9 9 0 0 0-15.36-6.36L3 8"
                    />
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={1.8}
                      d="M3 3v5h5"
                    />
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={1.8}
                      d="M3 12a9 9 0 0 0 15.36 6.36L21 16"
                    />
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={1.8}
                      d="M16 16h5v5"
                    />
                  </svg>
                </button>
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
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                    />
                  </svg>
                </div>
              </div>
            </form>
            <a
              href={withBase("/rss/all")}
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm text-muted hover:text-accent transition-colors flex items-center gap-1"
              title="RSS Feed"
            >
              <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                <path d="M5 3a1 1 0 011 1v12a1 1 0 11-2 0V4a1 1 0 011-1zm3.293 4.293a1 1 0 011.414 0L12 9.586l2.293-2.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z" />
              </svg>
              RSS
            </a>
          </div>
        </div>

        {(syncError || syncMessage) && (
          <p className={`mb-3 text-sm ${syncError ? "text-error" : "text-muted"}`}>
            {syncError || syncMessage}
          </p>
        )}

        <div className="flex flex-wrap gap-2 items-center">
          <Link
            to="/"
            className={`px-3 py-1.5 rounded-full text-sm font-medium transition-colors ${
              currentTool === ""
                ? "bg-accent text-white"
                : "bg-gray-100 text-muted hover:bg-gray-200"
            }`}
          >
            All
          </Link>
          {tools.map((tool) => (
            <Link
              key={tool.slug}
              to={`/?tool=${tool.slug}`}
              className={`px-3 py-1.5 rounded-full text-sm font-medium transition-colors ${
                currentTool === tool.slug
                  ? "bg-accent text-white"
                  : "bg-gray-100 text-muted hover:bg-gray-200"
              }`}
            >
              {tool.name}
            </Link>
          ))}
        </div>
      </div>
    </header>
  );
}

export default Header;
