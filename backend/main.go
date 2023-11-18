package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/0xPolygonID/onchain-nonmerklized-issuer-demo/backend/common"
	"github.com/0xPolygonID/onchain-nonmerklized-issuer-demo/backend/handlers"
	"github.com/0xPolygonID/onchain-nonmerklized-issuer-demo/backend/repository"
	"github.com/0xPolygonID/onchain-nonmerklized-issuer-demo/backend/services"
	ethcomm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/iden3/contracts-abi/state/go/abi"
	"github.com/iden3/go-jwz/v2"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.ngrok.com/ngrok"
	ngrokCfg "golang.ngrok.com/ngrok/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var isDevFlag = flag.Bool("dev", false, "run in dev mode")

func main() {
	flag.Parse()

	cr, err := initRepository()
	if err != nil {
		log.Fatal("failed connect to mongodb:", err)
	}
	pkg, err := initPakcer()
	if err != nil {
		log.Fatal("failed init packer:", err)
	}

	h := handlers.Handlers{
		CredentialService: &services.ClaimService{
			CredentialRepository: cr,
		},
		Packager: pkg,
	}

	r := chi.NewRouter()

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://*", "https://*", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	r.Use(corsMiddleware.Handler)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(5 * time.Minute))

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Post("/identities/{identifier}/claims", h.CreateClaim)
			r.Get("/identities/{identifier}/claims/{credentialId}", h.GetUserVCByID)
			r.Get("/identities/{identifier}/claims/offer", h.GetOffer)
			r.Post("/agent", h.Agent)
		})
	})

	if *isDevFlag {
		go func() {
			err := runNgrok(r)
			if err != nil {
				log.Fatalf("can't run ngrok, err: %v", err)
			}
		}()
	}

	http.ListenAndServe(":3333", r)
}

func runNgrok(r chi.Router) error {
	tun, err := ngrok.Listen(
		context.Background(),
		ngrokCfg.HTTPEndpoint(),
		ngrok.WithAuthtokenFromEnv(),
	)
	if err != nil {
		return err
	}
	url := tun.URL()
	common.ExternalServerHost = url
	fmt.Println("ngrok url: ", url)
	return http.Serve(tun, r)
}

func initRepository() (*repository.CredentialRepository, error) {
	tM := reflect.TypeOf(bson.M{})
	reg := bson.NewRegistryBuilder().RegisterTypeMapEntry(bsontype.EmbeddedDocument, tM).Build()

	fmt.Println("Connecting to MongoDB: ", common.MongoDBHost)
	opts := options.Client().ApplyURI(common.MongoDBHost).SetRegistry(reg)
	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to mongo")
	}
	if err = client.Ping(context.Background(), nil); err != nil {
		return nil, errors.Wrap(err, "failed to ping mongo")
	}
	rep, err := repository.NewCredentialRepository(client.Database("credentials"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create credential repository")
	}
	return rep, nil
}

func initPakcer() (*iden3comm.PackageManager, error) {
	stateContracts := map[string]*abi.State{}
	for chainPrefix, resolverSetting := range common.ResolverSettings {
		client, err := ethclient.Dial(resolverSetting.NetworkURL)
		if err != nil {
			return nil, err
		}
		add := ethcomm.HexToAddress(resolverSetting.ContractState)
		stateContract, err := abi.NewState(add, client)
		if err != nil {
			return nil, err
		}
		stateContracts[chainPrefix] = stateContract
	}

	authV2VerificationKey, err := os.ReadFile(common.AuthV2VerificationKeyPath)
	if err != nil {
		return nil, err
	}
	unpakOpts := map[jwz.ProvingMethodAlg]packers.VerificationParams{
		jwz.AuthV2Groth16Alg: {
			Key:            authV2VerificationKey,
			VerificationFn: services.StateVerificationHandler(stateContracts),
		},
	}
	zkpPacker := packers.NewZKPPacker(
		nil,
		unpakOpts,
	)
	packer := iden3comm.NewPackageManager()
	packer.RegisterPackers(
		zkpPacker,
		&packers.PlainMessagePacker{},
	)

	return packer, nil
}
