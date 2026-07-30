#!/bin/sh
set -e

KEYS_DIR="/app/keys"

# Create keys directory if missing
if [ ! -d "$KEYS_DIR" ]; then
    mkdir -p "$KEYS_DIR"
fi

# Generate JWT RSA keys if missing
if [ ! -f "$KEYS_DIR/private.pem" ] || [ ! -f "$KEYS_DIR/public.pem" ]; then
    echo ">> Generating JWT RSA key pair..."
    openssl genpkey -algorithm RSA -out "$KEYS_DIR/private.pem" -pkeyopt rsa_keygen_bits:2048
    openssl rsa -pubout -in "$KEYS_DIR/private.pem" -out "$KEYS_DIR/public.pem"
    chmod 600 "$KEYS_DIR/private.pem"
    chmod 644 "$KEYS_DIR/public.pem"
    echo ">> JWT keys generated successfully."
fi

echo ">> Starting SabtBrooker backend..."
exec /app/server
