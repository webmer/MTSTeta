package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Server serverConfig
}

type serverConfig struct {
	Port                            string `envconfig:"PORT" default:"3000"`
	GRPCPort                        string `envconfig:"GRPC_AUTH_PORT" default:"4000"`
	Profiling                       bool   `envconfig:"PROFILING" default:"false"`
	AuthorizationDBConnectionString string `envconfig:"AUTHORIZATION_DB_CONNECTION_STRING" default:""`
	AccessCookie                    string `envconfig:"ACCESS_COOKIE" default:"access_token"`
	RefreshCookie                   string `envconfig:"REFRESH_COOKIE" default:"refresh_token"`
	KafkaUrl                        string `envconfig:"KAFKA_URL" default:""`
	KafkaAnalyticTopic              string `envconfig:"KAFKA_ANALYTIC_TOPIC" default:""`
	KafkaGroupId                    string `envconfig:"KAFKA_GROUP_ID" default:""`
}

func New() (*Config, error) {
	var c Config

	err := envconfig.Process("", &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
