import { NavLink } from "react-router-dom";

export default function Sidebar() {
  return (
    <nav className="space-y-1">
      <NavLink to="/dashboard" className={({isActive}) => isActive ? "is-active" : ""}>Dashboard</NavLink>
      <NavLink to="/cpu-memory" className={({isActive}) => isActive ? "is-active" : ""}>CPU/Memory</NavLink>
      <NavLink to="/disk-io" className={({isActive}) => isActive ? "is-active" : ""}>Disk/IO</NavLink>
      <NavLink to="/tasks" className={({isActive}) => isActive ? "is-active" : ""}>Tasks</NavLink>
      <NavLink to="/logs" className={({isActive}) => isActive ? "is-active" : ""}>Logs</NavLink>
    </nav>
  );
}
