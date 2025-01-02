#!/bin/sh

set -e

# 1. Start MinIO in the background
echo "Starting MinIO in the background..."
/usr/bin/minio server /data --console-address ":${MINIO_CONSOLE_PORT}" &
MINIO_PID=$!

# 2. Wait for MinIO to come online, then create the bucket
/usr/bin/create-bucket.sh

# 3. Wait for MinIO to stop (or hand off control to MinIO)
wait $MINIO_PID
