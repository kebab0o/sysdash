import { useEffect, useState } from "react";

export function useMockMetrics() {
  const [cpu, setCpu] = useState(42);
  const [mem, setMem] = useState(73);
  const [io, setIo] = useState(32.5);
  const [net, setNet] = useState(540);

  useEffect(() => {
    const id = setInterval(() => {
      setCpu(v => clamp(wobble(v, 0.6), 5, 98));
      setMem(v => clamp(wobble(v, 0.5), 10, 97));
      setIo(v => clamp(wobble(v, 1.2), 1, 120));
      setNet(v => clamp(wobble(v, 8), 10, 2000));
    }, 1200);
    return () => clearInterval(id);
  }, []);

  return {
    cpu: `${cpu.toFixed(0)}%`,
    mem: `${mem.toFixed(0)}%`,
    io: `${io.toFixed(1)} MB/s`,
    net: `${net.toFixed(0)} KB/s`,
  };
}

function wobble(v: number, amp: number) { return v + (Math.random() - 0.5) * amp * 2; }
function clamp(v: number, lo: number, hi: number) { return Math.min(hi, Math.max(lo, v)); }
