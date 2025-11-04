import { useEffect, useMemo, useState } from "react";
import { api } from "../api";
import type { CPUPoint, DiskPoint } from "../api";

import { Sparkline } from "../components/ui/Sparkline";
import { StatCard } from "../components/ui/StatCard";
import { Skeleton } from "../components/ui/Skeleton";

type CPUResp = { range: string; points: CPUPoint[]; avg: number; p95: number };
type DiskResp = { range: string; mounts: { mount: string; points: DiskPoint[] }[] };

// tiny inline icons
const CpuIcon = (
  <svg width="20" height="20" viewBox="0 0 24 24" role="img" aria-label="cpu">
    <rect x="6" y="6" width="12" height="12" rx="2" fill="currentColor" opacity="0.15" />
    <rect x="8" y="8" width="8" height="8" rx="1.5" stroke="currentColor" fill="none" />
  </svg>
);
const P95Icon = (
  <svg width="20" height="20" viewBox="0 0 24 24" role="img" aria-label="percentile">
    <path d="M5 16l4-4 4 4 6-6" stroke="currentColor" strokeWidth="2" fill="none" />
  </svg>
);
const DiskIcon = (
  <svg width="20" height="20" viewBox="0 0 24 24" role="img" aria-label="disk">
    <circle cx="12" cy="12" r="9" stroke="currentColor" fill="none" />
    <circle cx="12" cy="12" r="2" fill="currentColor" />
  </svg>
);

function useCpu(range: string = "1h") {
  const [data, setData] = useState<CPUResp | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    let on = true;
    setLoading(true);
    api
      .cpu(range)
      .then((d) => {
        if (!on) return;
        setData(d ?? null);
        setErr(null);
      })
      .catch((e: unknown) => {
        if (!on) return;
        const msg = e instanceof Error ? e.message : String(e);
        setErr(msg);
      })
      .finally(() => {
        if (on) setLoading(false);
      });
    return () => {
      on = false;
    };
  }, [range]);

  const series = useMemo<number[]>(
    () => (data?.points ?? []).map((p) => p.v),
    [data]
  );

  return { data, series, loading, err };
}

function useDisk(range: string = "24h") {
  const [data, setData] = useState<DiskResp | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    let on = true;
    setLoading(true);
    api
      .disk(range)
      .then((d) => {
        if (!on) return;
        setData(d ?? null);
        setErr(null);
      })
      .catch((e: unknown) => {
        if (!on) return;
        const msg = e instanceof Error ? e.message : String(e);
        setErr(msg);
      })
      .finally(() => {
        if (on) setLoading(false);
      });
    return () => {
      on = false;
    };
  }, [range]);

  const primary = data?.mounts?.[0];
  const series = useMemo<number[]>(
    () => (primary?.points ?? []).map((p) => p.usedPct),
    [primary]
  );
  const latestPct = useMemo<number>(
    () =>
      primary?.points && primary.points.length > 0
        ? primary.points[primary.points.length - 1].usedPct
        : 0,
    [primary]
  );

  return { data, series, latestPct, loading, err };
}

export default function Dashboard() {
  const {
    data: cpu,
    series: cpuSeries,
    loading: cpuLoading,
    err: cpuErr,
  } = useCpu("1h");

  const {
    series: diskSeries,
    latestPct,
    loading: diskLoading,
    err: diskErr,
  } = useDisk("24h");

  const healthText =
    latestPct >= 90 || (cpu?.p95 ?? 0) >= 90
      ? "Hot"
      : latestPct >= 80 || (cpu?.p95 ?? 0) >= 75
      ? "Watch"
      : "Healthy";

  return (
    <main className="p-6 space-y-6">
      <header className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Sysdash</h1>
        <span className="inline-flex items-center rounded-full px-3 py-1 text-sm bg-neutral-200 dark:bg-neutral-800">
          {healthText}
        </span>
      </header>

      {/* KPI cards (StatCard expects icon, label, value, optional sub) */}
      <section className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <StatCard
          icon={CpuIcon}
          label="CPU avg (1h)"
          value={cpuLoading ? "…" : `${(cpu?.avg ?? 0).toFixed(1)}%`}
          sub={cpuErr ?? "Average over last hour"}
        />
        <StatCard
          icon={P95Icon}
          label="CPU p95 (1h)"
          value={cpuLoading ? "…" : `${(cpu?.p95 ?? 0).toFixed(1)}%`}
          sub="95th percentile"
        />
        <StatCard
          icon={DiskIcon}
          label="Disk usage"
          value={diskLoading ? "…" : `${latestPct.toFixed(1)}%`}
          sub={diskErr ?? "Primary mount"}
        />
      </section>

      {/* Trend charts below the cards */}
      <section className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="rounded-xl border border-neutral-200 dark:border-neutral-800 p-4">
          <h2 className="mb-2 text-sm text-neutral-500">CPU trend (1h)</h2>
          {cpuLoading ? (
            <Skeleton className="h-40 w-full" />
          ) : (
            <Sparkline data={cpuSeries} height={160} />
          )}
        </div>
        <div className="rounded-xl border border-neutral-200 dark:border-neutral-800 p-4">
          <h2 className="mb-2 text-sm text-neutral-500">Disk usage trend (24h)</h2>
          {diskLoading ? (
            <Skeleton className="h-40 w-full" />
          ) : (
            <Sparkline data={diskSeries} height={160} />
          )}
        </div>
      </section>
    </main>
  );
}
