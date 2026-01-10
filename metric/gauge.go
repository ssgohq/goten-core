package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Gauge is a wrapper around prometheus.Gauge with auto-registration.
type Gauge struct {
	gauge prometheus.Gauge
}

// GaugeVec is a wrapper around prometheus.GaugeVec with auto-registration.
type GaugeVec struct {
	gaugeVec *prometheus.GaugeVec
}

// NewGauge creates and registers a new Gauge.
func NewGauge(opts prometheus.GaugeOpts) *Gauge {
	return &Gauge{
		gauge: promauto.NewGauge(opts),
	}
}

// NewGaugeVec creates and registers a new GaugeVec.
func NewGaugeVec(opts prometheus.GaugeOpts, labelNames []string) *GaugeVec {
	return &GaugeVec{
		gaugeVec: promauto.NewGaugeVec(opts, labelNames),
	}
}

// Set sets the gauge to the given value.
func (g *Gauge) Set(v float64) {
	g.gauge.Set(v)
}

// Inc increments the gauge by 1.
func (g *Gauge) Inc() {
	g.gauge.Inc()
}

// Dec decrements the gauge by 1.
func (g *Gauge) Dec() {
	g.gauge.Dec()
}

// Add adds the given value to the gauge.
func (g *Gauge) Add(v float64) {
	g.gauge.Add(v)
}

// Sub subtracts the given value from the gauge.
func (g *Gauge) Sub(v float64) {
	g.gauge.Sub(v)
}

// WithLabelValues returns a gauge with the given label values.
func (g *GaugeVec) WithLabelValues(lvs ...string) prometheus.Gauge {
	return g.gaugeVec.WithLabelValues(lvs...)
}

// With returns a gauge with the given labels.
func (g *GaugeVec) With(labels prometheus.Labels) prometheus.Gauge {
	return g.gaugeVec.With(labels)
}

// Set sets the gauge with the given label values to the given value.
func (g *GaugeVec) Set(v float64, lvs ...string) {
	g.gaugeVec.WithLabelValues(lvs...).Set(v)
}

// Inc increments the gauge with the given label values by 1.
func (g *GaugeVec) Inc(lvs ...string) {
	g.gaugeVec.WithLabelValues(lvs...).Inc()
}

// Dec decrements the gauge with the given label values by 1.
func (g *GaugeVec) Dec(lvs ...string) {
	g.gaugeVec.WithLabelValues(lvs...).Dec()
}