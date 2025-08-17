package main

import (
	"context"
	"fmt"
	"itmo-ratings/internal/domain/rating/scrapper"
	rating "itmo-ratings/internal/domain/rating/student_rating_service"
	"itmo-ratings/internal/rpc/rating_summary"
	"itmo-ratings/pkg/info_handler"
	"itmo-ratings/pkg/middleware"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx := context.Background()
	host := "0.0.0.0"
	port := "8080"

	if len(os.Args) > 1 {
		host = os.Args[1]
	}
	if len(os.Args) > 2 {
		port = os.Args[2]
	}

	parser := scrapper.New(http.DefaultClient)

	ratingService := rating.New(parser)

	go func() {
		t := time.NewTicker(5 * time.Minute)
		for {
			select {
			case <-t.C:
				if err := ratingService.Enrich(ctx); err != nil {
					slog.Error("failed to update cache", "err", err.Error())
				}
			}
		}
	}()
	info := info_handler.New()
	mux := http.NewServeMux()

	mux.HandleFunc("/_info", info.ServeHTTP)
	mux.HandleFunc("/api/v1/rating/summary/{id}", rating_summary.New(ratingService).ServeHTTP)
	addr := fmt.Sprintf("%s:%s", host, port)

	logger := middleware.NewLogger(mux)
	rateLimiter := middleware.NewRateLimiter(logger, 10, 20)

	slog.Info("starting http server", "addr", addr)
	if err := http.ListenAndServe(addr, rateLimiter); err != nil {
		slog.Error("failed to start http server", "err", err.Error(), "addr", addr)
	}
}
