import { useEffect, useRef, useState } from "react";

export function useSparklineData(opts?: { length?: number; start?: number; amp?: number; clampMax?: number }) {
  const length = opts?.length ?? 60;
  const amp = opts?.amp ?? 8;
  const clampMax = opts?.clampMax ?? 1000;

  const [series, setSeries] = useState<number[]>(
    Array.from({ length }, (_, i) => (opts?.start ?? 50) + Math.sin(i / 4) * amp + jitter(amp / 3))
  );
  const timer = useRef<number | null>(null);

  useEffect(() => {
    timer.current = window.setInterval(() => {
      setSeries(prev => {
        const next = prev.slice(1);
        const last = next[next.length - 1] ?? prev[prev.length - 1] ?? (opts?.start ?? 50);
        const value = clamp(last + jitter(amp), 0, clampMax);
        next.push(value);
        return next;
      });
    }, 1200);
    return () => { if (timer.current) clearInterval(timer.current); };
  }, [amp, clampMax, opts?.start]);

  return series;
}

function jitter(a: number) {
  return (Math.random() - 0.5) * a * 2;
}
function clamp(v: number, lo: number, hi: number) {
  return Math.min(hi, Math.max(lo, v));
}
