package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "net/http/pprof"

	v1 "github.com/The-Gleb/go_metrics_and_alerting/internal/controller/http/v1/handler"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/controller/http/v1/middleware"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/service"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/usecase"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository/database"
	filestorage "github.com/The-Gleb/go_metrics_and_alerting/internal/repository/file_storage"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository/memory"
	"github.com/go-chi/chi/v5"
)

// postgres://metric_db:metric_db@localhost:5434/metric_db?sslmode=disable

var (
	BuildVersion string = "N/A"
	BuildDate    string = "N/A"
	BuildCommit  string = "N/A"
)

func main() {
	fmt.Printf(
		"Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		BuildVersion, BuildDate, BuildCommit,
	)

	if err := Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func Run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT)
	defer cancel()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	config := NewConfigFromFlags()

	if err := logger.Initialize(config.LogLevel); err != nil {
		logger.Log.Fatal(err)
		return err
	}
	logger.Log.Info(config)

	var repository service.MetricStorage
	var fileStorage service.FileStorage

	if config.FileStoragePath != "" {
		fileStorage = filestorage.MustGetFileStorage(config.FileStoragePath)
	}

	if config.DatabaseDSN != "" {
		db, err := database.NewMetricDB(ctx, config.DatabaseDSN)
		if err != nil {
			return err
		}
		repository = db
	} else {
		repository = memory.New()
	}

	metricServie := service.NewMetricService(repository)
	backupService := service.NewBackupService(repository, fileStorage, config.StoreInterval, config.Restore)

	updateMetricUsecase := usecase.NewUpdateMetricUsecase(metricServie, backupService)
	updateMetricSetUsecase := usecase.NewUpdateMetricSetUsecase(metricServie, backupService)
	getMetricUsecase := usecase.NewGetMetricUsecase(metricServie)
	getAllMetricsUsecase := usecase.NewGetAllMetricsUsecase(metricServie)

	updateMetricHandler := v1.NewUpdateMetricHandler(updateMetricUsecase)
	updateMetricJSONHandler := v1.NewUpdateMetricJSONHandler(updateMetricUsecase)
	getMetricHandler := v1.NewGetMetricHandler(getMetricUsecase)
	getMetricJSONHandler := v1.NewGetMetricJSONHandler(getMetricUsecase)
	updateMetricSetHandler := v1.NewUpdateMetricSetHandler(updateMetricSetUsecase)
	getAllMetricsHandler := v1.NewGetAllMetricsHandler(getAllMetricsUsecase)

	gzipMiddleware := middleware.NewGzipMiddleware()
	checkSignatureMiddleware := middleware.NewCheckSignatureMiddleware([]byte(config.SignKey))
	loggerMidleware := middleware.NewLoggerMiddleware(logger.Log)

	r := chi.NewMux()
	r.Use(gzipMiddleware.Do, checkSignatureMiddleware.Do, loggerMidleware.Do)

	updateMetricHandler.AddToRouter(r)
	updateMetricJSONHandler.AddToRouter(r)
	getMetricHandler.AddToRouter(r)
	getMetricJSONHandler.AddToRouter(r)
	updateMetricSetHandler.AddToRouter(r)
	getAllMetricsHandler.AddToRouter(r)

	s := http.Server{
		Addr:    config.Addres,
		Handler: r,
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := backupService.Run(ctx)
		if err != nil {
			cancel()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		<-ctx.Done()

		ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := s.Shutdown(ctxShutdown)
		if err != nil {
			panic(err)
		}
		logger.Log.Info("server shutdown")
	}()

	logger.Log.Info("starting server")
	if err := s.ListenAndServe(); err != nil {
		logger.Log.Error("server error", "error", err)
	}

	return nil
}
