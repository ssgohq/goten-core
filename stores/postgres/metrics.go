package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	prom "github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "goten"
	subsystem = "postgres"
)

var (
	// Connection pool metrics
	acquiredConns = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "connections_acquired",
		Help:      "Number of currently acquired connections",
	}, []string{"database"})

	idleConns = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "connections_idle",
		Help:      "Number of idle connections in the pool",
	}, []string{"database"})

	totalConns = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "connections_total",
		Help:      "Total number of connections in the pool",
	}, []string{"database"})

	maxConns = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "connections_max",
		Help:      "Maximum number of connections configured",
	}, []string{"database"})

	constructingConns = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "connections_constructing",
		Help:      "Number of connections being constructed",
	}, []string{"database"})

	acquireCount = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "connections_acquire_count_total",
		Help:      "Total number of successful connection acquires",
	}, []string{"database"})

	acquireDuration = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "connections_acquire_duration_seconds_total",
		Help:      "Total time spent acquiring connections",
	}, []string{"database"})

	canceledAcquireCount = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "connections_canceled_acquire_count_total",
		Help:      "Total number of acquire calls canceled by context",
	}, []string{"database"})

	emptyAcquireCount = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "connections_empty_acquire_count_total",
		Help:      "Total number of successful acquires from an empty pool",
	}, []string{"database"})

	newConnsCount = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "connections_new_count_total",
		Help:      "Total number of new connections opened",
	}, []string{"database"})

	maxLifetimeDestroyCount = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "connections_max_lifetime_destroy_count_total",
		Help:      "Total number of connections destroyed due to max lifetime",
	}, []string{"database"})

	maxIdleDestroyCount = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "connections_max_idle_destroy_count_total",
		Help:      "Total number of connections destroyed due to max idle time",
	}, []string{"database"})
)

func init() {
	prom.MustRegister(
		acquiredConns,
		idleConns,
		totalConns,
		maxConns,
		constructingConns,
		acquireCount,
		acquireDuration,
		canceledAcquireCount,
		emptyAcquireCount,
		newConnsCount,
		maxLifetimeDestroyCount,
		maxIdleDestroyCount,
	)
}

// MetricsCollector collects PostgreSQL connection pool metrics.
type MetricsCollector struct {
	pool     *pgxpool.Pool
	dbName   string
	interval time.Duration
	cancel   context.CancelFunc
}

// MetricsConfig configures the metrics collector.
type MetricsConfig struct {
	// DBName is a label used to identify this database in metrics.
	// If empty, defaults to "default".
	DBName string

	// CollectInterval is the interval between stats collection.
	// Default is 15 seconds.
	CollectInterval time.Duration
}

// NewMetricsCollector creates a new PostgreSQL metrics collector.
// Call Start() to begin collecting metrics, and Stop() to stop.
//
// Example:
//
//	pool, _ := postgres.New(ctx, cfg)
//	collector := postgres.NewMetricsCollector(pool, &postgres.MetricsConfig{
//	    DBName: "main",
//	})
//	collector.Start()
//	defer collector.Stop()
func NewMetricsCollector(pool *pgxpool.Pool, cfg *MetricsConfig) *MetricsCollector {
	if cfg == nil {
		cfg = &MetricsConfig{}
	}

	dbName := cfg.DBName
	if dbName == "" {
		dbName = "default"
	}

	interval := cfg.CollectInterval
	if interval == 0 {
		interval = 15 * time.Second
	}

	return &MetricsCollector{
		pool:     pool,
		dbName:   dbName,
		interval: interval,
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
	stat := c.pool.Stat()

	acquiredConns.WithLabelValues(c.dbName).Set(float64(stat.AcquiredConns()))
	idleConns.WithLabelValues(c.dbName).Set(float64(stat.IdleConns()))
	totalConns.WithLabelValues(c.dbName).Set(float64(stat.TotalConns()))
	maxConns.WithLabelValues(c.dbName).Set(float64(stat.MaxConns()))
	constructingConns.WithLabelValues(c.dbName).Set(float64(stat.ConstructingConns()))
	acquireCount.WithLabelValues(c.dbName).Set(float64(stat.AcquireCount()))
	acquireDuration.WithLabelValues(c.dbName).Set(stat.AcquireDuration().Seconds())
	canceledAcquireCount.WithLabelValues(c.dbName).Set(float64(stat.CanceledAcquireCount()))
	emptyAcquireCount.WithLabelValues(c.dbName).Set(float64(stat.EmptyAcquireCount()))
	newConnsCount.WithLabelValues(c.dbName).Set(float64(stat.NewConnsCount()))
	maxLifetimeDestroyCount.WithLabelValues(c.dbName).Set(float64(stat.MaxLifetimeDestroyCount()))
	maxIdleDestroyCount.WithLabelValues(c.dbName).Set(float64(stat.MaxIdleDestroyCount()))
}