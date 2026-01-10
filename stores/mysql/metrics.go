package mysql

import (
	"context"
	"database/sql"
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "goten"
	subsystem = "mysql"
)

var (
	// Connection pool metrics
	openConnections = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "connections_open",
		Help:      "Number of open connections to the MySQL database",
	}, []string{"database"})

	inUseConnections = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "connections_in_use",
		Help:      "Number of connections currently in use",
	}, []string{"database"})

	idleConnections = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "connections_idle",
		Help:      "Number of idle connections",
	}, []string{"database"})

	maxOpenConnections = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "connections_max_open",
		Help:      "Maximum number of open connections configured",
	}, []string{"database"})

	waitCount = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "connections_wait_count_total",
		Help:      "Total number of connections waited for",
	}, []string{"database"})

	waitDuration = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "connections_wait_duration_seconds_total",
		Help:      "Total time blocked waiting for a connection",
	}, []string{"database"})

	maxIdleClosed = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "connections_max_idle_closed_total",
		Help:      "Total connections closed due to max idle connections limit",
	}, []string{"database"})

	maxLifetimeClosed = prom.NewGaugeVec(prom.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "connections_max_lifetime_closed_total",
		Help:      "Total connections closed due to max lifetime limit",
	}, []string{"database"})
)

func init() {
	prom.MustRegister(
		openConnections,
		inUseConnections,
		idleConnections,
		maxOpenConnections,
		waitCount,
		waitDuration,
		maxIdleClosed,
		maxLifetimeClosed,
	)
}

// MetricsCollector collects MySQL connection pool metrics.
type MetricsCollector struct {
	db       *sql.DB
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

// NewMetricsCollector creates a new MySQL metrics collector.
// Call Start() to begin collecting metrics, and Stop() to stop.
//
// Example:
//
//	db, _ := mysql.New(cfg)
//	collector := mysql.NewMetricsCollector(db, &mysql.MetricsConfig{
//	    DBName: "main",
//	})
//	collector.Start()
//	defer collector.Stop()
func NewMetricsCollector(db *sql.DB, cfg *MetricsConfig) *MetricsCollector {
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
		db:       db,
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
	stats := c.db.Stats()

	openConnections.WithLabelValues(c.dbName).Set(float64(stats.OpenConnections))
	inUseConnections.WithLabelValues(c.dbName).Set(float64(stats.InUse))
	idleConnections.WithLabelValues(c.dbName).Set(float64(stats.Idle))
	maxOpenConnections.WithLabelValues(c.dbName).Set(float64(stats.MaxOpenConnections))
	waitCount.WithLabelValues(c.dbName).Set(float64(stats.WaitCount))
	waitDuration.WithLabelValues(c.dbName).Set(stats.WaitDuration.Seconds())
	maxIdleClosed.WithLabelValues(c.dbName).Set(float64(stats.MaxIdleClosed))
	maxLifetimeClosed.WithLabelValues(c.dbName).Set(float64(stats.MaxLifetimeClosed))
}