package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Counter is a wrapper around prometheus.Counter with auto-registration.
type Counter struct {
	counter prometheus.Counter
}

// CounterVec is a wrapper around prometheus.CounterVec with auto-registration.
type CounterVec struct {
	counterVec *prometheus.CounterVec
}

// NewCounter creates and registers a new Counter.
func NewCounter(opts prometheus.CounterOpts) *Counter {
	return &Counter{
		counter: promauto.NewCounter(opts),
	}
}

// NewCounterVec creates and registers a new CounterVec.
func NewCounterVec(opts prometheus.CounterOpts, labelNames []string) *CounterVec {
	return &CounterVec{
		counterVec: promauto.NewCounterVec(opts, labelNames),
	}
}

// Inc increments the counter by 1.
func (c *Counter) Inc() {
	c.counter.Inc()
}

// Add adds the given value to the counter.
func (c *Counter) Add(v float64) {
	c.counter.Add(v)
}

// WithLabelValues returns a counter with the given label values.
func (c *CounterVec) WithLabelValues(lvs ...string) prometheus.Counter {
	return c.counterVec.WithLabelValues(lvs...)
}

// With returns a counter with the given labels.
func (c *CounterVec) With(labels prometheus.Labels) prometheus.Counter {
	return c.counterVec.With(labels)
}

// Inc increments the counter with the given label values by 1.
func (c *CounterVec) Inc(lvs ...string) {
	c.counterVec.WithLabelValues(lvs...).Inc()
}

// Add adds the given value to the counter with the given label values.
func (c *CounterVec) Add(v float64, lvs ...string) {
	c.counterVec.WithLabelValues(lvs...).Add(v)
}