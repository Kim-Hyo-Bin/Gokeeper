/**
 * HTTP API: issue (7-day), get, verify, revoke.
 *
 * Env: BASE_URL (default http://localhost:8080)
 * Node 18+
 */

const BASE = (process.env.BASE_URL ?? "http://localhost:8080").replace(/\/$/, "");

const exp = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString();
const issueRes = await fetch(`${BASE}/v1/licenses`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ expires_at: exp }),
});
if (!issueRes.ok) throw new Error(`issue ${issueRes.status} ${await issueRes.text()}`);
const { id, license_key: licenseKey } = await issueRes.json();
console.log("issued id:", id);

const getRes = await fetch(`${BASE}/v1/licenses/${id}`);
console.log("get:", getRes.status);

const verifyRes = await fetch(`${BASE}/v1/licenses/verify`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ license_key: licenseKey }),
});
console.log("verify:", await verifyRes.json());

const revokeRes = await fetch(`${BASE}/v1/licenses/${id}/revoke`, { method: "POST" });
console.log("revoke:", revokeRes.status, await revokeRes.text());
