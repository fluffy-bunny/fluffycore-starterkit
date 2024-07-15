package config

import (
	fluffycore_contracts_config "github.com/fluffy-bunny/fluffycore/contracts/config"
	fluffycore_contracts_ddprofiler "github.com/fluffy-bunny/fluffycore/contracts/ddprofiler"
	fluffycore_contracts_otel "github.com/fluffy-bunny/fluffycore/contracts/otel"
)

type (
	JWTValidators struct {
		Issuers  []string `json:"issuers"`
		JWKSURLS []string `json:"jwksUrls"`
	}
	ConfigFiles struct {
		ClientPath string `json:"clientPath"`
	}
)
type EchoConfig struct {
	Port int `json:"port"`
}
type Config struct {
	fluffycore_contracts_config.CoreConfig `mapstructure:",squash"`

	ConfigFiles      ConfigFiles                             `json:"configFiles"`
	CustomString     string                                  `json:"customString"`
	SomeSecret       string                                  `json:"someSecret" redact:"true"`
	OAuth2Port       int                                     `json:"oauth2Port"`
	JWTValidators    JWTValidators                           `json:"jwtValidators"`
	DDProfilerConfig *fluffycore_contracts_ddprofiler.Config `json:"ddProfilerConfig"`
	Echo             EchoConfig                              `json:"echo"`
	OTELConfig       *fluffycore_contracts_otel.OTELConfig   `json:"otelConfig"`
}

// ConfigDefaultJSON default json
var ConfigDefaultJSON = []byte(`
{
	"APPLICATION_NAME": "in-environment",
	"APPLICATION_ENVIRONMENT": "in-environment",
	"PRETTY_LOG": false,
	"LOG_LEVEL": "info",
	"PORT": 50051,
	"REST_PORT": 50052,
	"oauth2Port": 50053,
	"customString": "some default value",
	"someSecret": "password",
	"GRPC_GATEWAY_ENABLED": true,
	"jwtValidators": {},
	"configFiles": {
		"clientPath": "./config/clients.json"
	},
	"ddProfilerConfig": {
		"enabled": false,
		"serviceName": "in-environment",
		"applicationEnvironment": "in-environment",
		"version": "1.0.0"
	},
	"echo": {
		"port": 9044 
	},
    "otelConfig": {
        "serviceName": "in-environment",
        "tracingConfig": {
            "enabled": false,
            "endpointType": "http",
            "endpoint": "localhost:4318"
        },
        "metricConfig": {
            "enabled": false,
            "endpointType": "http",
            "intervalSeconds": 10,
            "endpoint": "localhost:4318",
            "runtimeEnabled": false,
            "hostEnabled": false
        } 
    }

  }
`)
