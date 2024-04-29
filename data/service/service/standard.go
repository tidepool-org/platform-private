package service

import (
	"context"
	stderrors "errors"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/kelseyhightower/envconfig"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/go-common/asyncevents"
	eventsCommon "github.com/tidepool-org/go-common/events"
	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/application"
	dataDeduplicatorDeduplicator "github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	dataDeduplicatorFactory "github.com/tidepool-org/platform/data/deduplicator/factory"
	dataEvents "github.com/tidepool-org/platform/data/events"
	"github.com/tidepool-org/platform/data/service/api"
	dataServiceApiV1 "github.com/tidepool-org/platform/data/service/api/v1"
	dataSourceServiceClient "github.com/tidepool-org/platform/data/source/service/client"
	dataSourceStoreStructured "github.com/tidepool-org/platform/data/source/store/structured"
	dataSourceStoreStructuredMongo "github.com/tidepool-org/platform/data/source/store/structured/mongo"
	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/data/types/blood"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/evalalerts"
	"github.com/tidepool-org/platform/events"
	logInternal "github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/devlog"
	metricClient "github.com/tidepool-org/platform/metric/client"
	"github.com/tidepool-org/platform/permission"
	permissionClient "github.com/tidepool-org/platform/permission/client"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/service/server"
	"github.com/tidepool-org/platform/service/service"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	syncTaskMongo "github.com/tidepool-org/platform/synctask/store/mongo"
)

type Standard struct {
	*service.DEPRECATEDService
	metricClient              *metricClient.Client
	permissionClient          *permissionClient.Client
	dataDeduplicatorFactory   *dataDeduplicatorFactory.Factory
	dataStore                 *dataStoreMongo.Store
	dataSourceStructuredStore *dataSourceStoreStructuredMongo.Store
	syncTaskStore             *syncTaskMongo.Store
	dataClient                *Client
	alertsClient              *alerts.Client
	dataSourceClient          *dataSourceServiceClient.Client
	userEventsHandler         events.Runner
	alertsEventsHandler       events.Runner
	api                       *api.Standard
	server                    *server.Standard
}

func NewStandard() *Standard {
	return &Standard{
		DEPRECATEDService: service.NewDEPRECATEDService(),
	}
}

func (s *Standard) Initialize(provider application.Provider) error {
	if err := s.DEPRECATEDService.Initialize(provider); err != nil {
		return err
	}

	if err := s.initializeMetricClient(); err != nil {
		return err
	}
	if err := s.initializePermissionClient(); err != nil {
		return err
	}
	if err := s.initializeDataDeduplicatorFactory(); err != nil {
		return err
	}
	if err := s.initializeDataStore(); err != nil {
		return err
	}
	if err := s.initializeDataSourceStructuredStore(); err != nil {
		return err
	}
	if err := s.initializeSyncTaskStore(); err != nil {
		return err
	}
	if err := s.initializeDataClient(); err != nil {
		return err
	}
	if err := s.initializeDataSourceClient(); err != nil {
		return err
	}
	if err := s.initializeUserEventsHandler(); err != nil {
		return err
	}
	if err := s.initializeAlertsEventsHandler(); err != nil {
		return err
	}
	if err := s.initializeAPI(); err != nil {
		return err
	}
	return s.initializeServer()
}

func (s *Standard) Terminate() {
	if s.server != nil {
		if err := s.server.Shutdown(); err != nil {
			s.Logger().Errorf("Error while terminating the the server: %v", err)
		}
		s.server = nil
	}
	if s.userEventsHandler != nil {
		s.Logger().Info("Terminating the userEventsHandler")
		if err := s.userEventsHandler.Terminate(); err != nil {
			s.Logger().Errorf("Error while terminating the userEventsHandler: %v", err)
		}
		s.userEventsHandler = nil
	}
	if s.alertsEventsHandler != nil {
		if err := s.alertsEventsHandler.Terminate(); err != nil {
			s.Logger().Errorf("Error while terminating the alertsEventsHandler: %v", err)
		}
		s.alertsEventsHandler = nil
	}
	s.api = nil
	s.dataClient = nil
	s.alertsClient = nil
	if s.syncTaskStore != nil {
		s.syncTaskStore.Terminate(context.Background())
		s.syncTaskStore = nil
	}
	if s.dataSourceStructuredStore != nil {
		s.dataSourceStructuredStore.Terminate(context.Background())
		s.dataSourceStructuredStore = nil
	}
	if s.dataStore != nil {
		s.dataStore.Terminate(context.Background())
		s.dataStore = nil
	}
	s.dataDeduplicatorFactory = nil
	s.permissionClient = nil
	s.metricClient = nil

	s.DEPRECATEDService.Terminate()
}

func (s *Standard) Run() error {
	if s.server == nil {
		return errors.New("service not initialized")
	}

	errs := make(chan error)
	go func() {
		errs <- s.userEventsHandler.Run()
	}()
	go func() {
		errs <- s.alertsEventsHandler.Run()
	}()
	go func() {
		errs <- s.server.Serve()
	}()

	return <-errs
}

func (s *Standard) PermissionClient() permission.Client {
	return s.permissionClient
}

func (s *Standard) DataSourceStructuredStore() dataSourceStoreStructured.Store {
	return s.dataSourceStructuredStore
}

func (s *Standard) initializeMetricClient() error {
	s.Logger().Debug("Loading metric client config")

	cfg := platform.NewConfig()
	cfg.UserAgent = s.UserAgent()
	reporter := s.ConfigReporter().WithScopes("metric", "client")
	loader := platform.NewConfigReporterLoader(reporter)
	if err := cfg.Load(loader); err != nil {
		return errors.Wrap(err, "unable to load metric client config")
	}

	s.Logger().Debug("Creating metric client")

	clnt, err := metricClient.New(cfg, platform.AuthorizeAsUser, s.Name(), s.VersionReporter())
	if err != nil {
		return errors.Wrap(err, "unable to create metric client")
	}
	s.metricClient = clnt

	return nil
}

func (s *Standard) initializePermissionClient() error {
	s.Logger().Debug("Loading permission client config")

	cfg := platform.NewConfig()
	cfg.UserAgent = s.UserAgent()
	reporter := s.ConfigReporter().WithScopes("permission", "client")
	loader := platform.NewConfigReporterLoader(reporter)
	if err := cfg.Load(loader); err != nil {
		return errors.Wrap(err, "unable to load permission client config")
	}

	s.Logger().Debug("Creating permission client")

	clnt, err := permissionClient.New(cfg, platform.AuthorizeAsService)
	if err != nil {
		return errors.Wrap(err, "unable to create permission client")
	}
	s.permissionClient = clnt

	return nil
}

func (s *Standard) initializeDataDeduplicatorFactory() error {
	s.Logger().Debug("Creating device deactivate hash deduplicator")

	deviceDeactivateHashDeduplicator, err := dataDeduplicatorDeduplicator.NewDeviceDeactivateHash()
	if err != nil {
		return errors.Wrap(err, "unable to create device deactivate hash deduplicator")
	}

	s.Logger().Debug("Creating device truncate data set deduplicator")

	deviceTruncateDataSetDeduplicator, err := dataDeduplicatorDeduplicator.NewDeviceTruncateDataSet()
	if err != nil {
		return errors.Wrap(err, "unable to create device truncate data set deduplicator")
	}

	s.Logger().Debug("Creating data set delete origin deduplicator")

	dataSetDeleteOriginDeduplicator, err := dataDeduplicatorDeduplicator.NewDataSetDeleteOrigin()
	if err != nil {
		return errors.Wrap(err, "unable to create data set delete origin deduplicator")
	}

	s.Logger().Debug("Creating none deduplicator")

	noneDeduplicator, err := dataDeduplicatorDeduplicator.NewNone()
	if err != nil {
		return errors.Wrap(err, "unable to create none deduplicator")
	}

	s.Logger().Debug("Creating data deduplicator factory")

	deduplicators := []dataDeduplicatorFactory.Deduplicator{
		deviceDeactivateHashDeduplicator,
		deviceTruncateDataSetDeduplicator,
		dataSetDeleteOriginDeduplicator,
		noneDeduplicator,
	}

	factory, err := dataDeduplicatorFactory.New(deduplicators)
	if err != nil {
		return errors.Wrap(err, "unable to create data deduplicator factory")
	}
	s.dataDeduplicatorFactory = factory

	return nil
}

func (s *Standard) initializeDataStore() error {
	s.Logger().Debug("Loading data store DEPRECATED config")

	cfg := storeStructuredMongo.NewConfig()
	if err := cfg.Load(); err != nil {
		return errors.Wrap(err, "unable to load data store DEPRECATED config")
	}
	if err := cfg.SetDatabaseFromReporter(s.ConfigReporter().WithScopes("DEPRECATED", "data", "store")); err != nil {
		return errors.Wrap(err, "unable to load data source structured store config")
	}

	s.Logger().Debug("Creating data store")

	str, err := dataStoreMongo.NewStore(cfg)
	if err != nil {
		return errors.Wrap(err, "unable to create data store DEPRECATED")
	}
	s.dataStore = str

	s.Logger().Debug("Ensuring data store DEPRECATED indexes")

	err = s.dataStore.EnsureIndexes()
	if err != nil {
		return errors.Wrap(err, "unable to ensure data store DEPRECATED indexes")
	}

	return nil
}

func (s *Standard) initializeDataSourceStructuredStore() error {
	s.Logger().Debug("Loading data source structured store config")

	cfg := storeStructuredMongo.NewConfig()
	if err := cfg.Load(); err != nil {
		return errors.Wrap(err, "unable to load data source structured store config")
	}

	s.Logger().Debug("Creating data source structured store")

	str, err := dataSourceStoreStructuredMongo.NewStore(cfg)
	if err != nil {
		return errors.Wrap(err, "unable to create data source structured store")
	}
	s.dataSourceStructuredStore = str

	s.Logger().Debug("Ensuring data source structured store indexes")

	err = s.dataSourceStructuredStore.EnsureIndexes()
	if err != nil {
		return errors.Wrap(err, "unable to ensure data source structured store indexes")
	}

	return nil
}

func (s *Standard) initializeSyncTaskStore() error {
	s.Logger().Debug("Loading sync task store config")

	cfg := storeStructuredMongo.NewConfig()
	if err := cfg.Load(); err != nil {
		return errors.Wrap(err, "unable to load sync task store config")
	}
	if err := cfg.SetDatabaseFromReporter(s.ConfigReporter().WithScopes("sync_task", "store")); err != nil {
		return errors.Wrap(err, "unable to load sync task store config")
	}

	s.Logger().Debug("Creating sync task store")

	str, err := syncTaskMongo.NewStore(cfg)
	if err != nil {
		return errors.Wrap(err, "unable to create sync task store")
	}
	s.syncTaskStore = str

	return nil
}

func (s *Standard) initializeDataClient() error {
	s.Logger().Debug("Creating data client")

	clnt, err := NewClient(s.dataStore)
	if err != nil {
		return errors.Wrap(err, "unable to create data client")
	}
	s.dataClient = clnt

	return nil
}

func (s *Standard) initializeDataSourceClient() error {
	s.Logger().Debug("Creating data client")

	clnt, err := dataSourceServiceClient.New(s)
	if err != nil {
		return errors.Wrap(err, "unable to create source data client")
	}
	s.dataSourceClient = clnt

	return nil
}

func (s *Standard) initializeAPI() error {
	s.Logger().Debug("Creating api")

	newAPI, err := api.NewStandard(s, s.metricClient, s.permissionClient,
		s.dataDeduplicatorFactory,
		s.dataStore, s.syncTaskStore, s.dataClient, s.dataSourceClient)
	if err != nil {
		return errors.Wrap(err, "unable to create api")
	}
	s.api = newAPI

	s.Logger().Debug("Initializing api middleware")

	if err = s.api.InitializeMiddleware(); err != nil {
		return errors.Wrap(err, "unable to initialize api middleware")
	}

	s.Logger().Debug("Initializing api router")

	if err = s.api.DEPRECATEDInitializeRouter(dataServiceApiV1.Routes()); err != nil {
		return errors.Wrap(err, "unable to initialize api router")
	}

	return nil
}

func (s *Standard) initializeServer() error {
	s.Logger().Debug("Loading server config")

	serverConfig := server.NewConfig()
	if err := serverConfig.Load(s.ConfigReporter().WithScopes("server")); err != nil {
		return errors.Wrap(err, "unable to load server config")
	}

	s.Logger().Debug("Creating server")

	newServer, err := server.NewStandard(serverConfig, s.Logger(), s.api)
	if err != nil {
		return errors.Wrap(err, "unable to create server")
	}
	s.server = newServer

	return nil
}

func (s *Standard) initializeUserEventsHandler() error {
	s.Logger().Debug("Initializing user events handler")
	sarama.Logger = log.New(os.Stdout, "SARAMA ", log.LstdFlags|log.Lshortfile)
	ctx := logInternal.NewContextWithLogger(context.Background(), s.Logger())
	handler := dataEvents.NewUserDataDeletionHandler(ctx, s.dataStore, s.dataSourceStructuredStore)
	handlers := []eventsCommon.EventHandler{handler}
	runner := events.NewRunner(handlers)
	if err := runner.Initialize(); err != nil {
		return errors.Wrap(err, "unable to initialize user events handler runner")
	}
	s.userEventsHandler = runner

	return nil
}

func (s *Standard) initializeAlertsEventsHandler() error {
	s.Logger().Debug("Initializing alerts events handler")

	// TODO: dev only, replace with jsonlogger for prod
	devLogger, err := devlog.NewWithDefaults(os.Stderr)
	if err != nil {
		return err
	}

	cfg := platform.NewConfig()
	cfg.UserAgent = s.UserAgent()
	reporter := s.ConfigReporter().WithScopes("alerts", "client")
	loader := platform.NewConfigReporterLoader(reporter)
	if err := cfg.Load(loader); err != nil {
		return errors.Wrap(err, "unable to load alerts client config")
	}
	pc, err := platform.NewClient(cfg, platform.AuthorizeAsService)
	if err != nil {
		return errors.Wrap(err, "unable to create platform client for alerts events handler")
	} else if pc == nil {
		return errors.New("unable to create platform client (got nil reponse)")
	}

	commonCfg := eventsCommon.NewConfig()
	if err := commonCfg.LoadFromEnv(); err != nil {
		return err
	}

	alertsCfg := &AlertsEventsHandlerConfig{}
	if err := alertsCfg.LoadFromEnv(); err != nil {
		return err
	}

	alertsClient := alerts.NewClient(pc, s.AuthClient(), devLogger)

	devLogger.WithField("cfg", alertsCfg).Info("here's the alerts events handler's config")
	s.alertsEventsHandler = NewSaramaAlertsAdapter(alertsCfg, s.dataStore, alertsClient, commonCfg.SaramaConfig, devLogger)
	return nil
}

// AlertsEventsHandlerConfig provides Kafka-specific configuration for the
// alerts events handlers. Some of the names don't line up perfectly, because
// where possible, they're shared with other Kafka consumers.
type AlertsEventsHandlerConfig struct {
	Topics      []string `envconfig:"KAFKA_ALERTS_TOPICS" required:"true"`
	GroupID     string   `envconfig:"KAFKA_ALERTS_CONSUMER_GROUP" required:"true"`
	Brokers     []string `envconfig:"KAFKA_BROKERS" required:"true"`
	TopicPrefix string   `envconfig:"KAFKA_TOPIC_PREFIX" required:"true"`
	// Username   string   `envconfig:"KAFKA_USERNAME" required:"true"`
	// Password   string   `envconfig:"KAFKA_PASSWORD" required:"true"`
	// RequireSSL bool     `envconfig:"KAFKA_REQUIRE_SSL" required:"true"`
	// Version    string   `envconfig:"KAFKA_VERSION" required:"true"`
}

func (c *AlertsEventsHandlerConfig) LoadFromEnv() error {
	if err := envconfig.Process("", c); err != nil {
		return err
	}
	return nil
}

type SaramaAlertsAdapter struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex

	AlertsClient *alerts.Client
	Config       *AlertsEventsHandlerConfig
	DataStore    *dataStoreMongo.Store
	SaramaConfig *sarama.Config

	Logger logInternal.Logger
}

func NewSaramaAlertsAdapter(cfg *AlertsEventsHandlerConfig, dataStore *dataStoreMongo.Store,
	alertsClient *alerts.Client, saramaCfg *sarama.Config, logger logInternal.Logger,
) *SaramaAlertsAdapter {
	return &SaramaAlertsAdapter{
		AlertsClient: alertsClient,
		Config:       cfg,
		DataStore:    dataStore,
		SaramaConfig: saramaCfg,
		Logger:       logger,
	}
}

func (a *SaramaAlertsAdapter) Run() error {
	ctx := logInternal.NewContextWithLogger(context.Background(), a.Logger)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	a.mu.Lock()
	a.ctx = ctx
	a.cancel = cancel
	a.mu.Unlock()

	consumerGroup, err := sarama.NewConsumerGroup(a.Config.Brokers, a.Config.GroupID, a.SaramaConfig)
	if err != nil {
		return err
	}
	alertsConsumer := &alertsEventsConsumer{
		alertsClient: a.AlertsClient,
		dataStore:    a.DataStore,
		logger:       a.Logger,
	}
	retryConsumer := &asyncevents.NTimesRetryingConsumer{
		Times:    5, // TODO magic number
		Consumer: alertsConsumer,
	}
	giveupConsumer := &asyncevents.GiveUpConsumer{Consumer: retryConsumer} // TODO: only for dev
	groupHandler := asyncevents.NewSaramaConsumerGroupHandler(giveupConsumer, time.Minute)
	consumer := asyncevents.NewSaramaEventsConsumer(consumerGroup, groupHandler, a.topics()...)
	return consumer.Run(ctx)
}

func (a *SaramaAlertsAdapter) topics() []string {
	prefixedTopics := make([]string, 0, len(a.Config.Topics))
	for _, topic := range a.Config.Topics {
		prefixedTopics = append(prefixedTopics, "default.data."+topic) // TODO magic char
	}
	return prefixedTopics
}

func (a *SaramaAlertsAdapter) Initialize() error { return nil }

func (a *SaramaAlertsAdapter) Terminate() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.ctx != nil && a.cancel != nil {
		a.cancel()
		if err := a.ctx.Err(); err != nil && !stderrors.Is(err, context.Canceled) {
			return err
		}
	}

	return nil
}

type alertsEventsConsumer struct {
	logger       logInternal.Logger
	alertsClient *alerts.Client
	dataStore    *dataStoreMongo.Store
}

func (c *alertsEventsConsumer) Consume(ctx context.Context,
	session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) (err error) {

	if msg == nil {
		c.logger.Info("received nil msg, will return nil")
		return nil
	}

	switch {
	case strings.HasSuffix(msg.Topic, ".data.alerts"):
		return c.consumeAlerts(ctx, session, msg)
	case strings.HasSuffix(msg.Topic, ".data.deviceData.alerts"):
		return c.consumeDeviceData(ctx, session, msg)
	default:
		c.logger.Infof("skipping message (wrong topic) %s", msg.Topic)
	}

	return nil
}

type alertsMessage struct {
	FullDocument alerts.Config `json:"fullDocument"`
}

func (c *alertsEventsConsumer) consumeAlerts(ctx context.Context,
	session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) error {

	am := &alertsMessage{}
	err := bson.UnmarshalExtJSON(msg.Value, false, am)
	if err != nil {
		return errors.Wrap(err, "unwrapping message")
	}
	cfg := am.FullDocument
	c.logger.WithField("cfg", cfg).Debug("consuming an alerts message")

	dataRepo := c.dataStore.NewDataRepository()
	// TODO: pass it off to the evaulator
	err = evalalerts.Evaluate(ctx, dataRepo, c.alertsClient, cfg.FollowedUserID, cfg.UserID)
	if err != nil {
		return errors.Wrap(err, "Evaluate")
	}
	session.MarkMessage(msg, "")
	c.logger.WithField("message", msg).Debug("marked")
	return nil
}

func (c *alertsEventsConsumer) consumeDeviceData(ctx context.Context,
	session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) error {

	var err error
	bloodDoc := &bloodDocument{}
	err = bson.UnmarshalExtJSON(msg.Value, false, bloodDoc)
	if err != nil {
		c.logger.WithError(err).Debug("blood failed to unmarshal")
	}
	blood := bloodDoc.FullDocument
	c.logger.WithField("blood", blood).Debug("device data")

	// find users that have alerts for this user
	// there's a gatekeeper url for this, but I think that I probably shouldn't be using that..?

	// loop over said users
	c.logger.Debug("TODO: find users that have alerts for this user")

	session.MarkMessage(msg, "")
	c.logger.WithField("message", msg).Debug("marked")
	return nil
}

type bloodDocument struct {
	FullDocument blood.Blood `json:"fullDocument"`
}
