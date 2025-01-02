-- Write your migrate up statements here
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE images (
    id UUID PRIMARY KEY,
    storage_provider VARCHAR(30) NOT NULL,
    storage_key VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT unique_image_storage_provider_key UNIQUE (storage_provider, storage_key)
);

CREATE TABLE image_embeddings (
    id UUID PRIMARY KEY,
    image_id UUID NOT NULL,
    model_name VARCHAR(30) NOT NULL,
    embedding vector(512) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (image_id) REFERENCES images(id) ON DELETE CASCADE
);

CREATE INDEX embedding_cosine_idx ON image_embeddings USING hnsw (embedding vector_cosine_ops);

CREATE TABLE search_queries (
    id UUID PRIMARY KEY,
    model_name VARCHAR(30) NOT NULL,
    query_text VARCHAR(255) NOT NULL,
    query_embedding vector(512) NOT NULL,
    result_image_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (result_image_id) REFERENCES images(id) ON DELETE CASCADE
);

CREATE TABLE search_feedbacks (
    id UUID PRIMARY KEY,
    search_query_id UUID NOT NULL,
    rating VARCHAR(20) CHECK (rating IN ('POSITIVE', 'NEGATIVE')) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (search_query_id) REFERENCES search_queries(id) ON DELETE CASCADE,
    UNIQUE (search_query_id)
);

---- create above / drop below ----

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
DROP TABLE IF EXISTS search_feedbacks;
DROP TABLE IF EXISTS search_queries;
DROP INDEX IF EXISTS embedding_cosine_idx;
DROP TABLE IF EXISTS image_embeddings;
DROP TABLE IF EXISTS images;