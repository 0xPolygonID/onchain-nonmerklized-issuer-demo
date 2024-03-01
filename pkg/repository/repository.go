package repository

import (
	"context"
	"log/slog"

	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/iden3/go-service-template/pkg/logger"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CredentialRepository struct {
	db   *mongo.Database
	coll *mongo.Collection
}

func NewCredentialRepository(db *mongo.Database) (*CredentialRepository, error) {
	err := db.CreateCollection(
		context.Background(),
		"credentials",
		options.CreateCollection().SetCollation(
			&options.Collation{
				Locale:   "en",
				Strength: 2,
			},
		),
	)
	if err != nil {
		var comErr mongo.CommandError
		if errors.As(err, &comErr) && comErr.Code == 48 {
			// collection already exists
			logger.Info("collection already exists", slog.String("collection", "credentials"))
		} else {
			return nil, errors.Wrapf(err, "failed to create collection 'credentials'")
		}
	}

	collection := db.Collection("credentials")
	_, err = collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{
				{
					Key:   "id",
					Value: 1,
				},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{
					Key:   "issuer",
					Value: 1,
				},
			},
		},
		{
			Keys: bson.D{
				{
					Key:   "credentialsubject.id",
					Value: 1,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &CredentialRepository{
		db:   db,
		coll: collection,
	}, nil
}

func (cs *CredentialRepository) Create(
	ctx context.Context,
	vc *verifiable.W3CCredential,
) (string, error) {
	model, err := NewCredentailModelFromW3C(vc)
	if err != nil {
		return "", err
	}
	_, err = cs.coll.InsertOne(ctx, model)
	if err != nil {
		return "", errors.Wrapf(err, "failed to insert credential")
	}
	return vc.ID, nil
}

func (cs *CredentialRepository) GetByID(
	ctx context.Context,
	issuer string,
	credentialID string,
) (*verifiable.W3CCredential, error) {
	filter := bson.M{"issuer": issuer, "id": bson.M{"$regex": credentialID}}
	res := cs.coll.FindOne(ctx, filter)
	if res.Err() != nil {
		return nil, res.Err()
	}
	var model credentialModel
	if err := res.Decode(&model); err != nil {
		return nil,
			errors.Wrap(err, "failed to decode credential")
	}
	credential, err := model.ToW3C()
	if err != nil {
		return nil,
			errors.Wrap(err, "failed to convert credential model to W3C")
	}
	return credential, nil
}
