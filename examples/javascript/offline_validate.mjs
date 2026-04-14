/**
 * Offline Ed25519 license validation (Node built-in crypto).
 *
 * Env: PUBLIC_KEY_PATH (default ./license_public.pem)
 * Arg: license_key string (or LICENSE_KEY env)
 * Node 18+
 */

import { createPublicKey, verify } from "node:crypto";
import { readFileSync } from "node:fs";

const pemPath = process.env.PUBLIC_KEY_PATH ?? "license_public.pem";
const licenseKey =
  process.env.LICENSE_KEY ?? process.argv[2] ?? "";
if (!licenseKey) {
  console.error("set LICENSE_KEY or pass license_key as argv[2]");
  process.exit(1);
}

const parts = licenseKey.split(".");
if (parts.length !== 2) {
  console.error("invalid license key format");
  process.exit(1);
}

const payload = Buffer.from(parts[0], "base64url");
const sig = Buffer.from(parts[1], "base64url");
const pem = readFileSync(pemPath, "utf8");
const key = createPublicKey(pem);

const ok = verify(null, payload, key, sig);
if (!ok) {
  console.error("signature verify failed");
  process.exit(1);
}

const claims = JSON.parse(payload.toString("utf8"));
if (!claims.license_id) {
  console.error("invalid claims");
  process.exit(1);
}
const exp = Number(claims.exp ?? 0);
if (exp !== 0 && Math.floor(Date.now() / 1000) > exp) {
  console.error(`expired (exp=${exp})`);
  process.exit(1);
}
console.log(`ok license_id=${claims.license_id} exp=${exp}`);
