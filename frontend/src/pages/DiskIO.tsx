import { useEffect, useMemo, useState } from "react";
import { api, type DiskPoint, type DiskIOPoint } from "../api";
import { Sparkline } from "../components/ui/Sparkline";
import { Skeleton } from "../components/ui/Skeleton";


type DiskResp = { range: string; mounts: { mount: string; points: DiskPoint[] }[] };
type IOResp   = { range: string; points: DiskIOPoint[] };

export default function DiskIOPage() {
  const [range, setRange] = useState("24h");
  const [ioRange, setIORange] = useState("1h");

  const [mounts, setMounts] = useState<DiskResp["mounts"]>([]);
  const [selected, setSelected] = useState("");

  const [loading, setLoading] = useState(true);
  const [ioLoading, setIOLoading] = useState(true);

  const [ioPts, setIOPts] = useState<IOResp["points"]>([]);

  useEffect(() => {
    let on = true;
    setLoading(true);
    api.disk(range)
      .then((d) => {
        if (!on || !d) return;
        setMounts(d.mounts || []);
        if (!selected && d.mounts && d.mounts.length) setSelected(d.mounts[0].mount);
      })
      .finally(() => on && setLoading(false));
    return () => { on = false; };
  }, [range, selected]);

  useEffect(() => {
    let on = true;
    setIOLoading(true);
    api.diskio(ioRange)
      .then((d) => { if (on && d) setIOPts(d.points || []); })
      .finally(() => on && setIOLoading(false));
    return () => { on = false; };
  }, [ioRange]);

  const current = useMemo(() => mounts.find((m) => m.mount === selected), [mounts, selected]);

  const seriesPct = useMemo(() => (current?.points ?? []).map((p) => p.usedPct), [current]);
  const latestPct = useMemo(() => {
    const pts = current?.points ?? [];
    return pts.length ? pts[pts.length - 1].usedPct : 0;
  }, [current]);
  const latestGB = useMemo(() => {
    const pts = current?.points ?? [];
    return pts.length ? { used: pts[pts.length - 1].usedGB, total: pts[pts.length - 1].totalGB } : { used: 0, total: 0 };
  }, [current]);

  const ioRead = useMemo(() => ioPts.map((p) => p.readMBs), [ioPts]);
  const ioWrite = useMemo(() => ioPts.map((p) => p.writeMBs), [ioPts]);

  return (
    <main className="page">
      <h1 className="app-title">Disk & IO</h1>

      <div className="panel">
        <div className="panel-title">Partitions</div>
        {loading ? (
          <Skeleton className="h-40 w-full" />
        ) : mounts.length === 0 ? (
          <div className="empty">No partitions reported yet.</div>
        ) : (
          <>
            <div style={{ display: "flex", gap: 12, alignItems: "center", marginBottom: 8 }}>
              <select className="input" value={selected} onChange={(e) => setSelected(e.target.value)} style={{ width: 220 }}>
                {mounts.map((m) => (
                  <option key={m.mount} value={m.mount}>
                    {m.mount}
                  </option>
                ))}
              </select>
              <span className="small muted">
                Latest: <b>{latestPct.toFixed(1)}%</b> · <b>{latestGB.used.toFixed(1)}</b> GB / <b>{latestGB.total.toFixed(1)}</b> GB
              </span>
            </div>

            <div className="panel" style={{ marginTop: 8 }}>
              <div className="panel-title">Disk usage trend ({range})</div>
              <div style={{ marginBottom: 8 }}>
                <select className="input" value={range} onChange={(e) => setRange(e.target.value)} style={{ width: 160 }}>
                  <option value="1h">1h</option>
                  <option value="6h">6h</option>
                  <option value="24h">24h</option>
                </select>
              </div>
              {seriesPct.length < 2 ? <Skeleton className="h-40 w-full" /> : <Sparkline data={seriesPct} height={160} />}
              {seriesPct.length >= 2 && (
                <div className="chart-legend small">
                  <span className="muted">min {Math.min(...seriesPct).toFixed(1)}%</span>
                  <span className="muted">max {Math.max(...seriesPct).toFixed(1)}%</span>
                </div>
              )}
            </div>
          </>
        )}
      </div>

      <div className="panel" style={{ marginTop: 16 }}>
        <div className="panel-title">Disk IO (MB/s)</div>
        <div style={{ display: "flex", gap: 16, alignItems: "center", marginBottom: 8 }}>
          <select className="input" value={ioRange} onChange={(e) => setIORange(e.target.value)} style={{ width: 160 }}>
            <option value="30m">30m</option>
            <option value="1h">1h</option>
            <option value="6h">6h</option>
          </select>
        </div>
        {ioLoading ? (
          <Skeleton className="h-40 w-full" />
        ) : (
          <>
            <div className="panel" style={{ marginBottom: 12 }}>
              <div className="panel-title">Read</div>
              {ioRead.length < 2 ? <Skeleton className="h-40 w-full" /> : <Sparkline data={ioRead} height={120} />}
            </div>
            <div className="panel">
              <div className="panel-title">Write</div>
              {ioWrite.length < 2 ? <Skeleton className="h-40 w-full" /> : <Sparkline data={ioWrite} height={120} />}
            </div>
          </>
        )}
      </div>
    </main>
  );
}
