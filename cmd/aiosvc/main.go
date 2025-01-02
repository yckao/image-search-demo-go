package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/go-kit/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/oklog/pkg/group"
	pgxvector "github.com/pgvector/pgvector-go/pgx"
	"github.com/spf13/viper"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/yckao/image-search-demo-go/api/openapi"
	"github.com/yckao/image-search-demo-go/pkg/clients/clip"
	"github.com/yckao/image-search-demo-go/services/image/imageendpoint"
	"github.com/yckao/image-search-demo-go/services/image/imagerepository"
	"github.com/yckao/image-search-demo-go/services/image/imageservice"
	"github.com/yckao/image-search-demo-go/services/image/imagetransport"
	"github.com/yckao/image-search-demo-go/services/storage/storageendpoint"
	"github.com/yckao/image-search-demo-go/services/storage/storageservice"
	"github.com/yckao/image-search-demo-go/services/storage/storagetransport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func initConfig() {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	viper.ReadInConfig()

	viper.SetDefault("BIND_ADDR", "0.0.0.0:8080")

	viper.MustBindEnv("BASE_URL")

	viper.MustBindEnv("CLIP_GRPC_ADDR")
	viper.MustBindEnv("PGHOST")
	viper.MustBindEnv("PGPORT")
	viper.MustBindEnv("PGUSER")
	viper.MustBindEnv("PGPASSWORD")
	viper.MustBindEnv("PGDATABASE")

	viper.MustBindEnv("S3_ENDPOINT_URL")
	viper.MustBindEnv("S3_ACCESS_KEY")
	viper.MustBindEnv("S3_SECRET_KEY")
	viper.MustBindEnv("S3_BUCKET_NAME")
	viper.MustBindEnv("S3_URL_FORMAT")
}

func main() {
	initConfig()

	ctx := context.Background()
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	conn, err := grpc.NewClient(viper.GetString("CLIP_GRPC_ADDR"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log("clip", "error", err)
		os.Exit(1)
	}

	pgxconfig, err := pgxpool.ParseConfig(fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", viper.GetString("PGUSER"), viper.GetString("PGPASSWORD"), viper.GetString("PGHOST"), viper.GetInt("PGPORT"), viper.GetString("PGDATABASE")))
	if err != nil {
		logger.Log("db", "error", err)
		os.Exit(1)
	}
	pgxconfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgxvector.RegisterTypes(ctx, conn)
	}
	db, err := pgxpool.NewWithConfig(ctx, pgxconfig)

	if err != nil {
		logger.Log("db", "error", err)
		os.Exit(1)
	}
	clipService, err := clip.NewGRPCClient(conn)
	if err != nil {
		logger.Log("clip", "error", err)
		os.Exit(1)
	}

	var (
		storageService = storageservice.NewS3Service(logger, storageservice.S3ServiceConfig{
			Endpoint:  viper.GetString("S3_ENDPOINT_URL"),
			Bucket:    viper.GetString("S3_BUCKET_NAME"),
			AccessKey: viper.GetString("S3_ACCESS_KEY"),
			SecretKey: viper.GetString("S3_SECRET_KEY"),
			BaseURL:   viper.GetString("BASE_URL"),
			URLFormat: viper.GetString("S3_URL_FORMAT"),
		})
		storageEnpoints    = storageendpoint.New(storageService, logger)
		storageHTTPHandler = storagetransport.NewHTTPHandler(storageEnpoints, logger)
	)

	imageRepository := imagerepository.NewPGRepository(logger, db)

	var (
		imageService     = imageservice.New(logger, clipService, storageService, imageRepository)
		imageEndpoint    = imageendpoint.New(imageService, logger)
		imageHTTPHandler = imagetransport.NewHTTPHandler(imageEndpoint, logger)
	)

	httpHandler := http.NewServeMux()

	httpHandler.Handle("/images", imageHTTPHandler)
	httpHandler.Handle("/images/", imageHTTPHandler)
	httpHandler.Handle("/storage/", storageHTTPHandler)
	httpHandler.Handle("/openapi.json", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(openapi.OpenAPIJSON)
	}))
	httpHandler.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("%s/openapi.json", viper.GetString("BASE_URL"))),
	))

	var g group.Group
	{
		httpListener, err := net.Listen("tcp", viper.GetString("BIND_ADDR"))
		if err != nil {
			logger.Log("transport", "HTTP", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			logger.Log("transport", "HTTP", "addr", httpListener.Addr())
			return http.Serve(httpListener, httpHandler)
		}, func(error) {
			_ = httpListener.Close()
		})
	}
	logger.Log("exit", g.Run())
}
