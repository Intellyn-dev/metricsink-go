package store

import "time"

type MetricEvent struct {
	ID         string            `json:"id"`
	Service    string            `json:"service" binding:"required"`
	MetricName string            `json:"metric_name" binding:"required"`
	Value      float64           `json:"value" binding:"required"`
	Tags       map[string]string `json:"tags"`
	Timestamp  time.Time         `json:"timestamp"`
}

type Aggregate struct {
	Service     string    `json:"service"`
	MetricName  string    `json:"metric_name"`
	Avg         float64   `json:"avg"`
	Min         float64   `json:"min"`
	Max         float64   `json:"max"`
	Count       int64     `json:"count"`
	WindowStart time.Time `json:"window_start"`
}

type ServiceConfig struct {
	Name   string `json:"name"`
	Region string `json:"region"`
	Tier   string `json:"tier"`
}
