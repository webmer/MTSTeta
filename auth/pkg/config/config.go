package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Server serverConfig
}

type serverConfig struct {
	Port                            string `envconfig:"PORT" default:"3000"`
	GRPCPort                        string `envconfig:"GRPC_PORT" default:"4000"`
	Profiling                       bool   `envconfig:"PROFILING" default:"false"`
	AuthorizationDBConnectionString string `envconfig:"AUTHORIZATION_DB_CONNECTION_STRING" default:""`
	AuthMongoMech                   string `envconfig:"AUTH_MONGO_MECH" default:""`
	MongoDbName                     string `envconfig:"MONGO_DB_NAME" default:""`
	MongoUserName                   string `envconfig:"MONGO_USER_NAME" default:""`
	MongoUserPass                   string `envconfig:"MONGO_USER_PASS" default:""`
	MongoCollectionUser             string `envconfig:"MONGO_COLLECTION_USER" default:"user"`
	AccessSecret                    string `envconfig:"ACCESS_SECRET" default:"Basic access secret"`
	RefreshSecret                   string `envconfig:"REFRESH_SECRET" default:"Basic refresh secret"`
	AccessCookie                    string `envconfig:"ACCESS_COOKIE" default:"access_token"`
	RefreshCookie                   string `envconfig:"REFRESH_COOKIE" default:"refresh_token"`
}

func New() (*Config, error) {
	var c Config

	err := envconfig.Process("", &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
