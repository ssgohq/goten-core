package redis

import (
	"context"
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

const (
	namespace = "goten"
	subsystem = "redis"
)

var (
	// Connection pool metrics
	hits = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pool_hits_total",
		Help:      "Number of times free connection was found in the pool",
	}, []string{"instance"})

	misses = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pool_misses_total",
		Help:      "Number of times free connection was NOT found in the pool",
	}, []string{"instance"})

	timeouts = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pool_timeouts_total",
		Help:      "Number of times a wait timeout occurred",
	}, []string{"instance"})

	totalConns = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pool_connections_total",
		Help:      "Number of total connections in the pool",
	}, []string{"instance"})

	idleConns = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pool_connections_idle",
		Help:      "Number of idle connections in the pool",
	}, []string{"instance"})

	staleConns = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pool_connections_stale",
		Help:      "Number of stale connections removed from the pool",
	}, []string{"instance"})
)

func init() {
	prom.MustRegister(
		hits,
		misses,
		timeouts,
		totalConns,
		idleConns,
		staleConns,
	)
}

// MetricsCollector collects Redis connection pool metrics.
type MetricsCollector struct {
	client       *redis.Client
	instanceName string
	interval     time.Duration
	cancel       context.CancelFunc
}

// MetricsConfig configures the metrics collector.
type MetricsConfig struct {
	// InstanceName is a label used to identify this Redis instance in metrics.
	// If empty, defaults to "default".
	InstanceName string

	// CollectInterval is the interval between stats collection.
	// Default is 15 seconds.
	CollectInterval time.Duration
}

// NewMetricsCollector creates a new Redis metrics collector.
// Call Start() to begin collecting metrics, and Stop() to stop.
//
// Example:
//
//	client := redis.New(cfg)
//	collector := redis.NewMetricsCollector(client, &redis.MetricsConfig{
//	    InstanceName: "cache",
//	})
//	collector.Start()
//	defer collector.Stop()
func NewMetricsCollector(client *redis.Client, cfg *MetricsConfig) *MetricsCollector {
	if cfg == nil {
		cfg = &MetricsConfig{}
	}

	instanceName := cfg.InstanceName
	if instanceName == "" {
		instanceName = "default"
	}

	interval := cfg.CollectInterval
	if interval == 0 {
		interval = 15 * time.Second
	}

	return &MetricsCollector{
		client:       client,
		instanceName: instanceName,
		interval:     interval,
	}
}

// Start begins collecting metrics at the configured interval.
func (c *MetricsCollector) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	// Collect initial stats
	c.collect()

	// Start background collection
	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.collect()
			}
		}
	}()
}

// Stop stops the metrics collection.
func (c *MetricsCollector) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *MetricsCollector) collect() {
	stats := c.client.PoolStats()

	hits.WithLabelValues(c.instanceName).Set(float64(stats.Hits))
	misses.WithLabelValues(c.instanceName).Set(float64(stats.Misses))
	timeouts.WithLabelValues(c.instanceName).Set(float64(stats.Timeouts))
	totalConns.WithLabelValues(c.instanceName).Set(float64(stats.TotalConns))
	idleConns.WithLabelValues(c.instanceName).Set(float64(stats.IdleConns))
	staleConns.WithLabelValues(c.instanceName).Set(float64(stats.StaleConns))
}