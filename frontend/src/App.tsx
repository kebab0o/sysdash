import "./global.css";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import Sidebar from "./components/Sidebar";
import Headbar from "./components/Headbar";
import Dashboard from "./pages/Dashboard";
import CpuMem from "./pages/CpuMem";
import DiskIO from "./pages/DiskIO";
import Tasks from "./pages/Tasks";
import Logs from "./pages/Logs";

export default function App() {
  return (
    <BrowserRouter>
      <div className="app">
        <aside className="sidebar">
          <div className="logo">SysDash</div>
          <Sidebar />
        </aside>
        <section className="content">
          <Headbar />
          <Routes>
            <Route path="/" element={<Navigate to="/dashboard" replace />} />
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/cpu-memory" element={<CpuMem />} />
            <Route path="/disk-io" element={<DiskIO />} />
            <Route path="/tasks" element={<Tasks />} />
            <Route path="/logs" element={<Logs />} />
          </Routes>
        </section>
      </div>
    </BrowserRouter>
  );
}
