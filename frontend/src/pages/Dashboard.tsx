import { useEffect, useMemo, useState, useCallback } from "react";
import { api, type Task, type CPUPoint, type DiskPoint, type NetPoint } from "../api";
import { Sparkline } from "../components/ui/Sparkline";
import { StatCard } from "../components/ui/StatCard";
import { Skeleton } from "../components/ui/Skeleton";

type CPUResp  = { range: string; points: CPUPoint[]; avg: number; p95: number };
type DiskResp = { range: string; mounts: { mount: string; points: DiskPoint[] }[] };


const CpuIcon  = (<svg width="20" height="20" viewBox="0 0 24 24"><rect x="6" y="6" width="12" height="12" rx="2" fill="currentColor" opacity="0.15"/><rect x="8" y="8" width="8" height="8" rx="1.5" stroke="currentColor" fill="none"/></svg>);
const P95Icon  = (<svg width="20" height="20" viewBox="0 0 24 24"><path d="M5 16l4-4 4 4 6-6" stroke="currentColor" strokeWidth="2" fill="none"/></svg>);
const DiskIcon = (<svg width="20" height="20" viewBox="0 0 24 24"><circle cx="12" cy="12" r="9" stroke="currentColor" fill="none"/><circle cx="12" cy="12" r="2" fill="currentColor"/></svg>);
const MemIcon  = (<svg width="20" height="20" viewBox="0 0 24 24"><rect x="4" y="7" width="16" height="10" rx="2" stroke="currentColor" fill="none"/><rect x="7" y="10" width="3" height="4" fill="currentColor"/><rect x="14" y="10" width="3" height="4" fill="currentColor"/></svg>);


function useHealth() {
  const [ok, setOk] = useState(false);
  const [last, setLast] = useState("");
  const [now, setNow] = useState(new Date());
  useEffect(() => {
    let on = true;
    const poll = async () => {
      try {
        const h = await api.health();
        if (!on || !h) return;
        setLast(h.lastCollectorAt);
        setOk(Date.now() - new Date(h.lastCollectorAt).getTime() < 120_000);
      } catch {
        if (on) setOk(false);
      }
    };
    poll();
    const t1 = setInterval(poll, 5000);
    const t2 = setInterval(() => setNow(new Date()), 1000);
    return () => { on = false; clearInterval(t1); clearInterval(t2); };
  }, []);
  return { ok, last, now };
}



function useCpu(range = "1h") {
  const [data,setData]=useState<CPUResp|null>(null);
  const [loading,setLoading]=useState(true);
  const [err,setErr]=useState<string|null>(null);
  useEffect(()=>{ let on=true; setLoading(true);
    api.cpu(range).then(d=>{ if(on){ setData(d??null); setErr(null);} })
      .catch(e=>{ if(on) setErr(e instanceof Error? e.message:String(e)); })
      .finally(()=>{ if(on) setLoading(false); });
    return ()=>{ on=false; };
  },[range]);
  const series = useMemo(()=> (data?.points ?? []).map(p=>p.v), [data]);
  return { data, series, loading, err };
}

function useDisk(range = "24h") {
  const [data,setData]=useState<DiskResp|null>(null);
  const [loading,setLoading]=useState(true);
  const [err,setErr]=useState<string|null>(null);
  useEffect(()=>{ let on=true; setLoading(true);
    api.disk(range).then(d=>{ if(on){ setData(d??null); setErr(null);} })
      .catch(e=>{ if(on) setErr(e instanceof Error? e.message:String(e)); })
      .finally(()=>{ if(on) setLoading(false); });
    return ()=>{ on=false; };
  },[range]);
  const primary = data?.mounts?.[0];
  const series = useMemo(()=> (primary?.points ?? []).map(p => p.usedPct), [primary]);
  const latestPct = useMemo(()=> primary?.points?.length ? primary.points[primary.points.length-1].usedPct : 0, [primary]);
  return { series, latestPct, loading, err };
}

function useMem(range="1h"){
  const [latest,setLatest]=useState(0);
  const [loading,setLoading]=useState(true);
  const [err,setErr]=useState<string|null>(null);
  useEffect(()=>{ let on=true; setLoading(true);
    api.mem(range).then(d=>{ if(on && d){ setLatest(d.latest ?? 0); setErr(null);} })
      .catch(e=>{ if(on) setErr(e instanceof Error? e.message:String(e)); })
      .finally(()=>{ if(on) setLoading(false); });
    return ()=>{ on=false; };
  },[range]);
  return { latest, loading, err };
}

function useNetSplit(range="1h"){
  const [rx,setRx]=useState<number[]>([]);
  const [tx,setTx]=useState<number[]>([]);
  const [loading,setLoading]=useState(true);
  const [err,setErr]=useState<string|null>(null);
  useEffect(()=>{ let on=true; setLoading(true);
    api.net(range).then(d=>{
      if(!on || !d) return;
      let rxVals = (d.points as NetPoint[] ?? []).map(p => p.rxKBs || 0);
      let txVals = (d.points as NetPoint[] ?? []).map(p => p.txKBs || 0);
      if (rxVals.length < 2) rxVals = [0.05, 0.05];
      if (txVals.length < 2) txVals = [0.05, 0.05];
      if (rxVals.every(v => v === 0)) rxVals = rxVals.map(() => 0.05);
      if (txVals.every(v => v === 0)) txVals = txVals.map(() => 0.05);
      setRx(rxVals); setTx(txVals); setErr(null);
    }).catch(e=>{ if(on) setErr(e instanceof Error? e.message:String(e)); })
      .finally(()=>{ if(on) setLoading(false); });
    return ()=>{ on=false; };
  },[range]);
  return { rx, tx, loading, err };
}



export default function Dashboard() {
  const { ok, last, now } = useHealth();
  const { data:cpu, series:cpuSeries,  loading:cpuLoading,  err:cpuErr } = useCpu("1h");
  const { series:diskSeries, latestPct, loading:diskLoading, err:diskErr } = useDisk("24h");
  const { latest:memLatest, loading:memLoading, err:memErr } = useMem("1h");
  const { rx:netRx, tx:netTx, loading:netLoading } = useNetSplit("1h");

  
  const [tasks,setTasks]=useState<Task[]>([]);
  const [tLoading,setTLoading]=useState(true);
  const [busyId,setBusyId]=useState<string|null>(null);
  const [name,setName]=useState("Clear Temp");
  const [every,setEvery]=useState(60);
  const loadTasks = useCallback(async()=>{ setTLoading(true); try{ setTasks((await api.tasks.list())??[]);} finally{ setTLoading(false);} },[]);
  useEffect(()=>{ void loadTasks(); },[loadTasks]);
  async function addTask(e:React.FormEvent){ e.preventDefault(); if(!name.trim()||every<=0) return; await api.tasks.create(name.trim(),every); setName("Clear Temp"); setEvery(60); await loadTasks(); }
  async function runNow(id:string){ setBusyId(id); try{ await api.tasks.runNow(id); await loadTasks(); } finally{ setBusyId(null);} }
  async function removeTask(id:string){ setBusyId(id); try{ await api.tasks.del(id); await loadTasks(); } finally{ setBusyId(null);} }

  
  const dlAvg = useMemo(()=> netRx.length ? netRx.reduce((a,b)=>a+b,0)/netRx.length : 0, [netRx]);
  const dlPeak= useMemo(()=> netRx.length ? Math.max(...netRx) : 0, [netRx]);
  const ulAvg = useMemo(()=> netTx.length ? netTx.reduce((a,b)=>a+b,0)/netTx.length : 0, [netTx]);
  const ulPeak= useMemo(()=> netTx.length ? Math.max(...netTx) : 0, [netTx]);

  
  const topTasks = useMemo(() => {
    const copy = [...tasks];
    copy.sort((a,b)=>{
      const aT = a.lastRun ? new Date(a.lastRun).getTime() : 0;
      const bT = b.lastRun ? new Date(b.lastRun).getTime() : 0;
      return bT - aT;
    });
    return copy.slice(0,3);
  }, [tasks]);

  return (
    <main className="page">
      {}
      <div className="statusbar">
        <div className="status-left">
          <span className={`dot ${ok ? "dot-ok" : "dot-bad"}`} />
          <span className="status-strong">{ok ? "Sampling active" : "Sampling paused"}</span>
          <span className="muted">Last updated {last ? new Date(last).toLocaleTimeString() : "—"}</span>
        </div>
        <div className="status-right">{now.toLocaleTimeString()}</div>
      </div>

      {}
      <section className="kpis">
        <StatCard icon={CpuIcon}  label="CPU avg (1h)" value={cpuLoading ? "…" : `${(cpu?.avg ?? 0).toFixed(1)}%`} sub={cpuErr ?? "Average over last hour"} />
        <StatCard icon={P95Icon} label="CPU p95 (1h)" value={cpuLoading ? "…" : `${(cpu?.p95 ?? 0).toFixed(1)}%`} sub="95th percentile" />
        <StatCard icon={MemIcon} label="Memory used" value={memLoading ? "…" : `${memLatest.toFixed(1)}%`} sub={memErr ?? "Latest sample"} />
        <StatCard icon={DiskIcon} label="Disk usage" value={diskLoading ? "…" : `${latestPct.toFixed(1)}%`} sub={diskErr ?? "Primary mount"} />
      </section>

      {}
      <section className="panels">
        <div className="panel">
          <h2 className="panel-title">CPU trend (1h)</h2>
          {cpuLoading ? <Skeleton className="h-40 w-full"/> : <Sparkline data={cpuSeries} height={160} />}
        </div>
        <div className="panel">
          <h2 className="panel-title">Disk usage trend (24h)</h2>
          {diskLoading ? <Skeleton className="h-40 w-full"/> : <Sparkline data={diskSeries} height={160} />}
        </div>
        <div className="panel">
          <h2 className="panel-title">Summary</h2>
          <div className="small muted">
            <div>CPU avg <b>{cpuLoading ? "…" : `${(cpu?.avg ?? 0).toFixed(1)}%`}</b> · p95 <b>{cpuLoading ? "…" : `${(cpu?.p95 ?? 0).toFixed(1)}%`}</b></div>
            <div>Memory used <b>{memLoading ? "…" : `${memLatest.toFixed(1)}%`}</b></div>
            <div>Disk used <b>{diskLoading ? "…" : `${latestPct.toFixed(1)}%`}</b></div>
          </div>
        </div>
      </section>

      {}
      <section className="panels">
        <div className="panel">
          <h2 className="panel-title">Download (KB/s)</h2>
          {netLoading ? <Skeleton className="h-40 w-full" /> : <Sparkline data={netRx} height={160} />}
          <div className="chart-legend small"><span className="muted">avg {(dlAvg/1024).toFixed(2)} MB/s</span><span className="muted">peak {(dlPeak/1024).toFixed(2)} MB/s</span></div>
        </div>
        <div className="panel">
          <h2 className="panel-title">Upload (KB/s)</h2>
          {netLoading ? <Skeleton className="h-40 w-full" /> : <Sparkline data={netTx} height={160} />}
          <div className="chart-legend small"><span className="muted">avg {(ulAvg/1024).toFixed(2)} MB/s</span><span className="muted">peak {(ulPeak/1024).toFixed(2)} MB/s</span></div>
        </div>
        <div className="panel">
          <h2 className="panel-title">Insights</h2>
          <div className="small muted">
            <div>{cpu ? (cpu.p95 > 80 ? "High CPU: p95 > 80% in the last hour." : "CPU headroom is fine.") : "…"}</div>
            <div>{latestPct > 90 ? "Disk above 90% used. Consider cleanup." : "Disk usage within healthy range."}</div>
            <div>{(dlPeak+ulPeak) > 0 ? `Net peak ${(Math.max(dlPeak, ulPeak)/1024).toFixed(2)} MB/s; avg ${(Math.max(dlAvg, ulAvg)/1024).toFixed(2)} MB/s.` : "No network samples yet."}</div>
          </div>
        </div>
      </section>

      <section className="panels">
        <div className="panel">
          <h2 className="panel-title">Top tasks</h2>
          {tLoading ? (
            <div className="muted small">Loading…</div>
          ) : topTasks.length === 0 ? (
            <div className="muted small">No tasks yet.</div>
          ) : (
            <table className="table">
              <thead><tr><th>Name</th><th>Every</th><th>Last run</th><th className="right">Actions</th></tr></thead>
              <tbody>
                {topTasks.map(t => (
                  <tr key={t.id}>
                    <td>{t.name}</td>
                    <td>{t.everyMinutes} min</td>
                    <td>{t.lastRun ? new Date(t.lastRun).toLocaleString() : "—"}</td>
                    <td className="right">
                      <button onClick={()=>void runNow(t.id)} disabled={busyId===t.id} className="btn btn-sm btn-primary mr-2">Run Now</button>
                      <button onClick={()=>void removeTask(t.id)} disabled={busyId===t.id} className="btn btn-sm btn-ghost">Delete</button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
        <div className="panel">
          <h2 className="panel-title">Tip</h2>
          <div className="small muted">
            Add a “Flush DNS” task (every 1440 minutes) if you’re debugging weird name-resolution caching on Windows.
          </div>
        </div>
        <div className="panel">
          <h2 className="panel-title">Legend</h2>
          <div className="small muted">
            Network charts use rxKB/s and txKB/s. Bandwidth labels below each chart show avg and peak converted to MB/s.
          </div>
        </div>
      </section>

      {}
      <section className="tasks-grid">
        <div className="card">
          <div className="card-title">Add task</div>
          <form className="task-form" onSubmit={addTask}>
            <input className="input" value={name} onChange={(e)=>setName(e.target.value)} placeholder="Task name" />
            <input className="input input-num" type="number" min={1} value={every} onChange={(e)=>setEvery(Number(e.target.value))} placeholder="Every (min)" />
            <button className="btn btn-primary">Add</button>
          </form>
          <p className="muted small mt-2">Runs OS cleanups based on task name. Try “Clear Temp” or “Flush DNS”.</p>
        </div>

        <div className="card overflow">
          <div className="card-title">Tasks</div>
          {tLoading ? <div className="muted">Loading…</div> : tasks.length === 0 ? <div className="muted">No tasks yet.</div> : (
            <table className="table">
              <thead><tr><th>Name</th><th>Every</th><th>Last run</th><th>Status</th><th className="right">Actions</th></tr></thead>
              <tbody>
              {tasks.map(t=>(
                <tr key={t.id}>
                  <td>{t.name}</td>
                  <td>{t.everyMinutes} min</td>
                  <td>{t.lastRun ? new Date(t.lastRun).toLocaleString() : "—"}</td>
                  <td><span className={`pill ${t.status === "ERR" ? "pill-bad" : "pill-ok"}`}>{t.status || "OK"}</span></td>
                  <td className="right">
                    <button onClick={()=>void runNow(t.id)} disabled={busyId===t.id} className="btn btn-sm btn-primary mr-2">Run Now</button>
                    <button onClick={()=>void removeTask(t.id)} disabled={busyId===t.id} className="btn btn-sm btn-ghost">Delete</button>
                  </td>
                </tr>
              ))}
              </tbody>
            </table>
          )}
        </div>
      </section>
    </main>
  );
}
