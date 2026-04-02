package main

import (
	"log"
	"os"
	"time"

	"metricsink-go/internal/aggregator"
	"metricsink-go/internal/handlers"
	"metricsink-go/internal/store"

	"github.com/gin-gonic/gin"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	configURL := os.Getenv("CONFIG_SERVICE_URL")

	var s *store.PostgresStore
	if dsn != "" {
		var err error
		s, err = store.NewPostgresStore(dsn, configURL)
		if err != nil {
			log.Printf("Warning: DB unavailable: %v", err)
		}
	}

	agg := aggregator.NewAggregator(5 * time.Minute)
	h := handlers.NewMetricsHandler(s, agg)

	r := gin.Default()
	r.GET("/health", handlers.HealthHandler)
	r.POST("/metrics", h.IngestMetric)
	r.GET("/metrics/:service", h.GetServiceMetrics)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(r.Run(":" + port))
}
