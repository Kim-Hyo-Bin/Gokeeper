#!/usr/bin/env python3
"""Offline Ed25519 license validation (same wire format as GoKeeper).

Env: PUBLIC_KEY_PATH (default ./license_public.pem), LICENSE_KEY (required).

Depends: cryptography
"""
from __future__ import annotations

import base64
import json
import os
import sys
import time

from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric.ed25519 import Ed25519PublicKey


def b64url_decode(s: str) -> bytes:
    pad = "=" * (-len(s) % 4)
    return base64.urlsafe_b64decode(s + pad)


def main() -> None:
    pem_path = os.environ.get("PUBLIC_KEY_PATH", "license_public.pem")
    key_str = os.environ.get("LICENSE_KEY") or (sys.argv[1] if len(sys.argv) > 1 else "")
    if not key_str:
        sys.exit("set LICENSE_KEY or pass license_key as argv[1]")

    with open(pem_path, "rb") as f:
        pub = serialization.load_pem_public_key(f.read())
    if not isinstance(pub, Ed25519PublicKey):
        sys.exit("public key must be Ed25519")

    parts = key_str.split(".")
    if len(parts) != 2:
        sys.exit("invalid license key format")
    payload = b64url_decode(parts[0])
    sig = b64url_decode(parts[1])
    pub.verify(sig, payload)

    claims = json.loads(payload.decode("utf-8"))
    lic_id = claims.get("license_id")
    exp = int(claims.get("exp", 0))
    if not lic_id:
        sys.exit("invalid claims")
    if exp != 0 and time.time() > exp:
        sys.exit(f"expired (exp={exp})")
    print(f"ok license_id={lic_id} exp={exp}")


if __name__ == "__main__":
    main()
