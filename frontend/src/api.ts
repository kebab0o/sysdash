export type CPUPoint = { t: string; v: number };
export type MemPoint = { t: string; v: number };
export type DiskPoint = { t: string; mount: string; usedPct: number; usedGB: number; totalGB: number };
export type DiskIOPoint = { t: string; readMBs: number; writeMBs: number };
export type NetPoint = { t: string; rxKBs: number; txKBs: number };

const BASE = import.meta.env.VITE_API_URL || "http://localhost:8080";
const KEY  = import.meta.env.VITE_API_KEY || "";

async function req<T>(path: string, init: RequestInit = {}): Promise<T | null> {
  const headers = new Headers(init.headers ?? {});
  if (KEY) headers.set("X-API-Key", KEY);
  if (!headers.has("Content-Type") && init.body) headers.set("Content-Type", "application/json");
  const res = await fetch(`${BASE}${path}`, { ...init, headers });
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}: ${await res.text()}`);
  if (res.status === 204) return null;
  return (await res.json()) as unknown as T;
}

export const api = {
  health: () => req<{ status: string; now: string; lastCollectorAt: string }>("/api/health"),
  cpu:   (range = "1h")  => req<{ range: string; points: CPUPoint[]; avg: number; p95: number }>(`/api/metrics/cpu?range=${range}`),
  mem:   (range = "1h")  => req<{ range: string; points: MemPoint[]; latest: number }>(`/api/metrics/mem?range=${range}`),
  disk:  (range = "24h") => req<{ range: string; mounts: { mount: string; points: DiskPoint[] }[] }>(`/api/metrics/disk?range=${range}`),
  diskio:(range = "1h")  => req<{ range: string; points: DiskIOPoint[] }>(`/api/metrics/diskio?range=${range}`),
  net:   (range = "1h")  => req<{ range: string; points: NetPoint[] }>(`/api/metrics/net?range=${range}`),
  prune: () => req<{ status: string }>(`/api/tasks/prune`, { method: "POST" }),
};
