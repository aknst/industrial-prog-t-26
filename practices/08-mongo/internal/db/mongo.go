package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDeps struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// ConnectMongo подключается к MongoDB и проверяет соединение командой ping.
func ConnectMongo(ctx context.Context, uri, dbName string) (*MongoDeps, error) {
	dialCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cli, err := mongo.Connect(dialCtx, options.Client().ApplyURI(uri).SetServerSelectionTimeout(5*time.Second))
	if err != nil {
		return nil, err
	}
	pingCtx, cancelPing := context.WithTimeout(ctx, 3*time.Second)
	defer cancelPing()
	if err := cli.Ping(pingCtx, nil); err != nil {
		_ = cli.Disconnect(ctx)
		return nil, err
	}
	return &MongoDeps{Client: cli, Database: cli.Database(dbName)}, nil
}
