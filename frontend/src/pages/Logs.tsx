// frontend/src/pages/Logs.tsx
import { useEffect, useState, useMemo } from "react";
import { api, type LogEntry } from "../api";
import { Skeleton } from "../components/ui/Skeleton";

export default function Logs() {
  const [rows, setRows] = useState<LogEntry[]>([]);
  const [filter, setFilter] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let on = true;
    const load = async () => {
      setLoading(true);
      try {
        const data = await api.logs(filter);
        if (on) setRows(data ?? []);
      } finally {
        if (on) setLoading(false);
      }
    };
    void load();
    const t = setInterval(load, 5000);
    return () => { on = false; clearInterval(t); };
  }, [filter]);

  const pretty = useMemo(
    () =>
      rows.map((r, i) => ({
        id: i,
        ts: new Date(r.t).toLocaleString(),
        level: r.level,
        msg: r.msg,
      })),
    [rows]
  );

  return (
    <main className="page">
      <h1 className="app-title">Logs</h1>

      <div className="card" style={{ marginBottom: 12 }}>
        <div className="card-title">Filter</div>
        <input
          className="input"
          placeholder="type to filter text…"
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          style={{ maxWidth: 360 }}
        />
      </div>

      <div className="card overflow">
        <div className="card-title">Recent</div>
        {loading ? (
          <Skeleton className="h-40 w-full" />
        ) : pretty.length === 0 ? (
          <div className="empty">No log entries.</div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th style={{ width: 220 }}>Time</th>
                <th style={{ width: 80 }}>Level</th>
                <th>Message</th>
              </tr>
            </thead>
            <tbody>
              {pretty.map((r) => (
                <tr key={r.id}>
                  <td>{r.ts}</td>
                  <td>
                    <span className={`pill ${r.level === "ERROR" ? "pill-bad" : "pill-ok"}`}>{r.level}</span>
                  </td>
                  <td>{r.msg}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </main>
  );
}
