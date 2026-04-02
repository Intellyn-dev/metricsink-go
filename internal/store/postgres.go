package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db        *sql.DB
	configURL string
}

func NewPostgresStore(dsn, configURL string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &PostgresStore{db: db, configURL: configURL}, nil
}

func (s *PostgresStore) InsertMetric(ctx context.Context, event MetricEvent, region string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO metrics (service, metric_name, value, region, tags, created_at)
         VALUES ($1, $2, $3, $4, $5, NOW())`,
		event.Service, event.MetricName, event.Value, region, event.Tags,
	)
	return err
}

func (s *PostgresStore) GetAggregates(ctx context.Context, service string) ([]Aggregate, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT service, metric_name, AVG(value), MIN(value), MAX(value), COUNT(*), MIN(created_at)
         FROM metrics WHERE service = $1 AND created_at > NOW() - INTERVAL '1 hour'
         GROUP BY service, metric_name`,
		service,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var aggs []Aggregate
	for rows.Next() {
		var a Aggregate
		if err := rows.Scan(&a.Service, &a.MetricName, &a.Avg, &a.Min, &a.Max, &a.Count, &a.WindowStart); err != nil {
			return nil, err
		}
		aggs = append(aggs, a)
	}
	return aggs, nil
}

func fetchServiceConfig(url string) (*ServiceConfig, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("config service returned %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	var cfg ServiceConfig
	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *PostgresStore) GetServiceConfig() (*ServiceConfig, error) {
	if s.configURL == "" {
		return &ServiceConfig{Name: "unknown", Region: "us-east-1", Tier: "standard"}, nil
	}
	return fetchServiceConfig(s.configURL)
}

var _ = time.Now
