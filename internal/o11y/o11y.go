package o11y

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	totalEventCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "katulampa",
			Subsystem: "larkapp",
			Name:      "event_total",
			Help:      "Total number event",
		},
		[]string{"type", "success"},
	)
	eventDurration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "katulampa",
			Subsystem: "larkapp",
			Name:      "event_duration_seconds",
			Help:      "Duration processing events",
		},
		[]string{"type"},
	)
	postToLarkCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "katulampa",
			Subsystem: "larkapp",
			Name:      "post_to_lark_total",
			Help:      "Total post to lark",
		},
		[]string{"channel", "method", "status"},
	)
)

func IncreasePostToLarkCounter(channel, method string, success bool) {
	status := "failed"
	if success {
		status = "success"
	}
	postToLarkCounter.WithLabelValues(channel, method, status).Inc()
}

func ObserveEventHandler(eventType string, fn func() error) error {
	timer := prometheus.NewTimer(eventDurration.WithLabelValues(eventType))
	err := fn()
	timer.ObserveDuration()
	increaseEventCounter(eventType, err == nil)
	return err
}

func increaseEventCounter(eventName string, success bool) {
	totalEventCounter.WithLabelValues(eventName, strconv.FormatBool(success)).Inc()
	totalEventCounter.WithLabelValues(eventName, strconv.FormatBool(!success)).Add(0)
}
