import React from "react";
import { Cpu, MemoryStick, HardDrive, Wifi } from "lucide-react";
import { StatCard } from "../components/ui/StatCard";
import { Card } from "../components/ui/Card";
import { Skeleton } from "../components/ui/Skeleton";
import { BadgeDot } from "../components/ui/BadgeDot";
import { useMockMetrics } from "../features/useMockMetrics";
import { Sparkline } from "../components/ui/Sparkline";
import { useSparklineData } from "../features/useSparklineData";

export function Dashboard(){
  const m = useMockMetrics();

  return (
    <main className="main-pad">
      {/* KPI row */}
      <section className="kpi-row">
        <StatCard icon={<Cpu width={20} height={20}/>} label="CPU" value={m.cpu}/>
        <StatCard icon={<MemoryStick width={20} height={20}/>} label="Memory" value={m.mem}/>
        <StatCard icon={<HardDrive width={20} height={20}/>} label="Disk I/O" value={m.io}/>
        <StatCard icon={<Wifi width={20} height={20}/>} label="Network" value={m.net}/>
      </section>

      {/* Charts + Settings */}
      <section className="grid-main mt-16">
        <Card>
          <div className="section-title">
            <h3 className="text-dim" style={{ fontSize: 14 }}>Latency</h3>
            <span className="hint">1h</span>
          </div>
          <Skeleton />
        </Card>

        <Card>
          <h3 className="text-dim mb-12" style={{ fontSize: 14 }}>Settings</h3>
          <div>
            <Row label="Sample interval">10 s</Row>
            <Row label="Retention days">30</Row>
            <Row label="Token"><input className="input" placeholder="••••••••" /></Row>
            <Row label="Data dir"><input className="input" placeholder="/var/lib/sysdash" /></Row>
            <button className="button" style={{ marginTop: 8 }}>Save</button>
          </div>
        </Card>

        <Card>
          <div className="section-title">
            <h3 className="text-dim" style={{ fontSize: 14 }}>Error rate</h3>
            <span className="hint">6h</span>
          </div>
          <Skeleton />
        </Card>

        <Card>
          <h3 className="text-dim mb-12" style={{ fontSize: 14 }}>Tasks</h3>
          <table className="table">
            <thead>
              <tr>
                <Th>ID</Th><Th>Type</Th><Th>Schedule</Th><Th>Last run</Th><Th>Status</Th><Th>Actions</Th>
              </tr>
            </thead>
            <tbody>
              {["t1","t2","t3","t4"].map((id, i) => (
                <tr key={id}>
                  <Td>{id}</Td>
                  <Td>{["Backup Logs","Clear Temp","Reboot System","Rotate Keys"][i]}</Td>
                  <Td>{["Daily","Weekly","Monthly","Daily"][i]}</Td>
                  <Td>12:00</Td>
                  <Td>
                    <span style={{ display:"inline-flex", alignItems:"center", gap:8 }}>
                      <BadgeDot/> <span>OK</span>
                    </span>
                  </Td>
                  <Td className="right"><button className="button secondary">Run Now</button></Td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      </section>
    </main>
  );
}

/* helpers */
function Row({ label, children }: { label:string; children: React.ReactNode }){
  return (
    <div style={{ display:"flex", alignItems:"center", justifyContent:"space-between", gap:16, marginBottom:12 }}>
      <div className="text-dim">{label}</div>
      <div style={{ minWidth:160, textAlign:"right" }}>{children}</div>
    </div>
  );
}
function Th({ children, className="" }: React.ThHTMLAttributes<HTMLTableHeaderCellElement>){
  return <th className={className}>{children}</th>;
}
function Td({ children, className="" }: React.TdHTMLAttributes<HTMLTableCellElement>){
  return <td className={className}>{children}</td>;
}
