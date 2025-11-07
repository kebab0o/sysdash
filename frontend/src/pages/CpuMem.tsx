import { useEffect, useMemo, useState } from "react";
import { api, type CPUPoint, type MemPoint } from "../api";
import { Sparkline } from "../components/ui/Sparkline";
import { Skeleton } from "../components/ui/Skeleton";

export default function CpuMem(){
  const [cpu,setCpu]=useState<CPUPoint[]>([]);
  const [mem,setMem]=useState<MemPoint[]>([]);
  const [loading,setLoading]=useState(true);
  const [range,setRange]=useState("1h");

  useEffect(()=>{ let on=true; setLoading(true);
    Promise.all([api.cpu(range), api.mem(range)])
      .then(([c,m])=>{
        if(!on) return;
        setCpu((c?.points??[]));
        setMem((m?.points??[]));
      })
      .finally(()=>{ if(on) setLoading(false); });
    return ()=>{ on=false; };
  },[range]);

  const cpuSeries = useMemo(()=> cpu.map(p=>p.v),[cpu]);
  const memSeries = useMemo(()=> mem.map(p=>p.v),[mem]);

  return (
    <main className="page">
      <div className="statusbar">
        <div className="status-left"><span className="status-strong">CPU & Memory</span></div>
        <div className="status-right">
          <select className="input" value={range} onChange={e=>setRange(e.target.value)} style={{width:120}}>
            <option value="5m">5m</option><option value="1h">1h</option><option value="24h">24h</option>
          </select>
        </div>
      </div>

      <section className="panels">
        <div className="panel">
          <h2 className="panel-title">CPU % ({range})</h2>
          {loading ? <Skeleton className="h-40 w-full" /> : <Sparkline data={cpuSeries} height={180} />}
        </div>
        <div className="panel">
          <h2 className="panel-title">Memory used % ({range})</h2>
          {loading ? <Skeleton className="h-40 w-full" /> : <Sparkline data={memSeries} height={180} />}
        </div>
      </section>
    </main>
  );
}
