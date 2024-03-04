package main

import (
	// "bufio"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	// "time"

	_ "net/http/pprof"

	// "github.com/The-Gleb/go_metrics_and_alerting/internal/app"
	// "github.com/The-Gleb/go_metrics_and_alerting/internal/compressor"
	v1 "github.com/The-Gleb/go_metrics_and_alerting/internal/controller/http/v1/handler"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/controller/http/v1/middleware"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/service"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/usecase"
	// "github.com/The-Gleb/go_metrics_and_alerting/internal/filestorage"
	// "github.com/The-Gleb/go_metrics_and_alerting/internal/handlers"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	// "github.com/The-Gleb/go_metrics_and_alerting/internal/repositories"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository/database"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository/memory"
	// "github.com/The-Gleb/go_metrics_and_alerting/internal/repository"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository/file_storage"
	// "github.com/The-Gleb/go_metrics_and_alerting/internal/server"
	// "github.com/The-Gleb/go_metrics_and_alerting/pkg/utils/retry"
	"github.com/go-chi/chi/v5"
)

// postgres://metric_db:metric_db@localhost:5434/metric_db?sslmode=disable

// TODO: retries
func main() {

	if err := Run(); err != nil {
		log.Fatal(err)
	}

	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	// config := NewConfigFromFlags()

	// if err := logger.Initialize(config.LogLevel); err != nil {
	// 	logger.Log.Fatal(err)
	// 	return
	// }
	// logger.Log.Info(config)

	// var repository repositories.Repositiries
	// var fileStorage app.FileStorage

	// if config.FileStoragePath != "" {
	// 	repository = memory.New()
	// 	fileStorage = filestorage.NewFileStorage(config.FileStoragePath, config.StoreInterval, config.Restore)
	// }

	// if config.DatabaseDSN != "" {
	// 	var db *database.DB
	// 	var err error
	// 	err = retry.DefaultRetry(
	// 		context.Background(),
	// 		func(ctx context.Context) error {
	// 			db, err = database.ConnectDB(config.DatabaseDSN)
	// 			return err
	// 		},
	// 	)

	// 	if err != nil {
	// 		logger.Log.Fatal(err)
	// 		return
	// 	}
	// 	repository = db
	// }

	// app := app.NewApp(repository, fileStorage)
	// handlers := handlers.New(app)
	// s := server.NewWithProfiler(config.Addres, handlers, []byte(config.SignKey))

	// if config.Restore {
	// 	app.LoadDataFromFile(context.Background())
	// }

	// var wg sync.WaitGroup
	// ctx, cancel := context.WithCancel(context.Background())

	// if config.StoreInterval > 0 && config.DatabaseDSN == "" {
	// saveTicker := time.NewTicker(time.Duration(config.StoreInterval) * time.Second)
	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()
	// for {
	// 	select {
	// 	case <-saveTicker.C:
	// 		app.StoreDataToFile(context.Background())
	// 	case <-ctx.Done():
	// 		logger.Log.Debug("stop saving to file")
	// 		return
	// 	}
	// }

	// 	}()
	// }

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	ServerShutdownSignal := make(chan os.Signal, 1)
	// 	signal.Notify(ServerShutdownSignal, syscall.SIGINT)
	// 	<-ServerShutdownSignal
	// 	s.Shutdown(context.Background())
	// 	cancel()
	// }()

	// err := server.Run(s)
	// if err != nil && err != http.ErrServerClosed {
	// 	panic(err)
	// }
	// os.
	// 	wg.Wait()
	// logger.Log.Info("server shutdown")
}

func Run() error {

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
		fileStorage = filestorage.NewFileStorage(config.FileStoragePath)
	}

	if config.DatabaseDSN != "" {

		db, err := database.ConnectDB(config.DatabaseDSN)
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

	ctx, cancelCtx := context.WithCancel(context.Background())

	wg.Add(1)
	go func() {
		defer wg.Done()
		backupService.Run(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ServerShutdownSignal := make(chan os.Signal, 1)
		signal.Notify(ServerShutdownSignal, syscall.SIGINT)

		<-ServerShutdownSignal

		cancelCtx()
		err := s.Shutdown(context.Background())
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
