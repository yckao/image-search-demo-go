# Image Search Demo In **Go + Rust**

A text-to-image search service powered by OpenAI's CLIP model. This service allows users to search for images using natural language queries and provide feedback on search results.


## Highlight
- About 2x faster than the original Python implementation
- Way smaller image size
- Blazing fast startup time

## Features

- 🔍 Text-to-image search using CLIP embeddings
- 📊 User feedback collection on search results
- 🗄️ Vector similarity search with pgvector
- 🐳 Fully containerized with Docker
- 📦 S3-compatible object storage support

## Architecture

The service consists of several components:

- Go +Go-Kit Architecture
- Rust + candle for Model Inference
- PostgreSQL + pgvector for Vector Similarity Search
- MinIO for S3-compatible Object Storage

## Quick Start

1. Clone the repository:

> [!IMPORTANT]
> Pease ensure git-lfs is installed otherwise the assets will not be downloaded.
> For mac users, you can install it by running `brew install git-lfs && git lfs install`

```bash
git clone https://github.com/yckao/image-search-demo-go.git
cd image-search-demo-go
```

2. Create environent file for development

```bash
cp .env.example .env
```

3. Start the services using Docker Compose:

```bash
docker compose -f hack/compose/docker-compose.yaml up -d
```

4. [Opitional] Upload assets

```bash
docker compose -f hack/compose/docker-compose.yaml up -d dataset
```

5. Open the API documentation at http://localhost:8080/swagger/index.html

### Clean up

```bash
docker compose -f hack/compose/docker-compose.yaml down -v
```

## Future Work
- [ ] Multi-Thread GPU Support (Currently only CPU can run concurrently)
- [ ] Add validation middlewares
- [ ] Add tests
- [ ] Add observability (metrics, tracing, logging)
