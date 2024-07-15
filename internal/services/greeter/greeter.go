package greeter

import (
	"context"
	"time"

	di "github.com/fluffy-bunny/fluffy-dozm-di"
	contracts_config "github.com/fluffy-bunny/fluffycore-starterkit/internal/contracts/config"
	fluffycore_contracts_somedisposable "github.com/fluffy-bunny/fluffycore-starterkit/internal/contracts/somedisposable"
	proto_helloworld "github.com/fluffy-bunny/fluffycore-starterkit/proto/helloworld"
	fluffycore_contracts_GRPCClientFactory "github.com/fluffy-bunny/fluffycore/contracts/GRPCClientFactory"
	endpoint "github.com/fluffy-bunny/fluffycore/contracts/endpoint"
	fluffycore_grpcclient "github.com/fluffy-bunny/fluffycore/grpcclient"
	grpc_gateway_runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	zerolog "github.com/rs/zerolog"
	otel "go.opentelemetry.io/otel"
	attribute "go.opentelemetry.io/otel/attribute"
	trace "go.opentelemetry.io/otel/trace"
	grpc "google.golang.org/grpc"
)

type (
	service struct {
		proto_helloworld.GreeterFluffyCoreServer

		config               *contracts_config.Config
		scopedSomeDisposable fluffycore_contracts_somedisposable.IScopedSomeDisposable
		grpcClientFactory    fluffycore_contracts_GRPCClientFactory.IGRPCClientFactory
	}
)

var (
	stemService = (*service)(nil)
)
var tracer = otel.Tracer("grpc-example")

func init() {
	var _ proto_helloworld.IFluffyCoreGreeterServer = (*service)(nil)
	var _ endpoint.IEndpointRegistration = (*service)(nil)
}

func (s *service) Ctor(
	config *contracts_config.Config,
	grpcClientFactory fluffycore_contracts_GRPCClientFactory.IGRPCClientFactory,
	scopedSomeDisposable fluffycore_contracts_somedisposable.IScopedSomeDisposable) proto_helloworld.IFluffyCoreGreeterServer {
	return &service{
		config:               config,
		grpcClientFactory:    grpcClientFactory,
		scopedSomeDisposable: scopedSomeDisposable,
	}
}
func (s *service) RegisterFluffyCoreHandler(gwmux *grpc_gateway_runtime.ServeMux, conn *grpc.ClientConn) {
	proto_helloworld.RegisterGreeterHandler(context.Background(), gwmux, conn)
}

func AddGreeterService(builder di.ContainerBuilder) {
	proto_helloworld.AddGreeterServerWithExternalRegistration(builder,
		stemService.Ctor,
		func() endpoint.IEndpointRegistration {
			return &service{}
		})
}

func (s *service) SayHello(ctx context.Context, request *proto_helloworld.HelloRequest) (*proto_helloworld.HelloReply, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("SayHello")

	s.workHard(ctx)
	time.Sleep(50 * time.Millisecond)

	opts := []fluffycore_grpcclient.GrpcClientOption{
		fluffycore_grpcclient.WithHost("localhost"),
		fluffycore_grpcclient.WithPort(50051),
		fluffycore_grpcclient.WithInsecure(true),
	}

	grpcClient, err := s.grpcClientFactory.NewGrpcClient(opts...)
	if err != nil {
		log.Error().Err(err).Msg("Creating gRPC client")
		return nil, err
	}
	defer grpcClient.Close()

	cli := proto_helloworld.NewGreeterClient(grpcClient.GetConnection())
	reply, err := cli.SayHelloDownstream(ctx, request)
	return reply, err
}
func (s *service) SayHelloDownstream(ctx context.Context, request *proto_helloworld.HelloRequest) (*proto_helloworld.HelloReply, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("SayHelloDownstream")

	return &proto_helloworld.HelloReply{
		Message: "Hello " + request.Name,
	}, nil
}
func (s *service) workHard(ctx context.Context) {
	_, span := tracer.Start(ctx, "workHard",
		trace.WithAttributes(attribute.String("extra.key", "extra.value")))
	defer span.End()

	time.Sleep(50 * time.Millisecond)
}
