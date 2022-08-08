package mongo

import (
	"context"
	"fmt"
	"gitlab.com/g6834/team26/auth/pkg/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	DB *mongo.Client
	c  *config.Config
}

func New(ctx context.Context, config *config.Config) (*Database, error) {
	opt := options.Credential{
		AuthMechanism: config.Server.AuthMongoMech,
		AuthSource:    config.Server.MongoDbName,
		Username:      config.Server.MongoUserName,
		Password:      config.Server.MongoUserPass,
	}
	c := options.Client().ApplyURI(config.Server.AuthorizationDBConnectionString).SetAuth(opt)

	pool, err := mongo.Connect(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("create pool failed: %w", err)
	}

	err = pool.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("ping mongo failed: %w", err)
	}

	return &Database{DB: pool, c: config}, nil
}

func (db *Database) UserCollection() *mongo.Collection {
	return db.DB.Database(db.c.Server.MongoDbName).Collection(db.c.Server.MongoCollectionUser)
}

func (db *Database) Disconnect(ctx context.Context) error {
	return db.DB.Disconnect(ctx)
}
