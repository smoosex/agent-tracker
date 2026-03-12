import {
  Link,
  matchPath,
  useLocation,
  useNavigate,
  useSearchParams,
} from "react-router-dom";
import { useCallback, useEffect, useState } from "react";
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
  const [cooldownUntil, setCooldownUntil] = useState(null);
  const [cooldownRemaining, setCooldownRemaining] = useState(0);
  const [showRefreshTooltip, setShowRefreshTooltip] = useState(false);
  const [syncError, setSyncError] = useState("");
  const [syncMessage, setSyncMessage] = useState("");

  const applySyncStatus = useCallback((data) => {
    setSyncRunning(Boolean(data.running));
    setCooldownUntil(data.cooldown_until || null);
    setCooldownRemaining(data.cooldown_remaining_seconds || 0);
  }, []);

  const fetchSyncStatus = useCallback(async () => {
    try {
      const response = await fetch(withBase("/api/sync/status"));
      if (!response.ok) return;
      const data = await response.json();
      applySyncStatus(data);
    } catch {}
  }, [applySyncStatus]);

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
    fetchSyncStatus();

    const eventSource = new EventSource(withBase("/api/sync/events"));

    const handleSyncStatus = (event) => {
      try {
        const data = JSON.parse(event.data);
        applySyncStatus(data);
      } catch {}
    };

    eventSource.addEventListener("sync-status", handleSyncStatus);
    eventSource.onerror = () => {
      fetchSyncStatus();
    };
    return () => {
      eventSource.removeEventListener("sync-status", handleSyncStatus);
      eventSource.onerror = null;
      eventSource.close();
    };
  }, [applySyncStatus, fetchSyncStatus]);

  useEffect(() => {
    if (!cooldownUntil) {
      setCooldownRemaining(0);
      return;
    }

    const updateCooldown = () => {
      const remaining = Math.max(
        0,
        Math.ceil((new Date(cooldownUntil).getTime() - Date.now()) / 1000),
      );
      setCooldownRemaining(remaining);
      if (remaining === 0) {
        setCooldownUntil(null);
      }
    };

    updateCooldown();
    const intervalId = window.setInterval(updateCooldown, 1000);
    return () => window.clearInterval(intervalId);
  }, [cooldownUntil]);

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
    if (syncRunning || syncing || cooldownRemaining > 0) return;

    try {
      setSyncing(true);
      setSyncError("");
      setSyncMessage("");

      const response = await fetch(withBase("/api/sync"), { method: "POST" });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) {
        if (response.status === 409) {
          applySyncStatus(data);
          setSyncMessage("Sync already in progress");
          return;
        }
        if (response.status === 429) {
          applySyncStatus(data);
          setSyncMessage(
            data.error ||
              `Please wait ${data.cooldown_remaining_seconds || 10}s before syncing again`,
          );
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
      fetchSyncStatus();
    }
  };

  const refreshTooltip =
    syncing || syncRunning
      ? "Syncing..."
      : cooldownRemaining > 0
        ? `Please wait ${cooldownRemaining}s before refreshing again`
        : "";

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
                <div
                  className="relative flex"
                  onMouseEnter={() => setShowRefreshTooltip(true)}
                  onMouseLeave={() => setShowRefreshTooltip(false)}
                  onFocus={() => setShowRefreshTooltip(true)}
                  onBlur={() => setShowRefreshTooltip(false)}
                >
                  <button
                    type="button"
                    onClick={handleSync}
                    disabled={syncing || syncRunning || cooldownRemaining > 0}
                    className="h-8 w-8 shrink-0 rounded-lg border border-border bg-white text-muted transition-colors hover:text-accent hover:border-accent disabled:cursor-not-allowed disabled:opacity-50"
                    aria-label={
                      syncing || syncRunning
                        ? "Syncing data"
                        : cooldownRemaining > 0
                          ? `Refresh available in ${cooldownRemaining} seconds`
                          : "Refresh data"
                    }
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
                  {showRefreshTooltip && refreshTooltip && (
                    <div className="pointer-events-none absolute left-1/2 top-full z-20 mt-2 -translate-x-1/2 whitespace-nowrap rounded-md border border-border bg-slate-950 px-2.5 py-1.5 text-xs text-white shadow-lg">
                      {refreshTooltip}
                    </div>
                  )}
                </div>
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
            <Link
              to="/logs"
              className="text-sm text-muted hover:text-accent transition-colors"
            >
              Logs
            </Link>
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
