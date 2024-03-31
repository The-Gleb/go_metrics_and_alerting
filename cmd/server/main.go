package main

//go:generate go run ../../internal/encryption

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

	"github.com/The-Gleb/go_metrics_and_alerting/internal/controller"
	grpcserver "github.com/The-Gleb/go_metrics_and_alerting/internal/controller/grpc/server"

	// httpServer "github.com/The-Gleb/go_metrics_and_alerting/internal/controller/http/v1/server"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/service"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/usecase"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository/database"
	filestorage "github.com/The-Gleb/go_metrics_and_alerting/internal/repository/file_storage"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository/memory"
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
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	config, err := BuildConfig()
	if err != nil {
		return err
	}

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

	var s controller.Server
	s, err = grpcserver.NewServer(
		config.Address,
		[]byte(config.SignKey),
		logger.Log,
		config.PrivateKeyPath,
		config.TrustedSubnet,
		updateMetricUsecase,
		updateMetricSetUsecase,
		getMetricUsecase,
		getAllMetricsUsecase,
	)
	if err != nil {
		return err
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
		err := s.Stop(ctxShutdown)
		if err != nil {
			panic(err)
		}
		logger.Log.Info("server shutdown")
	}()

	logger.Log.Info("starting server")
	if err := s.Start(); err != nil {
		logger.Log.Error("server error", "error", err)
	}

	return nil
}
