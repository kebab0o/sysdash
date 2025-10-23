import { Card } from "./Card";
import { type ReactNode } from "react";

export function StatCard({ icon, label, value, sub }: {
  icon: ReactNode; label: string; value: string; sub?: string;
}){
  return (
    <Card className="kpi">
      <div className="icon-box">{icon}</div>
      <div>
        <div className="mono-dim" style={{ fontSize: 12 }}>{label}</div>
        <div style={{ fontSize: 20, fontWeight: 600 }}>{value}</div>
        {sub && <div className="mono-dim" style={{ fontSize: 12 }}>{sub}</div>}
      </div>
    </Card>
  );
}
