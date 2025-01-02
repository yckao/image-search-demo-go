package imagerepository

import (
	"context"
	"errors"

	"github.com/go-kit/log"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"github.com/yckao/image-search-demo-go/pkg/errortypes"
	"github.com/yckao/image-search-demo-go/pkg/models"
	"github.com/yckao/image-search-demo-go/services/image/imagemodel"
)

type PGRepository struct {
	logger log.Logger
	db     *pgxpool.Pool
}

func NewPGRepository(logger log.Logger, db *pgxpool.Pool) Repository {
	return &PGRepository{logger: logger, db: db}
}

func (r *PGRepository) CreateImage(ctx context.Context, image *models.Image, embedding *imagemodel.ImageEmbedding) (*models.Image, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if image.ID == uuid.Nil {
		image.ID = uuid.Must(uuid.NewV7())
	}

	if _, err = tx.Exec(ctx,
		"INSERT INTO images (id, storage_provider, storage_key) VALUES ($1, $2, $3)",
		image.ID, image.StorageProvider, image.StorageKey); err != nil {
		return nil, err
	}

	if embedding.ID == uuid.Nil {
		embedding.ID = uuid.Must(uuid.NewV7())
	}

	if _, err = tx.Exec(ctx,
		"INSERT INTO image_embeddings (id, image_id, model_name, embedding) VALUES ($1, $2, $3, $4)",
		embedding.ID, image.ID, embedding.ModelName, pgvector.NewVector(embedding.Embedding)); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	if err = r.db.QueryRow(ctx,
		"SELECT created_at FROM images WHERE id = $1",
		image.ID).Scan(&image.CreatedAt); err != nil {
		return nil, err
	}

	return image, nil
}

func (r *PGRepository) GetImage(ctx context.Context, id uuid.UUID) (*models.Image, error) {
	image := models.Image{}

	if err := r.db.QueryRow(ctx, "SELECT id, storage_provider, storage_key, created_at FROM images WHERE id = $1", id).
		Scan(&image.ID, &image.StorageProvider, &image.StorageKey, &image.CreatedAt); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, errortypes.NewErrImageNotFound(id)
	} else if err != nil {
		return nil, err
	}

	return &image, nil
}

func (r *PGRepository) CreateSearchQuery(ctx context.Context, searchQuery *imagemodel.SearchQuery) (*models.SearchWithImage, error) {
	if searchQuery.ID == uuid.Nil {
		searchQuery.ID = uuid.Must(uuid.NewV7())
	}

	var imageID uuid.UUID
	if err := r.db.QueryRow(ctx,
		"SELECT image_id FROM image_embeddings WHERE model_name = $1 ORDER BY embedding <=> $2 ASC, created_at DESC LIMIT 1",
		searchQuery.ModelName, pgvector.NewVector(searchQuery.Embedding)).Scan(&imageID); err != nil && errors.Is(err, pgx.ErrNoRows) {
		return nil, errortypes.NewErrNoImageAvailable(searchQuery.ModelName)
	} else if err != nil {
		return nil, err
	}

	if _, err := r.db.Exec(ctx,
		"INSERT INTO search_queries (id, model_name, query_text, query_embedding, result_image_id) VALUES ($1, $2, $3, $4, $5)",
		searchQuery.ID, searchQuery.ModelName, searchQuery.QueryText, pgvector.NewVector(searchQuery.Embedding), imageID); err != nil {
		return nil, err
	}

	image := models.Image{}

	if err := r.db.QueryRow(ctx,
		"SELECT q.created_at, i.id, storage_provider, storage_key, i.created_at FROM search_queries q JOIN images i ON q.result_image_id = i.id WHERE q.id = $1",
		searchQuery.ID).Scan(&searchQuery.CreatedAt, &image.ID, &image.StorageProvider, &image.StorageKey, &image.CreatedAt); err != nil {
		return nil, err
	}

	searchWithImage := &models.SearchWithImage{
		Search: models.Search{
			ID:        searchQuery.ID,
			ModelName: searchQuery.ModelName,
			QueryText: searchQuery.QueryText,
			CreatedAt: searchQuery.CreatedAt,
		},
		Image: models.Image{
			ID:              image.ID,
			StorageProvider: image.StorageProvider,
			StorageKey:      image.StorageKey,
			CreatedAt:       image.CreatedAt,
		},
	}

	return searchWithImage, nil
}

func (r *PGRepository) GetSearchQuery(ctx context.Context, id uuid.UUID) (*models.SearchWithImage, error) {
	searchQuery := models.SearchWithImage{}

	if err := r.db.QueryRow(ctx,
		"SELECT s.id, s.model_name, s.query_text, s.created_at, i.id, i.storage_provider, i.storage_key, i.created_at FROM search_queries s LEFT JOIN images i ON s.result_image_id = i.id WHERE s.id = $1", id).
		Scan(&searchQuery.ID, &searchQuery.ModelName, &searchQuery.QueryText, &searchQuery.CreatedAt, &searchQuery.Image.ID, &searchQuery.Image.StorageProvider, &searchQuery.Image.StorageKey, &searchQuery.Image.CreatedAt); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, errortypes.NewErrSearchQueryNotFound(id)
	} else if err != nil {
		return nil, err
	}

	return &searchQuery, nil
}

func (r *PGRepository) CreateSearchFeedback(ctx context.Context, feedback *models.SearchFeedbackWithQuery) (*models.SearchFeedbackWithQuery, error) {
	if feedback.ID == uuid.Nil {
		feedback.ID = uuid.Must(uuid.NewV7())
	}

	if _, err := r.db.Exec(ctx, "INSERT INTO search_feedbacks (id, search_query_id, rating) VALUES ($1, $2, $3)", feedback.ID, feedback.Query.ID, string(feedback.Rating)); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.ForeignKeyViolation {
				return nil, errortypes.NewErrSearchQueryNotFound(feedback.Query.ID)
			}

			if pgErr.Code == pgerrcode.UniqueViolation {
				return nil, errortypes.NewErrSearchFeedbackAlreadyExists(feedback.Query.ID)
			}
		}
		return nil, err
	}

	if err := r.db.QueryRow(ctx,
		"SELECT f.created_at, sq.id, sq.model_name, sq.query_text, sq.created_at FROM search_feedbacks f JOIN search_queries sq ON f.search_query_id = sq.id WHERE f.id = $1", feedback.ID).
		Scan(&feedback.CreatedAt, &feedback.Query.ID, &feedback.Query.ModelName, &feedback.Query.QueryText, &feedback.Query.CreatedAt); err != nil {
		return nil, err
	}

	return feedback, nil
}
