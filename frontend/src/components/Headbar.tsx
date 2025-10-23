// src/components/HeaderBar.tsx
export function Headbar() {
  return (
  <header className="header">
      <div className="text-dim">
        <span className="badge-dot" /> <span>Sampling active</span>
        <span className="mono-dim" style={{ marginLeft: 8 }}>Last updated 12:03:22</span>
      </div>
      <div className="badge-dot" />
    </header>
  );
}
