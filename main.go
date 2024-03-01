package main

import (
	"context"
	_ "embed"
	"log"
	"net/http"
	"reflect"
	"strconv"

	"github.com/ethereum/go-ethereum/ethclient"
	auth "github.com/iden3/go-iden3-auth/v2"
	"github.com/iden3/go-iden3-auth/v2/loaders"
	"github.com/iden3/go-iden3-auth/v2/pubsignals"
	"github.com/iden3/go-iden3-auth/v2/state"
	schemaLoaders "github.com/iden3/go-schema-processor/v2/loaders"
	"github.com/iden3/go-service-template/config"
	"github.com/iden3/go-service-template/pkg/logger"
	"github.com/iden3/go-service-template/pkg/packagemanager"
	"github.com/iden3/go-service-template/pkg/repository"
	httprouter "github.com/iden3/go-service-template/pkg/router/http"
	"github.com/iden3/go-service-template/pkg/router/http/handlers"
	"github.com/iden3/go-service-template/pkg/services/authentication"
	"github.com/iden3/go-service-template/pkg/services/iden3comm"
	"github.com/iden3/go-service-template/pkg/services/issuer"
	"github.com/iden3/go-service-template/pkg/services/system"
	"github.com/iden3/go-service-template/pkg/shutdown"
	httptransport "github.com/iden3/go-service-template/pkg/transport/http"
	libiden3comm "github.com/iden3/iden3comm/v2"
	"github.com/piprate/json-gold/ld"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	//go:embed cache/w3cCredentialSchemaV1.json
	w3cCredentialSchemaV1 []byte
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	if err = logger.SetDefaultLogger(
		cfg.Log.Environment,
		cfg.Log.LogLevel(),
	); err != nil {
		log.Fatalf("failed to set default logger: %v", err)
	}

	// init dependencies
	authverifier, err := initializationAuthVerifier(cfg)
	if err != nil {
		logger.WithError(err).Fatal("error creating auth verifier")
	}
	pkgmanager, err := packagemanager.NewPackageManager(
		cfg.SupportedRPC,
		cfg.SupportedStateContracts,
		cfg.KeysDirPath,
	)
	if err != nil {
		logger.WithError(err).Fatal("error creating package manager")
	}

	reg := bson.NewRegistry()
	reg.RegisterTypeMapEntry(bson.TypeEmbeddedDocument, reflect.TypeOf(bson.M{}))
	opts := options.Client().ApplyURI(cfg.MongoDBConnectionString).SetRegistry(reg)
	mongoClient, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		logger.WithError(err).Fatal("error connecting to mongodb")
	}
	credentialRepository, err := repository.NewCredentialRepository(mongoClient.Database("credentials"))
	if err != nil {
		logger.WithError(err).Fatal("error creating credential repository")
	}

	ethclients, err := initializationEthClients(cfg.SupportedRPC)
	if err != nil {
		logger.WithError(err).Fatal("error creating eth clients")
	}

	documentLoader, err := initDocumentLoaderWithCache()
	if err != nil {
		logger.WithError(err).Fatal("error creating document loader")
	}

	httpserver := newHTTPServer(
		cfg,
		authverifier,
		pkgmanager,
		credentialRepository,
		ethclients,
		cfg.Issuers,
		documentLoader,
	)
	newShutdownManager(httpserver).HandleShutdownSignal()
}

func newHTTPServer(
	cfg *config.Config,
	authverifier *auth.Verifier,
	pkgmanager *libiden3comm.PackageManager,
	credentialRepository *repository.CredentialRepository,
	ethclients map[string]*ethclient.Client,
	issuers []string,
	documentLoader ld.DocumentLoader,
) *httptransport.Server {
	// init services
	authenticationService := authentication.NewAuthenticationService(
		authverifier,
	)
	iden3commService := iden3comm.NewIden3commService(
		pkgmanager,
		credentialRepository,
	)
	issuerService := issuer.NewIssuerService(
		credentialRepository,
		ethclients,
		issuers,
		documentLoader,
	)

	// init handlers
	systemHandlers := handlers.NewSystemHandler(
		system.NewReadinessService(),
		system.NewLivenessService(),
	)
	authenticationHandlers := handlers.NewAuthenticationHandlers(
		cfg.ExternalHost,
		authenticationService,
	)
	iden3commHandlers := handlers.NewIden3commHandlers(
		iden3commService,
	)
	issuerHandlers := handlers.NewIssuerHandlers(
		issuerService,
		cfg.ExternalHost,
	)

	// init routers
	h := httprouter.NewHandlers(
		systemHandlers,
		authenticationHandlers,
		iden3commHandlers,
		issuerHandlers,
	)
	routers := h.NewRouter(
		httprouter.WithOrigins(cfg.HTTPServer.Origins),
	)

	// run http server
	httpserver := httptransport.New(
		routers,
		httptransport.WithWriteTimeout(0),
		httptransport.WithHost(cfg.HTTPServer.Host, cfg.HTTPServer.Port),
	)

	go func() {
		err := httpserver.Start()
		if errors.Is(err, http.ErrServerClosed) {
			logger.Info("HTTP server closed by request")
		} else {
			logger.WithError(err).Fatal("http server closed with error")
		}
	}()

	return httpserver
}

func newShutdownManager(toclose ...shutdown.Shutdown) *shutdown.Manager {
	m := shutdown.NewManager()
	for _, s := range toclose {
		m.Register(s)
	}
	return m
}

func chainIDToDIDPrefix(chainID int) string {
	p := map[int]string{
		137:   "polygon:main",
		80001: "polygon:mumbai",
	}
	return p[chainID]
}

func initializationAuthVerifier(configuration *config.Config) (*auth.Verifier, error) {
	var resolvers = make(map[string]pubsignals.StateResolver, len(configuration.SupportedStateContracts))
	for network, contractAddress := range configuration.SupportedStateContracts {
		rpcURL, ok := configuration.SupportedRPC[network]
		if !ok {
			return nil, errors.Errorf("no rpc for network %s", network)
		}
		chainID, _ := strconv.Atoi(network)
		resolvers[chainIDToDIDPrefix(chainID)] = state.NewETHResolver(rpcURL, contractAddress)
	}

	verifier, err := auth.NewVerifier(loaders.FSKeyLoader{
		Dir: configuration.KeysDirPath,
	}, resolvers)
	if err != nil {
		return nil, errors.Errorf("error creating verifier: %v", err)
	}
	return verifier, nil
}

func initializationEthClients(supportedRPC map[string]string) (map[string]*ethclient.Client, error) {
	ethClients := make(map[string]*ethclient.Client, len(supportedRPC))
	for network, rpcURL := range supportedRPC {
		ec, err := ethclient.Dial(rpcURL)
		if err != nil {
			return nil, errors.Errorf("error creating eth client: %v", err)
		}
		ethClients[network] = ec
	}
	return ethClients, nil
}

func initDocumentLoaderWithCache() (ld.DocumentLoader, error) {
	opts := schemaLoaders.WithEmbeddedDocumentBytes(
		"https://www.w3.org/2018/credentials/v1",
		w3cCredentialSchemaV1,
	)
	memoryCacheEngine, err := schemaLoaders.NewMemoryCacheEngine(opts)
	if err != nil {
		return nil, err
	}
	l := schemaLoaders.NewDocumentLoader(nil, "", schemaLoaders.WithCacheEngine(memoryCacheEngine))
	return l, nil
}
