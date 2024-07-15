package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	di "github.com/fluffy-bunny/fluffy-dozm-di"
	internal_auth "github.com/fluffy-bunny/fluffycore-starterkit/internal/auth"
	contracts_config "github.com/fluffy-bunny/fluffycore-starterkit/internal/contracts/config"
	myechoserver "github.com/fluffy-bunny/fluffycore-starterkit/internal/myechoserver"
	services "github.com/fluffy-bunny/fluffycore-starterkit/internal/services"
	services_greeter "github.com/fluffy-bunny/fluffycore-starterkit/internal/services/greeter"
	services_health "github.com/fluffy-bunny/fluffycore-starterkit/internal/services/health"
	services_mystream "github.com/fluffy-bunny/fluffycore-starterkit/internal/services/mystream"
	services_somedisposable "github.com/fluffy-bunny/fluffycore-starterkit/internal/services/somedisposable"
	internal_version "github.com/fluffy-bunny/fluffycore-starterkit/internal/version"
	fluffycore_async "github.com/fluffy-bunny/fluffycore/async"
	fluffycore_contracts_GRPCClientFactory "github.com/fluffy-bunny/fluffycore/contracts/GRPCClientFactory"
	fluffycore_contracts_ddprofiler "github.com/fluffy-bunny/fluffycore/contracts/ddprofiler"
	fluffycore_contracts_middleware_auth_jwt "github.com/fluffy-bunny/fluffycore/contracts/middleware/auth/jwt"
	fluffycore_contracts_otel "github.com/fluffy-bunny/fluffycore/contracts/otel"
	fluffycore_contracts_runtime "github.com/fluffy-bunny/fluffycore/contracts/runtime"
	core_echo_runtime "github.com/fluffy-bunny/fluffycore/echo/runtime"
	fluffycore_middleware_auth_jwt "github.com/fluffy-bunny/fluffycore/middleware/auth/jwt"
	mocks_contracts_oauth2 "github.com/fluffy-bunny/fluffycore/mocks/contracts/oauth2"
	mocks_oauth2_echo "github.com/fluffy-bunny/fluffycore/mocks/oauth2/echo"
	fluffycore_runtime_otel "github.com/fluffy-bunny/fluffycore/runtime/otel"
	fluffycore_services_GRPCClientFactory "github.com/fluffy-bunny/fluffycore/services/GRPCClientFactory"
	fluffycore_services_ddprofiler "github.com/fluffy-bunny/fluffycore/services/ddprofiler"
	fluffycore_utils_redact "github.com/fluffy-bunny/fluffycore/utils/redact"
	async "github.com/reugn/async"
	zerolog "github.com/rs/zerolog"
)

type (
	startup struct {
		*fluffycore_runtime_otel.FluffyCoreOTELStartup

		configOptions *fluffycore_contracts_runtime.ConfigOptions
		config        *contracts_config.Config

		mockOAuth2Server       *mocks_oauth2_echo.MockOAuth2Service
		mockOAuth2ServerFuture async.Future[fluffycore_async.AsyncResponse]
		ddProfiler             fluffycore_contracts_ddprofiler.IDataDogProfiler
		myEchoServerFuture     async.Future[fluffycore_async.AsyncResponse]
		myEchoServerRuntime    *core_echo_runtime.Runtime
	}
)

func NewStartup() fluffycore_contracts_runtime.IStartup {
	return &startup{
		FluffyCoreOTELStartup: fluffycore_runtime_otel.NewFluffyCoreOTELStartup(&fluffycore_runtime_otel.FluffyCoreOTELStartupConfig{
			FuncAuthGetEntryPointConfigs: internal_auth.BuildGrpcEntrypointPermissionsClaimsMap,
		}),
	}
}

func (s *startup) GetConfigOptions() *fluffycore_contracts_runtime.ConfigOptions {
	s.config = &contracts_config.Config{}
	s.configOptions = &fluffycore_contracts_runtime.ConfigOptions{
		Destination: s.config,
		RootConfig:  contracts_config.ConfigDefaultJSON,
	}
	return s.configOptions
}
func (s *startup) ConfigureServices(ctx context.Context, builder di.ContainerBuilder) {
	log := zerolog.Ctx(ctx).With().Str("method", "Configure").Logger()
	dst, err := fluffycore_utils_redact.CloneAndRedact(s.configOptions.Destination)
	if err != nil {
		panic(err)
	}
	log.Info().Interface("config", dst).Msg("config")
	config := s.configOptions.Destination.(*contracts_config.Config)
	// need to set the OTEL Config in the base startup
	if config.OTELConfig == nil {
		config.OTELConfig = &fluffycore_contracts_otel.OTELConfig{}
	}
	config.OTELConfig.ServiceName = config.ApplicationName
	s.FluffyCoreOTELStartup.SetConfig(config.OTELConfig)
	// add grpcclient factory that is config aware.  Will make sure that you get one that has otel tracing if enabled.
	fluffycore_contracts_GRPCClientFactory.AddGRPCClientConfig(builder,
		&fluffycore_contracts_GRPCClientFactory.GRPCClientConfig{
			OTELTracingEnabled: config.OTELConfig.TracingConfig.Enabled,
		})
	fluffycore_services_GRPCClientFactory.AddSingletonIGRPCClientFactory(builder)

	config.DDProfilerConfig.ApplicationEnvironment = config.ApplicationEnvironment
	config.DDProfilerConfig.ServiceName = config.ApplicationName
	config.DDProfilerConfig.Version = internal_version.Version()
	di.AddInstance[*fluffycore_contracts_ddprofiler.Config](builder, config.DDProfilerConfig)
	di.AddInstance[*contracts_config.Config](builder, config)

	services.ConfigureServices(builder)
	fluffycore_services_ddprofiler.AddSingletonIProfiler(builder)
	services_health.AddHealthService(builder)
	services_greeter.AddGreeterService(builder)
	services_somedisposable.AddScopedSomeDisposable(builder)
	services_mystream.AddMyStreamService(builder)
	issuerConfigs := &fluffycore_contracts_middleware_auth_jwt.IssuerConfigs{}
	for idx := range s.config.JWTValidators.Issuers {
		issuerConfigs.IssuerConfigs = append(issuerConfigs.IssuerConfigs,
			&fluffycore_contracts_middleware_auth_jwt.IssuerConfig{
				OAuth2Config: &fluffycore_contracts_middleware_auth_jwt.OAuth2Config{
					Issuer:  s.config.JWTValidators.Issuers[idx],
					JWKSUrl: s.config.JWTValidators.JWKSURLS[idx],
				},
			})
	}
	fluffycore_middleware_auth_jwt.AddValidators(builder, issuerConfigs)
}

// OnPreServerStartup ...
func (s *startup) OnPreServerStartup(ctx context.Context) error {
	log := zerolog.Ctx(ctx).With().Str("method", "OnPreServerStartup").Logger()

	err := s.FluffyCoreOTELStartup.OnPreServerStartup(ctx)
	if err != nil {
		return err
	}

	clientsJSON, err := os.ReadFile(s.config.ConfigFiles.ClientPath)
	var clients []mocks_contracts_oauth2.Client
	if err != nil {
		return err
	}
	err = json.Unmarshal(clientsJSON, &clients)
	if err != nil {
		return err
	}

	log.Info().Interface("clients", clients).Msg("clients")
	s.mockOAuth2Server = mocks_oauth2_echo.NewOAuth2TestServer(&mocks_contracts_oauth2.MockOAuth2Config{
		Clients: clients,
	})
	s.myEchoServerRuntime = core_echo_runtime.New(myechoserver.NewStartup())
	s.myEchoServerFuture = fluffycore_async.ExecuteWithPromiseAsync(func(promise async.Promise[fluffycore_async.AsyncResponse]) {
		var err error
		defer func() {
			promise.Success(&fluffycore_async.AsyncResponse{
				Message: "End Serve - echoServer",
				Error:   err,
			})
		}()
		log.Info().Msg("echoServer starting up")
		err = s.myEchoServerRuntime.Run()
		if err != nil {
			log.Error().Err(err).Msg("failed to start server")
		}
	})
	s.mockOAuth2ServerFuture = fluffycore_async.ExecuteWithPromiseAsync(func(promise async.Promise[fluffycore_async.AsyncResponse]) {
		var err error
		defer func() {
			promise.Success(&fluffycore_async.AsyncResponse{
				Message: "End Serve - mockOAuth2Server",
				Error:   err,
			})
		}()
		log.Info().Msg("mockOAuth2Server starting up")
		err = s.mockOAuth2Server.Start(fmt.Sprintf(":%d", s.config.OAuth2Port))
		if err != nil && http.ErrServerClosed == err {
			err = nil
		}
		if err != nil {
			log.Error().Err(err).Msg("failed to start server")
		}
	})

	s.ddProfiler, err = di.TryGet[fluffycore_contracts_ddprofiler.IDataDogProfiler](s.RootContainer)
	if err != nil {
		log.Warn().Err(err).Msg("No DataDog Profiler")
	} else {
		s.ddProfiler.Start(ctx)
	}
	return nil
}

// OnPreServerShutdown ...
func (s *startup) OnPreServerShutdown(ctx context.Context) {
	log := zerolog.Ctx(ctx).With().Str("method", "OnPreServerShutdown").Logger()
	log.Info().Msg("mockOAuth2Server shutting down")
	s.mockOAuth2Server.Shutdown(ctx)
	s.mockOAuth2ServerFuture.Join()
	log.Info().Msg("mockOAuth2Server shutdown complete")

	log.Info().Msg("myEchoServerRuntime shutting down")
	s.myEchoServerRuntime.Stop()
	s.myEchoServerFuture.Join()
	log.Info().Msg("myEchoServerRuntime shutdown complete")

	log.Info().Msg("Stopping Datadog Tracer and Profiler")
	if s.ddProfiler != nil {
		s.ddProfiler.Stop(ctx)
	}
	log.Info().Msg("Datadog Tracer and Profiler stopped")

	log.Info().Msg("FluffyCoreOTELStartup stopped")
	s.FluffyCoreOTELStartup.OnPreServerShutdown(ctx)
}
