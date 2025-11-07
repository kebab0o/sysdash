import { useCallback, useEffect, useState } from "react";
import { api } from "../api";

type Task = { id: string; name: string; everyMinutes: number; lastRun?: string; status?: string };

export default function Tasks(){
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

  return (
    <main className="page">
      <div className="statusbar"><div className="status-left"><span className="status-strong">Tasks</span></div></div>

      <section className="tasks-grid">
        <div className="card">
          <div className="card-title">Add task</div>
          <form className="task-form" onSubmit={addTask}>
            <input className="input" value={name} onChange={(e)=>setName(e.target.value)} placeholder="Task name" />
            <input className="input input-num" type="number" min={1} value={every} onChange={(e)=>setEvery(Number(e.target.value))} placeholder="Every (min)" />
            <button className="btn btn-primary">Add</button>
          </form>
          <p className="muted small mt-2">Tasks run your OS-temp cleanup and retention pruning.</p>
        </div>

        <div className="card overflow">
          <div className="card-title">All tasks</div>
          {tLoading ? <div className="muted">Loading…</div> :
           tasks.length===0 ? <div className="muted">No tasks yet.</div> :
           (<table className="table">
              <thead><tr><th>Name</th><th>Every</th><th>Last run</th><th>Status</th><th className="right">Actions</th></tr></thead>
              <tbody>
                {tasks.map(t=>(
                  <tr key={t.id}>
                    <td>{t.name}</td>
                    <td>{t.everyMinutes} min</td>
                    <td>{t.lastRun ? new Date(t.lastRun).toLocaleString() : "—"}</td>
                    <td><span className={`pill ${t.status==="ERR"?"pill-bad":"pill-ok"}`}>{t.status||"OK"}</span></td>
                    <td className="right">
                      <button onClick={()=>void runNow(t.id)} disabled={busyId===t.id} className="btn btn-sm btn-primary mr-2">Run Now</button>
                      <button onClick={()=>void removeTask(t.id)} disabled={busyId===t.id} className="btn btn-sm btn-ghost">Delete</button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>)}
        </div>
      </section>
    </main>
  );
}
