package boilerplate

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

type OpsHttpServer struct {
	mux     *http.ServeMux
	ready   bool
	healthy bool
	srv     *http.Server
}

func NewOpsHttpServer() *OpsHttpServer {
	h := &OpsHttpServer{
		mux:     http.NewServeMux(),
		ready:   false,
		healthy: true,
	}

	h.registerHttpHealthCheck()
	h.registerHttpReadinessCheck()
	h.registerMetrics()

	return h
}

func (r *OpsHttpServer) StartAsync(port int) *OpsHttpServer {
	if r.srv != nil {
		return r
	}

	r.srv = &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%d", port),
		Handler:           r.mux,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       90 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info().Msgf("HTTP server started on port [%d]", port)
		if err := r.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	return r
}

func (r *OpsHttpServer) registerMetrics() {
	r.mux.Handle("/metrics", promhttp.Handler())
}

func (r *OpsHttpServer) Stop() {
	r.healthy = false
	r.ready = false

	if r.srv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = r.srv.Shutdown(ctx)
	}
}

func (r *OpsHttpServer) registerHttpHealthCheck() {
	r.mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		if r.healthy {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	})
}

func (r *OpsHttpServer) registerHttpReadinessCheck() {
	r.mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
		if r.ready {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	})
}

func (r *OpsHttpServer) Ready() {
	r.ready = true
}
