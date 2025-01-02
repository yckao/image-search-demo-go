#!/bin/sh

echo "Waiting for MinIO to be ready..."
until curl -s "http://localhost:$MINIO_SERVER_PORT/minio/health/ready" >/dev/null 2>&1; do
  sleep 1
done

echo "Configuring mc client..."
mc config host add local http://localhost:$MINIO_SERVER_PORT "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"

echo "Creating bucket: $S3_BUCKET_NAME"
mc mb local/"$S3_BUCKET_NAME" || echo "Bucket may already exist."
mc admin user add local "$S3_ACCESS_KEY" "$S3_SECRET_KEY"
mc admin policy attach local readwrite --user "$S3_ACCESS_KEY"

echo "Bucket-creation script complete."
