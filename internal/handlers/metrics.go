package handlers

import (
	"metricsink-go/internal/aggregator"
	"metricsink-go/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MetricsHandler struct {
	store *store.PostgresStore
	agg   *aggregator.Aggregator
}

func NewMetricsHandler(s *store.PostgresStore, agg *aggregator.Aggregator) *MetricsHandler {
	return &MetricsHandler{store: s, agg: agg}
}

func (h *MetricsHandler) IngestMetric(c *gin.Context) {
	var event store.MetricEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	event.ID = uuid.New().String()

	region := event.Tags["region"]

	h.agg.Add(event.Service, event.MetricName, event.Value)

	if h.store != nil {
		if err := h.store.InsertMetric(c.Request.Context(), event, region); err != nil {
			c.JSON(500, gin.H{"error": "failed to store metric"})
			return
		}
	}
	c.JSON(201, gin.H{"status": "ok", "id": event.ID})
}

func (h *MetricsHandler) GetServiceMetrics(c *gin.Context) {
	service := c.Param("service")
	metric := c.Query("metric")
	if metric == "" {
		metric = "latency"
	}
	avg := h.agg.GetAvg(service, metric)
	c.JSON(200, gin.H{
		"service": service,
		"metric":  metric,
		"avg":     avg,
	})
}
