package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/bruli-lab/go-core/cqs"
	"github.com/bruli-lab/go-core/event"
	"github.com/bruli-lab/watersystem-ml/internal/app"
	"github.com/bruli-lab/watersystem-ml/internal/config"
	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	httpinfra "github.com/bruli-lab/watersystem-ml/internal/infra/http"
	"github.com/bruli-lab/watersystem-ml/internal/infra/influxdb2"
	"github.com/bruli-lab/watersystem-ml/internal/infra/listener"
	"github.com/bruli-lab/watersystem-ml/internal/infra/memory"
	"github.com/bruli-lab/watersystem-ml/internal/infra/ntfy"
	postgresinfra "github.com/bruli-lab/watersystem-ml/internal/infra/postgres"
	"github.com/bruli-lab/watersystem-ml/internal/infra/python"
	"github.com/bruli-lab/watersystem-ml/internal/infra/tracing"
	watersystem "github.com/bruli-lab/watersystem-ml/internal/infra/water_system"
	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"go.opentelemetry.io/otel"
)

const serviceName = "watersystem-ml"

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conf, err := config.New()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return err
	}
	log := buildLog(conf.LogLevel)

	zones, err := config.LoadZones(conf.ZonesFile)
	if err != nil {
		log.ErrorContext(ctx, "Error loading zones", "err", err)
		return err
	}
	humidityReferences, err := buildHumidityReferences(zones, time.Now())
	if err != nil {
		log.ErrorContext(ctx, "Error building humidity references", "err", err)
		return err
	}

	sqlDB, err := sql.Open("postgres", conf.PostgresDataSource())
	if err != nil {
		log.ErrorContext(ctx, "Error connecting to database", "err", err)
		return err
	}
	defer func() {
		_ = sqlDB.Close()
	}()
	if err := postgresinfra.RunMigrations(sqlDB); err != nil {
		log.ErrorContext(ctx, "Error running postgres migrations", "err", err)
		return err
	}

	db := bun.NewDB(sqlDB, pgdialect.New())
	defer func() {
		_ = db.Close()
	}()

	tracingProv, err := tracing.InitTracing(ctx, serviceName)
	if err != nil {
		log.ErrorContext(ctx, "Error initializing tracing", "err", err)
		return err
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()

		if err := tracingProv.Shutdown(shutdownCtx); err != nil {
			log.ErrorContext(ctx, "Error shutting down tracing", "err", err)
		}
	}()

	tracer := otel.Tracer(serviceName)

	trainExecutor := python.NewTrainingExecutor(30*time.Minute, conf.PythonPath, tracer, log)
	predictionRepo := python.NewPredictionRepository(tracer, conf.PythonPath, 3*time.Minute)
	dryingTrainExecutor := python.NewDryingTrainingExecutor(10*time.Minute, conf.PythonPath, tracer, log)
	shadowPredictor := python.NewShadowPredictor(tracer, conf.PythonPath, 3*time.Minute)
	ntfyPublisher, err := ntfy.NewPublisher(conf.NtfyUser, conf.NtfyPassword, conf.NtfyURL, conf.NtfyTopic, tracer)
	if err != nil {
		log.ErrorContext(ctx, "Error creating ntfy publisher", "err", err)
		return err
	}
	soilMeasureRepo := influxdb2.NewSoilMeasureRepository(
		conf.InfluxDBURL, conf.InfluxDBToken, conf.InfluxDBOrg, conf.InfluxDBBucket, zones.MoistureVoltageMeasurements(), tracer,
	)
	waterSystemExecutor, err := watersystem.NewExecutor(
		ctx, 5*time.Second, tracer, conf.WaterSystemHost, conf.WaterSystemPort, conf.WaterSystemToken, log,
	)
	if err != nil {
		log.ErrorContext(ctx, "Error creating water system executor", "err", err)
		return err
	}
	forecastRepo := influxdb2.NewForecastRepository(conf.InfluxDBURL, conf.InfluxDBToken, conf.InfluxDBOrg, conf.InfluxDBBucket, tracer)
	executionRepo := influxdb2.NewExecutionRepository(
		conf.InfluxDBURL, conf.InfluxDBToken, conf.InfluxDBOrg, conf.InfluxDBBucket, zones.Names(), tracer,
	)
	humidityRepo := memory.NewHumidityReferenceRepository(humidityReferences)
	waterSkippedLogRepo := influxdb2.NewWateringSkippedLogRepository(
		conf.InfluxDBURL, conf.InfluxDBToken, conf.InfluxDBOrg, conf.InfluxDBBucket, tracer,
	)
	predictionLogRepo := postgresinfra.NewPredictionLogRepository(db, tracer)
	modelHealthRepo := postgresinfra.NewModelHealthRepository(db, tracer)
	modelTrainingRepo := influxdb2.NewModelTrainingLogRepository(
		conf.InfluxDBURL, conf.InfluxDBToken, conf.InfluxDBOrg, conf.InfluxDBBucket, tracer,
	)
	modelTrainStateRepo := postgresinfra.NewModelTrainingStateRepository(db, tracer)
	shadowPredictionRepo := influxdb2.NewShadowPredictionRepository(
		conf.InfluxDBURL, conf.InfluxDBToken, conf.InfluxDBOrg, conf.InfluxDBBucket, tracer,
	)

	loc := buildTimeLocation(ctx, log)
	calculateWateringHnd := app.NewCalculateWatering(
		predictionRepo, soilMeasureRepo, humidityRepo, executionRepo, predictionLogRepo, waterSystemExecutor, forecastRepo,
		func() time.Time {
			return time.Now().In(loc)
		},
	)

	trainCh := make(chan struct{ Zone string }, 1)

	logChMiddleware := cqs.NewCommandHndErrorMiddleware(log, tracer)

	commandBus := cqs.NewCommandBus()
	commandBus.Subscribe(app.ExecuteWateringCommandName, logChMiddleware(app.NewExecuteWatering(waterSystemExecutor)))
	execWatOnWaterSysList := listener.NewExecuteWateringOnWateringRequested(commandBus)
	publishMsgList := listener.NewPublishMessageOnWateringRequested(commandBus)
	saveWaterSkippedLogWZList := listener.NewSaveWateringSkippedLogOnWateringZoneSkipped(commandBus)
	saveWaterSkippedLogWSList := listener.NewSaveWateringSkippedLogOnWateringSystemSkipped(commandBus)
	publishMsgPredErrList := listener.NewPublishMessageOnPredictionError(commandBus)
	publishMsgSensorList := listener.NewPublishMessageOnSoilMeasureUnavailable(commandBus)
	checkModelPredValFailedList := listener.NewCheckModelOnPredictionValidationFailed(commandBus)
	trainModelZoneModelDegradedList := listener.NewTrainModelOnZoneModelDegraded(trainCh)

	eventBus := event.NewBus()
	eventBus.Subscribe(
		ml.WateringRequestedEventName, publishMsgList, execWatOnWaterSysList, publishMsgPredErrList, publishMsgSensorList,
	)
	eventBus.Subscribe(ml.WateringZoneSkippedEventName, saveWaterSkippedLogWZList, publishMsgPredErrList, publishMsgSensorList)
	eventBus.Subscribe(ml.WateringSystemSkippedEventName, saveWaterSkippedLogWSList)
	eventBus.Subscribe(ml.PredictionValidationFailedEventName, checkModelPredValFailedList)
	eventBus.Subscribe(ml.ZoneModelDegradedEventName, trainModelZoneModelDegradedList)

	listenerMdw := cqs.NewCommandHandlerEventBusMiddleware(new(eventBus), tracer)
	multiCHMdw := cqs.CommandHandlerMultiMiddleware(logChMiddleware, listenerMdw)

	commandBus.Subscribe(app.CalculateWateringEventName, multiCHMdw(calculateWateringHnd))
	commandBus.Subscribe(app.PublishMessageCommandName, logChMiddleware(app.NewPublishMessage(ntfyPublisher)))
	commandBus.Subscribe(app.SaveWateringSkippedLogCommandName, logChMiddleware(app.NewSaveWateringSkippedLog(waterSkippedLogRepo)))
	validatePredictionHnd := app.NewValidatePrediction(soilMeasureRepo, predictionLogRepo, time.Now)
	commandBus.Subscribe(app.ValidatePredictionCommandName, multiCHMdw(validatePredictionHnd))
	commandBus.Subscribe(app.SavePredictionLogCommandName, logChMiddleware(app.NewSavePredictionLog(predictionLogRepo, time.Now)))
	checkFailedModelHnd := app.NewCheckFailedModel(modelHealthRepo, modelTrainStateRepo, time.Now)
	commandBus.Subscribe(app.CheckFailedModelCommandName, multiCHMdw(checkFailedModelHnd))
	trainingZoneHnd := app.NewTrainingZone(trainExecutor, modelTrainStateRepo, modelTrainingRepo, time.Now)
	commandBus.Subscribe(app.TrainingZoneCommandName, logChMiddleware(trainingZoneHnd))
	commandBus.Subscribe(app.TrainDryingModelCommandName, logChMiddleware(app.NewTrainDryingModel(dryingTrainExecutor)))
	commandBus.Subscribe(
		app.RunShadowPredictionCommandName,
		logChMiddleware(app.NewRunShadowPrediction(shadowPredictor, shadowPredictionRepo, time.Now)),
	)

	go runValidatePrediction(ctx, commandBus, log)

	go runTrainingZone(ctx, conf, zones.Names(), log, trainCh, commandBus)
	go runCalculateWatering(ctx, commandBus, log)
	go runShadowPrediction(ctx, commandBus, log)

	serverListener, err := new(net.ListenConfig).Listen(ctx, "tcp", conf.ServerHost)
	log.InfoContext(ctx, "Starting server", "host", conf.ServerHost)
	if err != nil {
		log.ErrorContext(ctx, "Error starting server", "err", err)
		return err
	}
	defer func() {
		_ = serverListener.Close()
	}()

	srv := httpinfra.NewServer(conf.ServerHost)
	errCh := make(chan error, 1)
	go func() {
		errCh <- runHTTPServer(ctx, srv, log, serverListener)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return <-errCh
	}
}

func buildHumidityReferences(zones config.Zones, now time.Time) (map[string]*ml.HumidityReference, error) {
	references := make(map[string]*ml.HumidityReference, len(zones))
	for _, zone := range zones {
		calibration, ok := zone.CalibrationAt(now)
		if !ok {
			return nil, fmt.Errorf("zone %q has no calibration valid at %s", zone.Name, now.Format(time.DateOnly))
		}
		references[zone.Name] = ml.NewHumidityReference(calibration.V40, calibration.V100)
	}
	return references, nil
}

func buildTimeLocation(ctx context.Context, log *slog.Logger) *time.Location {
	loc, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		log.ErrorContext(ctx, "Error loading location, using local time", "err", err)
		loc = time.Local
	}
	return loc
}

func runTrainingZone(
	ctx context.Context,
	conf *config.Config,
	zones []string,
	log *slog.Logger,
	ch <-chan struct{ Zone string },
	bus cqs.CommandBus,
) {
	initialTraining(ctx, conf, zones, log, bus)
	initialDryingTraining(ctx, zones, log, bus)
	log.InfoContext(ctx, "Starting training zone worker")
	for {
		select {
		case <-ctx.Done():
			return
		case zone := <-ch:
			func() {
				defer recoverBackground(ctx, log, "training zone")
				_, _ = bus.Handle(ctx, app.TrainingZoneCommand{Zone: zone.Zone})
			}()
		}
	}
}

func runValidatePrediction(ctx context.Context, bus cqs.CommandBus, log *slog.Logger) {
	tick := time.NewTicker(time.Minute)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			func() {
				defer recoverBackground(ctx, log, "validate prediction")
				_, _ = bus.Handle(ctx, app.ValidatePredictionCommand{
					Limit: time.Now(),
				})
			}()
		}
	}
}

func initialTraining(ctx context.Context, conf *config.Config, zones []string, log *slog.Logger, ch cqs.CommandHandler) {
	defer recoverBackground(ctx, log, "initial training")
	exists, empty, err := checkDir(conf.ModelDir)
	if err != nil {
		log.ErrorContext(ctx, "Error checking model dir", "err", err)
		return
	}

	if exists && !empty {
		log.InfoContext(ctx, "Model dir exists and is not empty, ignoring initial training")
		return
	}

	log.InfoContext(ctx, "Model dir is empty or does not exist, run initial training")
	executeTraining(ctx, ch, zones)
}

func initialDryingTraining(ctx context.Context, zones []string, log *slog.Logger, ch cqs.CommandHandler) {
	defer recoverBackground(ctx, log, "initial drying training")
	for _, zone := range zones {
		_, _ = ch.Handle(ctx, app.TrainDryingModelCommand{Zone: zone, IfMissing: true})
	}
}

func executeTraining(ctx context.Context, svc cqs.CommandHandler, zones []string) {
	for _, zone := range zones {
		_, _ = svc.Handle(ctx, app.TrainingZoneCommand{Zone: zone, Force: true})
	}
}

func checkDir(path string) (exists, empty bool, err error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, true, nil
		}
		return false, false, err
	}

	for _, entry := range entries {
		if entry.Name() == "lost+found" {
			continue
		}
		return true, false, nil
	}

	return true, true, nil
}

func runCalculateWatering(ctx context.Context, ch cqs.CommandHandler, log *slog.Logger) {
	tick := time.NewTicker(15 * time.Minute)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			runCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
			func() {
				defer cancel()
				defer recoverBackground(runCtx, log, "calculate watering")
				_, _ = ch.Handle(runCtx, app.CalculateWateringCommand{})
			}()
		}
	}
}

func runShadowPrediction(ctx context.Context, ch cqs.CommandHandler, log *slog.Logger) {
	tick := time.NewTicker(15 * time.Minute)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			runCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
			func() {
				defer cancel()
				defer recoverBackground(runCtx, log, "shadow prediction")
				_, _ = ch.Handle(runCtx, app.RunShadowPredictionCommand{})
			}()
		}
	}
}

func buildLog(level string) *slog.Logger {
	levels := map[string]slog.Level{
		"info":  slog.LevelInfo,
		"debug": slog.LevelDebug,
		"error": slog.LevelError,
		"warn":  slog.LevelWarn,
	}
	l := levels[level]

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: l,
	})

	log := slog.New(handler)
	return log.With("service", serviceName)
}

func recoverBackground(ctx context.Context, log *slog.Logger, worker string) {
	if r := recover(); r != nil {
		log.ErrorContext(ctx,
			"panic in background worker",
			slog.String("worker", worker),
			slog.Any("panic", r),
			slog.String("stack", string(debug.Stack())),
		)
	}
}

func runHTTPServer(ctx context.Context, srv *http.Server, log *slog.Logger, serverListener net.Listener) error {
	go shutdown(ctx, srv, log)

	if err := srv.Serve(serverListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.ErrorContext(ctx, "Error starting server", "err", err)
		return err
	}
	return nil
}

func shutdown(ctx context.Context, srv *http.Server, log *slog.Logger) {
	<-ctx.Done()
	log.InfoContext(ctx, "Shutdown signal received, stopping server")

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("error shutting down server", "err", err)
	}
}
