package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/uber-go/tally"
	"github.com/uber-go/tally/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"

	//"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
	"movieexample.com/gen"
	"movieexample.com/movie/internal/controller/movie"
	metadataGateway "movieexample.com/movie/internal/gateway/metadata/grpc"
	ratingGateway "movieexample.com/movie/internal/gateway/rating/grpc"
	grpchandler "movieexample.com/movie/internal/handler/grpc"
	"movieexample.com/pkg/discovery"
	"movieexample.com/pkg/discovery/consul"
	"movieexample.com/pkg/tracing"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	f, err := os.Open("base.yaml")
	if err != nil {
		logger.Fatal("Failed to open configuration", zap.Error(err))
	}
	var cfg config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		logger.Fatal("Failed to parse configuration", zap.Error(err))
	}
	port := cfg.API.Port

	logger.Info("Starting the movie service", zap.Int("port", port))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tp, err := tracing.NewJaegerProvider(cfg.Jaeger.URL, cfg.ServiceName)
	if err != nil {
		logger.Fatal("Failed to initialize Jaeger provider", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal("Failed to shut down Jaeger prodiver", zap.Error(err))
		}
	}()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	reporter := prometheus.NewReporter(prometheus.Options{})
	scope, closer := tally.NewRootScope(
		tally.ScopeOptions{
			Tags:           map[string]string{"service": cfg.ServiceName},
			CachedReporter: reporter,
		},
		10*time.Second,
	)
	defer closer.Close()
	http.Handle("/metrics", reporter.HTTPHandler())
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Prometheus.MetricsPort), nil); err != nil {
			logger.Fatal("Failed to start the metrics handler", zap.Error(err))
		}
	}()

	counter := scope.Tagged(map[string]string{
		"service": cfg.ServiceName,
	}).Counter("service_started")
	counter.Inc(1)

	registry, err := consul.NewRegistry(cfg.Consul.URL)
	if err != nil {
		logger.Fatal("Failed to initialize registry with consul", zap.Error(err))
	}
	instanceID := discovery.GenerateInstanceID(cfg.ServiceName)
	if err := registry.Register(ctx, instanceID, cfg.ServiceName, fmt.Sprintf("%s:%d", cfg.ServiceName, port)); err != nil {
		logger.Fatal("Failed register gRPC instance in consul", zap.Error(err))
	}
	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, cfg.ServiceName); err != nil {
				logger.Error("Failed to report healthy state for gRPC", zap.Error(err))
			}
			time.Sleep(1 * time.Second)
		}
	}()
	serviceNameHTTP := cfg.ServiceName + "-http"
	instanceIDHTTP := discovery.GenerateInstanceID(serviceNameHTTP)
	if err := registry.Register(ctx, instanceIDHTTP, serviceNameHTTP, fmt.Sprintf("%s:%d", cfg.ServiceName, cfg.Prometheus.MetricsPort)); err != nil {
		logger.Fatal("Failed register HTTP instance in consul", zap.Error(err))
	}
	go func() {
		for {
			if err := registry.ReportHealthyState(instanceIDHTTP, serviceNameHTTP); err != nil {
				logger.Error("Failed to report healthy state for HTTP", zap.Error(err))
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, cfg.ServiceName)
	metadataGateway := metadataGateway.New(registry)
	ratingGateway := ratingGateway.New(registry)
	ctrl := movie.New(ratingGateway, metadataGateway)
	h := grpchandler.New(ctrl)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.ServiceName, port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}
	//const limit = 100
	//const burst = 100
	//l := newLimiter(limit, burst)
	srv := grpc.NewServer(
		//grpc.UnaryInterceptor(ratelimit.UnaryServerInterceptor(l)),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	reflection.Register(srv)
	gen.RegisterMovieServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		logger.Fatal("Failed to serve", zap.Error(err))
	}
}

// type limiter struct {
// 	l *rate.Limiter
// }

// func newLimiter(limit int, burst int) *limiter {
// 	return &limiter{rate.NewLimiter(rate.Limit(limit), burst)}
// }

// func (l *limiter) Limit() bool {
// 	return l.l.Allow()
// }
