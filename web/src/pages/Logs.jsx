import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { withBase } from "../lib/paths";

function formatTime(value) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toLocaleString();
}

function Logs() {
  const [logs, setLogs] = useState([]);
  const [logPath, setLogPath] = useState("");
  const [failures, setFailures] = useState([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState("");

  const fetchData = async (showLoading = true) => {
    try {
      if (showLoading) {
        setLoading(true);
      } else {
        setRefreshing(true);
      }
      setError("");

      const [logsResponse, failuresResponse] = await Promise.all([
        fetch(withBase("/api/logs?limit=200")),
        fetch(withBase("/api/sync/failures?limit=50")),
      ]);

      if (!logsResponse.ok || !failuresResponse.ok) {
        throw new Error("Failed to load logs");
      }

      const logsData = await logsResponse.json();
      const failuresData = await failuresResponse.json();

      setLogs(logsData.lines || []);
      setLogPath(logsData.path || "");
      setFailures(failuresData.failures || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  useEffect(() => {
    fetchData(true);
  }, []);

  return (
    <div>
      <div className="mb-6 flex items-center justify-between gap-4">
        <div>
          <Link to="/" className="text-accent hover:text-accent-hover text-sm">
            ← Back to all releases
          </Link>
          <h1 className="text-2xl font-bold text-text mt-3">Sync Logs</h1>
          <p className="text-muted mt-1">
            Recent sync failures and the latest server log lines
          </p>
        </div>
        <button
          type="button"
          onClick={() => fetchData(false)}
          disabled={loading || refreshing}
          className="px-4 py-2 bg-accent text-white rounded-lg hover:bg-accent-hover transition-colors disabled:opacity-50"
        >
          {refreshing ? "Refreshing..." : "Refresh"}
        </button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-error rounded-lg p-4 mb-6">
          {error}
        </div>
      )}

      <div className="bg-surface border border-border rounded-lg p-6 mb-6">
        <div className="flex items-center justify-between gap-4 mb-4">
          <h2 className="text-lg font-semibold text-text">Recent Failures</h2>
          <span className="text-sm text-muted">{failures.length} items</span>
        </div>

        {loading ? (
          <p className="text-muted">Loading failures...</p>
        ) : failures.length === 0 ? (
          <p className="text-muted">No recent sync failures.</p>
        ) : (
          <div className="space-y-3">
            {failures.map((failure) => (
              <div
                key={failure.id}
                className="rounded-lg border border-border bg-white p-4"
              >
                <div className="flex flex-wrap items-center gap-3 mb-2 text-sm">
                  <span className="font-semibold text-text">{failure.tool_slug}</span>
                  <span className="text-muted">{formatTime(failure.created_at)}</span>
                  <span className="text-muted">
                    {failure.full_sync ? "Full sync" : "Incremental sync"}
                  </span>
                </div>
                <p className="text-sm text-text whitespace-pre-wrap break-words">
                  {failure.error}
                </p>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="bg-surface border border-border rounded-lg p-6">
        <div className="flex items-center justify-between gap-4 mb-4">
          <div>
            <h2 className="text-lg font-semibold text-text">Recent Log Lines</h2>
            {logPath && <p className="text-sm text-muted mt-1">{logPath}</p>}
          </div>
          <span className="text-sm text-muted">{logs.length} lines</span>
        </div>

        {loading ? (
          <p className="text-muted">Loading logs...</p>
        ) : logs.length === 0 ? (
          <p className="text-muted">No log lines found.</p>
        ) : (
          <pre className="bg-slate-950 text-slate-100 text-xs rounded-lg p-4 overflow-x-auto whitespace-pre-wrap break-words">
            {logs.join("\n")}
          </pre>
        )}
      </div>
    </div>
  );
}

export default Logs;
