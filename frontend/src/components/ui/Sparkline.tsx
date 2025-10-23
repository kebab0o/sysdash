import React from "react";

type SparklineProps = {
  data: number[];
  width?: number;
  height?: number;
  stroke?: string;
  fill?: string;
  grid?: boolean;
  yUnit?: string;
  className?: string;
};

export function Sparkline({
  data,
  width = 800,
  height = 180,
  stroke = "rgba(180, 206, 255, 1)",
  fill = "rgba(180, 206, 255, .12)",
  grid = true,
  yUnit,
  className = "",
}: SparklineProps) {
  if (!data || data.length < 2) {
    return <div className={`skeleton ${className}`} style={{ height }} />;
  }

  const w = 600;     // internal viewBox width
  const h = 120;     // internal viewBox height
  const pad = 8;

  const min = Math.min(...data);
  const max = Math.max(...data);
  const span = max - min || 1;

  const stepX = (w - pad * 2) / (data.length - 1);

  const pts = data.map((v, i) => {
    const x = pad + i * stepX;
    const y = pad + (h - pad * 2) * (1 - (v - min) / span);
    return [x, y];
  });

  const lineD = pts.map(([x, y], i) => `${i === 0 ? "M" : "L"}${x},${y}`).join(" ");
  const areaD = `${lineD} L${pad + (data.length - 1) * stepX},${h - pad} L${pad},${h - pad} Z`;

  const gridLines: React.ReactNode[] = [];
  if (grid) {
    const rows = 4;
    for (let r = 0; r <= rows; r++) {
      const y = pad + (h - pad * 2) * (r / rows);
      gridLines.push(
        <line
          key={`g${r}`}
          x1={pad}
          y1={y}
          x2={w - pad}
          y2={y}
          stroke="rgba(255,255,255,.06)"
          strokeWidth={1}
        />
      );
    }
  }

  return (
    <div className={`chart ${className}`} style={{ width, height }}>
      <svg viewBox={`0 0 ${w} ${h}`} width="100%" height="100%" role="img" aria-label="sparkline">
        {gridLines}
        <path d={areaD} fill={fill} />
        <path d={lineD} fill="none" stroke={stroke} strokeWidth={2} strokeLinejoin="round" strokeLinecap="round" />
      </svg>
      <div className="chart-legend">
        <span className="mono-dim">min {min.toFixed(0)}{yUnit ?? ""}</span>
        <span className="mono-dim">max {max.toFixed(0)}{yUnit ?? ""}</span>
      </div>
    </div>
  );
}
