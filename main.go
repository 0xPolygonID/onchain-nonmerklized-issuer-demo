package main

import (
	"log"
	"net/http"
	"strconv"

	auth "github.com/iden3/go-iden3-auth/v2"
	"github.com/iden3/go-iden3-auth/v2/loaders"
	"github.com/iden3/go-iden3-auth/v2/pubsignals"
	"github.com/iden3/go-iden3-auth/v2/state"
	"github.com/iden3/go-service-template/config"
	"github.com/iden3/go-service-template/pkg/logger"
	httprouter "github.com/iden3/go-service-template/pkg/router/http"
	"github.com/iden3/go-service-template/pkg/router/http/handlers"
	"github.com/iden3/go-service-template/pkg/services/authentication"
	"github.com/iden3/go-service-template/pkg/services/issuer"
	"github.com/iden3/go-service-template/pkg/services/system"
	"github.com/iden3/go-service-template/pkg/shutdown"
	httptransport "github.com/iden3/go-service-template/pkg/transport/http"
	"github.com/pkg/errors"
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

	httpserver := newHTTPServer(
		cfg,
		authverifier,
		cfg.Issuers,
	)
	newShutdownManager(httpserver).HandleShutdownSignal()
}

func newHTTPServer(
	cfg *config.Config,
	authverifier *auth.Verifier,
	issuers []string,
) *httptransport.Server {
	// init services
	authenticationService := authentication.NewAuthenticationService(
		authverifier,
	)
	issuerService := issuer.NewIssuerService(
		issuers,
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
	issuerHandlers := handlers.NewIssuerHandlers(
		issuerService,
	)

	// init routers
	h := httprouter.NewHandlers(
		systemHandlers,
		authenticationHandlers,
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
		80002: "polygon:amoy",
		21000: "privado:main",
		21001: "privado:test",
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
