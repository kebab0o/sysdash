import { Activity, Cpu, HardDrive, Settings } from "lucide-react";

export function Sidebar(){
  return (
    <aside className="sidebar">
      <div className="logo">SysDash</div>
      <nav>
        <a className="is-active" href="#"><Activity width={16} height={16}/> Dashboard</a>
        <a href="#"><Cpu width={16} height={16}/> CPU/Memory</a>
        <a href="#"><HardDrive width={16} height={16}/> Disk/IO</a>
        <a href="#"><Settings width={16} height={16}/> Tasks</a>
      </nav>
    </aside>
  );
}
