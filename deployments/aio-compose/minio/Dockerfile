# Dockerfile.minio
FROM minio/minio:latest

ENV MINIO_SERVER_PORT=9000
ENV MINIO_CONSOLE_PORT=9001

# Copy your bucket-creation script into the image
COPY create-bucket.sh /usr/bin/create-bucket.sh
RUN chmod +x /usr/bin/create-bucket.sh

# Copy an entrypoint script
COPY docker-entrypoint.sh /usr/bin/docker-entrypoint.sh
RUN chmod +x /usr/bin/docker-entrypoint.sh

# Override the ENTRYPOINT
ENTRYPOINT ["/usr/bin/docker-entrypoint.sh"]
