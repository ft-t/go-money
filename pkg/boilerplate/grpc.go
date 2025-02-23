package boilerplate

import (
	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"context"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"
)

func GetDefaultGrpcServerBuilder() DefaultGrpcServerBuilder {
	return NewDefaultGrpcServerBuild(http.NewServeMux())
}

type DefaultGrpcServer struct {
	b                  *DefaultGrpcServerBuilder
	reflectionServices []string
	srv                *http.Server
}

type DefaultGrpcServerBuilder struct {
	httpMux      *http.ServeMux
	middlewares  []connect.UnaryInterceptorFunc
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func NewDefaultGrpcServerBuild(mux *http.ServeMux) DefaultGrpcServerBuilder {
	return DefaultGrpcServerBuilder{
		httpMux:      mux,
		readTimeout:  30 * time.Second,
		writeTimeout: 30 * time.Second,
	}
}

func (d DefaultGrpcServerBuilder) AddServerMiddleware(middleware connect.UnaryInterceptorFunc) DefaultGrpcServerBuilder {
	d.middlewares = append(d.middlewares, middleware)

	return d
}

func (d DefaultGrpcServerBuilder) WithReadTimeout(duration time.Duration) DefaultGrpcServerBuilder {
	d.readTimeout = duration

	return d
}

func (d DefaultGrpcServerBuilder) WithWriteTimeout(duration time.Duration) DefaultGrpcServerBuilder {
	d.writeTimeout = duration

	return d
}

func (d DefaultGrpcServerBuilder) Build() *DefaultGrpcServer {
	return &DefaultGrpcServer{
		b: &d,
	}
}

func (d *DefaultGrpcServer) GetMux() *http.ServeMux {
	return d.b.httpMux
}

func (d *DefaultGrpcServer) AddReflectionService(serviceName string) {
	d.reflectionServices = append(d.reflectionServices, serviceName)
}

func (d *DefaultGrpcServer) GetDefaultHandlerOptions() []connect.HandlerOption {
	var options []connect.HandlerOption

	if len(d.b.middlewares) > 0 {
		var converted []connect.Interceptor

		for _, r := range d.b.middlewares {
			converted = append(converted, r)
		}

		options = append(options,
			connect.WithInterceptors(converted...))
	}

	return options
}

func (d *DefaultGrpcServer) addProfiler() {
	mux := d.GetMux()
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.HandleFunc("/debug/pprof/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/debug/pprof/")
		if handler := pprof.Handler(name); handler != nil {
			handler.ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
}

func (d *DefaultGrpcServer) ServeAsync(grpcPort int) {
	grpcWebAddress := fmt.Sprintf("0.0.0.0:%d", grpcPort)

	if len(d.reflectionServices) > 0 {
		d.GetMux().Handle(grpcreflect.NewHandlerV1(
			grpcreflect.NewStaticReflector(d.reflectionServices...),
		))

		d.GetMux().Handle(grpcreflect.NewHandlerV1Alpha(
			grpcreflect.NewStaticReflector(d.reflectionServices...),
		))
	}

	d.addProfiler()

	go func() {
		d.srv = &http.Server{
			Addr: grpcWebAddress,
			Handler: h2c.NewHandler(
				newCORSMiddleware().Handler(d.GetMux()),
				&http2.Server{},
			),
			ReadHeaderTimeout: time.Second,
			ReadTimeout:       d.b.readTimeout,
			WriteTimeout:      d.b.writeTimeout,
			MaxHeaderBytes:    http.DefaultMaxHeaderBytes * 2, // 2 MB
		}

		log.Logger.Info().Msgf("server grpc listening at %v", grpcWebAddress)

		if err := d.srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				panic(err)
			}
		}
	}()
}

func (d *DefaultGrpcServer) Shutdown(ctx context.Context) error {
	return d.srv.Shutdown(ctx)
}

func newCORSMiddleware() *cors.Cors {
	return cors.New(cors.Options{
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowOriginFunc: func(origin string) bool {
			// Allow all origins, which effectively disables CORS.
			return true
		},
		AllowedHeaders: []string{"*"},
		ExposedHeaders: []string{
			// Content-Type is in the default safelist.
			"Accept",
			"Accept-Encoding",
			"Accept-Post",
			"Connect-Accept-Encoding",
			"Connect-Content-Encoding",
			"Content-Encoding",
			"Grpc-Accept-Encoding",
			"Grpc-Encoding",
			"Grpc-Message",
			"Grpc-Status",
			"Grpc-Status-Details-Bin",
		},
	})
}
