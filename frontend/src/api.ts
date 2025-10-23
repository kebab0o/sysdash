const BASE = import.meta.env.VITE_API_URL || "http://localhost:8080";
const KEY  = import.meta.env.VITE_API_KEY || "";

export async function listItems() {
  const res = await fetch(`${BASE}/api/items`, {
    headers: KEY ? { "X-API-Key": KEY } : {}
  });
  if (!res.ok) throw new Error("list failed");
  return res.json();
}