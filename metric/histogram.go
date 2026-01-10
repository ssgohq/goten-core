package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Histogram is a wrapper around prometheus.Histogram with auto-registration.
type Histogram struct {
	histogram prometheus.Histogram
}

// HistogramVec is a wrapper around prometheus.HistogramVec with auto-registration.
type HistogramVec struct {
	histogramVec *prometheus.HistogramVec
}

// NewHistogram creates and registers a new Histogram.
func NewHistogram(opts prometheus.HistogramOpts) *Histogram {
	return &Histogram{
		histogram: promauto.NewHistogram(opts),
	}
}

// NewHistogramVec creates and registers a new HistogramVec.
func NewHistogramVec(opts prometheus.HistogramOpts, labelNames []string) *HistogramVec {
	return &HistogramVec{
		histogramVec: promauto.NewHistogramVec(opts, labelNames),
	}
}

// Observe adds an observation to the histogram.
func (h *Histogram) Observe(v float64) {
	h.histogram.Observe(v)
}

// WithLabelValues returns an observer with the given label values.
func (h *HistogramVec) WithLabelValues(lvs ...string) prometheus.Observer {
	return h.histogramVec.WithLabelValues(lvs...)
}

// With returns an observer with the given labels.
func (h *HistogramVec) With(labels prometheus.Labels) prometheus.Observer {
	return h.histogramVec.With(labels)
}

// Observe adds an observation with the given label values.
func (h *HistogramVec) Observe(v float64, lvs ...string) {
	h.histogramVec.WithLabelValues(lvs...).Observe(v)
}

// DefaultBuckets is the default histogram buckets for latency metrics (in seconds).
var DefaultBuckets = []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

// DefaultSizeBuckets is the default histogram buckets for size metrics (in bytes).
var DefaultSizeBuckets = []float64{100, 1000, 10000, 100000, 1000000, 10000000}