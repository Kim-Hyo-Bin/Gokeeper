#!/usr/bin/env python3
"""HTTP API: issue (7-day), get, verify, revoke.

Env: BASE_URL (default http://localhost:8080)
"""
from __future__ import annotations

import os
from datetime import datetime, timedelta, timezone

import requests

BASE = os.environ.get("BASE_URL", "http://localhost:8080").rstrip("/")


def main() -> None:
    exp = (datetime.now(timezone.utc) + timedelta(days=7)).isoformat()
    r = requests.post(
        f"{BASE}/v1/licenses",
        json={"expires_at": exp},
        timeout=15,
    )
    r.raise_for_status()
    data = r.json()
    lic_id = data["id"]
    lic_key = data["license_key"]
    print("issued id:", lic_id)

    r2 = requests.get(f"{BASE}/v1/licenses/{lic_id}", timeout=15)
    print("get:", r2.status_code)

    r3 = requests.post(
        f"{BASE}/v1/licenses/verify",
        json={"license_key": lic_key},
        timeout=15,
    )
    print("verify:", r3.json())

    r4 = requests.post(f"{BASE}/v1/licenses/{lic_id}/revoke", timeout=15)
    print("revoke:", r4.status_code, r4.text)


if __name__ == "__main__":
    main()
