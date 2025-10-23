import "./global.css";
import { Sidebar } from "./components/Sidebar";
import { Headbar } from "./components/Headbar";
import { Dashboard } from "./pages/Dashboard";

export default function App(){
  return (
    <div className="app">
      <Sidebar />
      <div className="content">
        <Headbar />
        <Dashboard />
      </div>
    </div>
  );
}
