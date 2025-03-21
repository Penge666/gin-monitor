package metrics

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

type Metric struct {
	//	监控指标的类型
	Type MetricType
	// 用于存储监控指标的名称
	Name string
	// 用于描述监控指标的帮助信息
	Help string
	// 用于存储监控指标的标签
	Labels []string
	//	用于存储监控指标的桶
	Buckets []float64
	// 用于存储监控指标的目标值
	Objectives map[float64]float64
	// 用于存储监控指标的向量
	vec prometheus.Collector
}

func counterHandler(metric *Metric) error {
	metric.vec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: metric.Name,
		Help: metric.Help,
	}, metric.Labels)
	return nil
}

func gaugeHandler(metric *Metric) error {
	metric.vec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: metric.Name,
		Help: metric.Help,
	}, metric.Labels)
	return nil
}
func histogramHandler(metric *Metric) error {
	metric.vec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    metric.Name,
		Help:    metric.Help,
		Buckets: metric.Buckets,
	}, metric.Labels)
	return nil
}
func summaryHandler(metric *Metric) error {
	metric.vec = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       metric.Name,
		Help:       metric.Help,
		Objectives: metric.Objectives,
	}, metric.Labels)
	return nil
}

func (m *Metric) Inc(labelValues []string) error {
	if m.Type == None {
		return fmt.Errorf("metric '%s' not existed.", m.Name)
	}

	if m.Type != GAUGE && m.Type != COUNTER {
		return fmt.Errorf("metric '%s' not Gauge or Counter type", m.Name)
	}
	switch m.Type {
	case COUNTER:
		m.vec.(*prometheus.CounterVec).WithLabelValues(labelValues...).Inc() // 等价于：m.vec.(*prometheus.CounterVec).WithLabelValues("GET", "200").Inc()
		break
	case GAUGE:
		m.vec.(*prometheus.GaugeVec).WithLabelValues(labelValues...).Inc()
		break
	}
	return nil
}
func (m *Metric) Add(labelValues []string, value float64) error {
	if m.Type == None {
		return fmt.Errorf("metric '%s' not existed.", m.Name)
	}

	if m.Type != GAUGE && m.Type != COUNTER {
		return fmt.Errorf("metric '%s' not Gauge or Counter type", m.Name)
	}
	switch m.Type {
	case COUNTER:
		m.vec.(*prometheus.CounterVec).WithLabelValues(labelValues...).Add(value)
		break
	case GAUGE:
		m.vec.(*prometheus.GaugeVec).WithLabelValues(labelValues...).Add(value)
		break
	}
	return nil
}
func (m *Metric) Observe(labelValues []string, value float64) error {
	if m.Type == None {
		return fmt.Errorf("metric '%s' not existed.", m.Name)
	}

	if m.Type != SUMMARY && m.Type != HISTOGRAM {
		return fmt.Errorf("metric '%s' not Summary or Histogram type", m.Name)
	}
	switch m.Type {
	case SUMMARY:
		/*记录一次 HTTP 请求耗时 1.2s
		m.Observe([]string{"GET", "200"}, 1.2)*/
		m.vec.(*prometheus.SummaryVec).WithLabelValues(labelValues...).Observe(value)
		break
	case HISTOGRAM:
		m.vec.(*prometheus.HistogramVec).WithLabelValues(labelValues...).Observe(value)
		break
	}
	return nil
}
