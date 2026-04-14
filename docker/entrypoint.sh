#!/bin/sh
# First container start: generate Ed25519 PKCS#8 key pair if no private key is present.
# Override with LICENSE_PRIVATE_KEY_PEM, or mount LICENSE_PRIVATE_KEY_PATH, or set
# LICENSE_AUTO_GENERATE_KEY=false to require a pre-existing key file.
set -eu

if [ -n "${LICENSE_PRIVATE_KEY_PEM:-}" ]; then
	unset LICENSE_PRIVATE_KEY_PATH || true
	exec /app/server "$@"
fi

if [ -z "${LICENSE_PRIVATE_KEY_PATH:-}" ]; then
	KEY_PATH=/data/license_private.pem
else
	KEY_PATH="$LICENSE_PRIVATE_KEY_PATH"
fi
export LICENSE_PRIVATE_KEY_PATH="$KEY_PATH"

PUB_PATH="$(dirname "$KEY_PATH")/license_public.pem"

if [ ! -f "$KEY_PATH" ]; then
	case "${LICENSE_AUTO_GENERATE_KEY:-true}" in
	false | 0 | no | NO)
		echo "gokeeper: private key not found at $KEY_PATH (mount a key or set LICENSE_AUTO_GENERATE_KEY=true)" >&2
		exit 1
		;;
	esac
	d=$(dirname "$KEY_PATH")
	mkdir -p "$d"
	openssl genpkey -algorithm ED25519 -out "$KEY_PATH"
	chmod 600 "$KEY_PATH" 2>/dev/null || true
	openssl pkey -in "$KEY_PATH" -pubout -out "$PUB_PATH"
	chmod 644 "$PUB_PATH" 2>/dev/null || true
	echo "gokeeper: generated Ed25519 key pair (persist /data or your key volume; rotate for production)" >&2
	echo "gokeeper:   private $KEY_PATH" >&2
	echo "gokeeper:   public  $PUB_PATH" >&2
fi

exec /app/server "$@"
