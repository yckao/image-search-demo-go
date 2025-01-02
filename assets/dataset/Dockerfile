FROM python:3.12-slim

WORKDIR /app

ARG DEBIAN_FRONTEND=noninteractive

RUN --mount=type=cache,target=/var/cache/apt \
    --mount=type=cache,target=/var/lib/apt/lists \
    apt-get update && apt-get install -y --no-install-recommends unzip

RUN --mount=type=bind,source=mlops_take_home_quiz_2024.zip,target=/app/mlops_take_home_quiz_2024.zip \
    unzip mlops_take_home_quiz_2024.zip -d /app

COPY prepare_dataset.py /app/prepare_dataset.py

CMD ["python", "/app/prepare_dataset.py"]
