import { useEffect, useState } from "react";
import { api, type Task } from "../api";

export function TasksPanel() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [name, setName] = useState("Clear Temp");
  const [every, setEvery] = useState(60);
  const [busyId, setBusyId] = useState<string | null>(null);

  async function load() {
    setLoading(true);
    try {
      const t = await api.tasks.list();
      setTasks(t ?? []);
    } finally {
      setLoading(false);
    }
  }
  useEffect(() => { void load(); }, []);

  async function add(e: React.FormEvent) {
    e.preventDefault();
    if (every <= 0) return;
    await api.tasks.create(name, every);
    setName("Clear Temp");
    setEvery(60);
    void load();
  }

  async function runNow(id: string) {
    setBusyId(id);
    try {
      await api.tasks.runNow(id);
      await load();
    } finally {
      setBusyId(null);
    }
  }

  async function remove(id: string) {
    setBusyId(id);
    try {
      await api.tasks.del(id);
      await load();
    } finally {
      setBusyId(null);
    }
  }

  return (
    <div className="rounded-xl border border-neutral-800 p-4">
      <h2 className="text-sm text-neutral-400 mb-3">Tasks</h2>

      <form className="flex gap-2 mb-3" onSubmit={add}>
        <input
          className="flex-1 rounded-md bg-neutral-900 border border-neutral-700 px-2 py-1 text-sm"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Task name"
        />
        <input
          className="w-32 rounded-md bg-neutral-900 border border-neutral-700 px-2 py-1 text-sm"
          type="number"
          min={1}
          value={every}
          onChange={(e) => setEvery(Number(e.target.value))}
          placeholder="Every (min)"
        />
        <button className="rounded-md bg-blue-600 hover:bg-blue-500 px-3 py-1 text-sm">
          Add
        </button>
      </form>

      {loading ? (
        <div className="text-sm text-neutral-500">Loading…</div>
      ) : tasks.length === 0 ? (
        <div className="text-sm text-neutral-500">No tasks yet.</div>
      ) : (
        <table className="w-full text-sm">
          <thead className="text-neutral-500">
            <tr>
              <th className="text-left py-1">Name</th>
              <th className="text-left py-1">Every</th>
              <th className="text-left py-1">Last run</th>
              <th className="text-left py-1">Status</th>
              <th className="text-right py-1">Actions</th>
            </tr>
          </thead>
          <tbody className="text-neutral-300">
            {tasks.map((t) => (
              <tr key={t.id} className="border-t border-neutral-800">
                <td className="py-1">{t.name}</td>
                <td className="py-1">{t.everyMinutes} min</td>
                <td className="py-1">{t.lastRun ? new Date(t.lastRun).toLocaleTimeString() : "—"}</td>
                <td className="py-1">{t.status || "OK"}</td>
                <td className="py-1 text-right space-x-2">
                  <button
                    onClick={() => void runNow(t.id)}
                    disabled={busyId === t.id}
                    className="px-2 py-1 rounded bg-blue-600 hover:bg-blue-500 disabled:opacity-50"
                  >
                    Run Now
                  </button>
                  <button
                    onClick={() => void remove(t.id)}
                    disabled={busyId === t.id}
                    className="px-2 py-1 rounded bg-neutral-800 hover:bg-neutral-700 disabled:opacity-50"
                  >
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
