import { useState, useEffect, useCallback } from "react";
import { useSearchParams } from "react-router-dom";
import ReleaseCard from "../components/ReleaseCard";
import LoadingSkeleton from "../components/LoadingSkeleton";
import { withBase } from "../lib/paths";

function Home() {
  const [searchParams] = useSearchParams();
  const [entries, setEntries] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [hasMore, setHasMore] = useState(false);
  const [cursor, setCursor] = useState(null);
  const [loadingMore, setLoadingMore] = useState(false);

  const tool = searchParams.get("tool") || "";

  const fetchEntries = useCallback(
    async (cursorParam = null, append = false) => {
      try {
        if (append) {
          setLoadingMore(true);
        } else {
          setLoading(true);
          setEntries([]);
        }
        setError(null);

        let url = withBase(`/api/entries?limit=20`);
        if (tool) url += `&tool=${tool}`;
        if (cursorParam) url += `&cursor=${cursorParam}`;

        const response = await fetch(url);
        if (!response.ok) throw new Error("Failed to fetch entries");

        const data = await response.json();

        if (append) {
          setEntries((prev) => [...prev, ...data.entries]);
        } else {
          setEntries(data.entries);
        }
        setHasMore(data.hasMore);
        if (data.nextCursor) {
          setCursor(data.nextCursor);
        }
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
        setLoadingMore(false);
      }
    },
    [tool],
  );

  useEffect(() => {
    setCursor(null);
    fetchEntries();
  }, [tool, fetchEntries]);

  useEffect(() => {
    const handleSync = () => {
      setCursor(null);
      fetchEntries();
    };

    window.addEventListener("tracker:sync", handleSync);
    return () => window.removeEventListener("tracker:sync", handleSync);
  }, [fetchEntries]);

  const loadMore = () => {
    if (cursor && hasMore && !loadingMore) {
      fetchEntries(cursor, true);
    }
  };

  return (
    <div>
      <div className="mb-6">
        <div>
          <h1 className="text-2xl font-bold text-text">
            {tool ? `Releases for ${tool}` : "Latest Releases"}
          </h1>
          <p className="text-muted mt-1">
            Tracking changelogs and releases from AI coding tools
          </p>
        </div>
      </div>

      {loading ? (
        <LoadingSkeleton />
      ) : error ? (
        <div className="text-center py-12">
          <p className="text-error mb-4">{error}</p>
          <button
            onClick={() => fetchEntries()}
            className="px-4 py-2 bg-accent text-white rounded-lg hover:bg-accent-hover transition-colors"
          >
            Try Again
          </button>
        </div>
      ) : entries.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-muted">No releases found</p>
          {tool && (
            <p className="text-sm text-muted mt-2">
              No releases for this tool yet. Try syncing the data.
            </p>
          )}
        </div>
      ) : (
        <>
          <div className="space-y-4">
            {entries.map((entry) => (
              <ReleaseCard key={entry.id} entry={entry} currentTool={tool} />
            ))}
          </div>

          {hasMore && (
            <div className="mt-6 text-center">
              <button
                onClick={loadMore}
                disabled={loadingMore}
                className="px-6 py-2 bg-gray-100 text-text rounded-lg hover:bg-gray-200 transition-colors disabled:opacity-50"
              >
                {loadingMore ? "Loading..." : "Load More"}
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}

export default Home;
