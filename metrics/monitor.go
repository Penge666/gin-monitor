package metrics

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type MetricType int

const (
	None MetricType = iota
	// 计数器类型
	COUNTER
	// 仪表盘类型
	GAUGE
	// 直方图类型
	HISTOGRAM
	// 摘要类型
	SUMMARY
)
const (
	defaultMetricPath = "/metrics"
	defaultSlowTime   = int32(5)
)

var (
	defaultExcludePaths = []string{}
	defaultDuration     = []float64{0.1, 0.3, 1.2, 5, 10}
)

var (
	monitor *Monitor
	once    sync.Once
)
var (
	promTypeHandler = map[MetricType]func(metric *Metric) error{
		COUNTER:   counterHandler,
		GAUGE:     gaugeHandler,
		HISTOGRAM: histogramHandler,
		SUMMARY:   summaryHandler,
	}
)
var (
	metricRequestTotal    = "gin_request_total"
	metricRequestUVTotal  = "gin_request_uv_total"
	metricURIRequestTotal = "gin_uri_request_total"
	metricRequestBody     = "gin_request_body_total"
	metricResponseBody    = "gin_response_body_total"
	metricRequestDuration = "gin_request_duration"
	metricSlowRequest     = "gin_slow_request_total"

	bloomFilter *BloomFilter
)

type Monitor struct {
	// 慢查询时间
	slowTime int32
	// 监控路径
	metricPath string // 默认通常是 `/metrics`
	//  不需要监控的路径列表
	excludePath []string
	// 请求持续时间的统计阈值,用于计算不同百分位（如 p95、p99）的请求延时数据
	reqDuration []float64
	// 存储所有的监控指标
	metrics map[string]*Metric
	// 存储元数据
	metadata map[string]string
}

func GetNewMonitor() *Monitor {
	// 初始化一个监控器
	once.Do(func() {
		monitor = &Monitor{
			slowTime:    defaultSlowTime,
			metricPath:  defaultMetricPath,
			excludePath: defaultExcludePaths,
			reqDuration: defaultDuration,
			metrics:     make(map[string]*Metric),
			metadata:    make(map[string]string),
		}
	})
	return monitor
}

func (m *Monitor) SetSlowTime(slowTime int32) {
	m.slowTime = slowTime
}
func (m *Monitor) SetMetricPath(metricPath string) {
	m.metricPath = metricPath
}
func (m *Monitor) SetExcludePath(excludePath []string) {
	m.excludePath = excludePath
}
func (m *Monitor) SetReqDuration(reqDuration []float64) {
	m.reqDuration = reqDuration
}
func (m *Monitor) SetMetadata(data interface{}) {
	switch v := data.(type) {
	case map[string]string:
		for k, val := range v {
			m.metadata[k] = val
		}
	case struct{ key, value string }:
		m.metadata[v.key] = v.value
	default:
		fmt.Println("unsupported type")
	}
}
func (m *Monitor) GetMetrics(name string) (*Metric, error) {
	metric, ok := m.metrics[name]
	if !ok {
		return &Metric{}, fmt.Errorf("metric %s not found", name)
	}
	return metric, nil
}

func (m *Monitor) Use(r gin.IRoutes) {
	// 开始拦截
	m.initGinMetrics()
	r.Use(m.monitorInterceptor)
	r.GET(m.metricPath, func(ctx *gin.Context) {
		promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request)
	})
}

func (m *Monitor) initGinMetrics() {
	bloomFilter = NewBloomFilter()
	// 添加我要用的指标，然后注册到prometheus上面进行监测
	err := m.AddMetric(&Metric{
		Type:   COUNTER,
		Name:   metricRequestTotal,
		Help:   "Total number of HTTP requests made.",
		Labels: m.getMetricLabelsIncludingMetadata(metricRequestTotal),
	})
	if err != nil {
		fmt.Printf("AddMetric failed: %v\n", err)
	}
	err = m.AddMetric(&Metric{
		Type:   COUNTER,
		Name:   metricRequestUVTotal,
		Help:   "Total number of unique visitors.",
		Labels: m.getMetricLabelsIncludingMetadata(metricRequestUVTotal),
	})
	if err != nil {
		fmt.Printf("AddMetric failed: %v\n", err)
	}
	err = m.AddMetric(&Metric{
		Type:   COUNTER,
		Name:   metricURIRequestTotal,
		Help:   "Total number of HTTP requests made.",
		Labels: m.getMetricLabelsIncludingMetadata(metricURIRequestTotal),
	})
	if err != nil {
		fmt.Printf("AddMetric failed: %v\n", err)
	}
	err = m.AddMetric(&Metric{
		Type:   COUNTER,
		Name:   metricRequestBody,
		Help:   "Total number of HTTP requests made.",
		Labels: m.getMetricLabelsIncludingMetadata(metricRequestBody),
	})
	if err != nil {
		fmt.Printf("AddMetric failed: %v\n", err)
	}
	err = m.AddMetric(&Metric{
		Type:   COUNTER,
		Name:   metricResponseBody,
		Help:   "Total number of HTTP requests made.",
		Labels: m.getMetricLabelsIncludingMetadata(metricResponseBody),
	})
	if err != nil {
		fmt.Printf("AddMetric failed: %v\n", err)
	}
	err = m.AddMetric(&Metric{
		Type:   HISTOGRAM,
		Name:   metricRequestDuration,
		Help:   "The HTTP request latencies in seconds.",
		Labels: m.getMetricLabelsIncludingMetadata(metricRequestDuration),
	})
	if err != nil {
		fmt.Printf("AddMetric failed: %v\n", err)
	}
	err = m.AddMetric(&Metric{
		Type:   COUNTER,
		Name:   metricSlowRequest,
		Help:   "Total number of slow HTTP requests made.",
		Labels: m.getMetricLabelsIncludingMetadata(metricSlowRequest),
	})
	if err != nil {
		fmt.Printf("AddMetric failed: %v\n", err)
	}
}

func (m *Monitor) AddMetric(metric *Metric) error {
	if _, ok := m.metrics[metric.Name]; ok {
		return fmt.Errorf("metric %s already exists", metric.Name)
	}
	if metric.Name == "" {
		return fmt.Errorf("metric name cannot empty")
	}
	// 将指标注册到prometheus中
	// 1.合成指标
	if err := promTypeHandler[metric.Type](metric); err != nil {
		return fmt.Errorf("register metric %s failed: %v", metric.Name, err)
	}
	// 2.注册指标
	prometheus.MustRegister(metric.vec)
	m.metrics[metric.Name] = metric
	return nil
}

// getMetricLabelsIncludingMetadata 返回指定监控指标的标签列表，并根据配置是否包含元数据
func (m *Monitor) getMetricLabelsIncludingMetadata(metricName string) []string {
	// 判断是否需要包含元数据
	includes_metadata := m.includesMetadata()

	// 获取元数据标签（如果有）
	metadata_labels, _ := m.getMetadata()

	// 根据不同的监控指标返回相应的标签列表
	switch metricName {
	case metricRequestDuration:
		// 请求持续时间指标，仅包含 "uri" 作为默认标签
		metric_labels := []string{"uri"}
		// 如果需要包含元数据，则追加到标签列表
		if includes_metadata {
			metric_labels = append(metric_labels, metadata_labels...)
		}
		return metric_labels

	case metricURIRequestTotal:
		// 请求总数指标，包含 "uri", "method", "code" 作为默认标签
		metric_labels := []string{"uri", "method", "code"}
		// 如果需要包含元数据，则追加到标签列表
		if includes_metadata {
			metric_labels = append(metric_labels, metadata_labels...)
		}
		return metric_labels

	case metricSlowRequest:
		// 慢请求指标，包含 "uri", "method", "code" 作为默认标签
		metric_labels := []string{"uri", "method", "code"}
		// 如果需要包含元数据，则追加到标签列表
		if includes_metadata {
			metric_labels = append(metric_labels, metadata_labels...)
		}
		return metric_labels

	default:
		// 对于未知的指标，默认不包含标签
		var metric_labels []string = nil
		// 但如果需要包含元数据，则返回元数据标签
		if includes_metadata {
			metric_labels = metadata_labels
		}
		return metric_labels
	}
}

// 检查 Monitor 结构体的 metadata 是否包含内容。
func (m *Monitor) includesMetadata() bool {
	return len(m.metadata) > 0
}

// getMetadata 返回元数据的标签（keys）和对应的值（values）。
//
// 示例：
// 假设 Monitor 结构体的 metadata 为：
//
//	metadata: map[string]string{
//	    "host": "server1",
//	    "env":  "production",
//	    "app":  "my_service",
//	}
//
// 调用 getMetadata() 将返回：
//
//	Labels: ["host", "env", "app"]
//	Values: ["server1", "production", "my_service"]
func (m *Monitor) getMetadata() ([]string, []string) {
	metadata_labels := []string{}
	metadata_values := []string{}

	for v := range m.metadata {
		metadata_labels = append(metadata_labels, v)
		metadata_values = append(metadata_values, m.metadata[v])
	}
	return metadata_labels, metadata_values
}

func (m *Monitor) monitorInterceptor(ctx *gin.Context) {
	if m.isExcludePath(ctx.Request.URL.Path) {
		ctx.Next()
		return
	}
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			// 这里可以记录日志或者做异常监控
			fmt.Println("Recovered from panic:", r)
			// 如果发生 panic，直接返回 500 错误
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		// 即使发生 panic，仍然执行监控处理
		m.ginMetricHandle(ctx, start)
	}()

	ctx.Next()
}

func (m *Monitor) isExcludePath(path string) bool {
	for _, p := range m.excludePath {
		if p == path {
			return true
		}
	}
	return false
}

func (m *Monitor) ginMetricHandle(ctx *gin.Context, start time.Time) {
	// 如果请求已经被中止（例如由于 panic），则不再处理指标
	if ctx.IsAborted() {
		return
	}

	r := ctx.Request
	w := ctx.Writer

	// 1.请求数量
	metric, err := m.GetMetrics(metricRequestTotal)
	if err != nil {
		fmt.Printf("GetMetrics failed: %v\n", err)
	} else {
		metric.Inc(m.getMetricValues(nil))
	}

	// 2.请求uv
	metric, err = m.GetMetrics(metricRequestUVTotal)
	if err != nil {
		fmt.Printf("GetMetrics failed: %v\n", err)
	} else if clientIP := ctx.ClientIP(); !bloomFilter.Contains(clientIP) {
		bloomFilter.Add(clientIP)
		_ = metric.Inc(m.getMetricValues(nil))
	}

	// 3.请求uri数量
	metric, err = m.GetMetrics(metricURIRequestTotal)
	if err != nil {
		fmt.Printf("GetMetrics failed: %v\n", err)
	} else {
		err = metric.Inc(m.getMetricValues([]string{ctx.FullPath(), r.Method, strconv.Itoa(w.Status())}))
		if err != nil {
			fmt.Printf("Inc failed: %v\n", err)
		}
	}

	// 4.请求body大小
	if r.ContentLength > 0 {
		metric, err = m.GetMetrics(metricRequestBody)
		if err != nil {
			fmt.Printf("GetMetrics failed: %v\n", err)
		} else {
			err = metric.Add(m.getMetricValues([]string{ctx.FullPath(), r.Method, strconv.Itoa(w.Status())}), float64(r.ContentLength))
			if err != nil {
				fmt.Printf("Add failed: %v\n", err)
			}
		}
	}

	// 5.响应body大小
	if w.Size() > 0 {
		metric, err = m.GetMetrics(metricResponseBody)
		if err != nil {
			fmt.Printf("GetMetrics failed: %v\n", err)
		} else {
			err = metric.Add(m.getMetricValues(nil), float64(w.Size()))
			if err != nil {
				fmt.Printf("Add failed: %v\n", err)
			}
		}
	}
	elapsed := time.Since(start).Seconds()
	// 6.请求持续时间
	metric, err = m.GetMetrics(metricRequestDuration)
	if err != nil {
		fmt.Printf("GetMetrics failed: %v\n", err)
	} else {

		err = metric.Observe(m.getMetricValues([]string{ctx.FullPath()}), elapsed)
		if err != nil {
			fmt.Printf("Observe failed: %v\n", err)
		}
	}

	// 7.慢请求
	if elapsed > float64(m.slowTime) {
		metric, err = m.GetMetrics(metricSlowRequest)
		if err != nil {
			fmt.Printf("GetMetrics failed: %v\n", err)
		} else {
			err = metric.Inc(m.getMetricValues([]string{ctx.FullPath()}))
			if err != nil {
				fmt.Printf("Inc failed: %v\n", err)
			}
		}
	}
}

func (m *Monitor) getMetricValues(metric_values []string) []string {
	haveMetadata := m.includesMetadata()
	if haveMetadata {
		_, metadata_values := m.getMetadata()
		metric_values = append(metric_values, metadata_values...)
	}
	return metric_values
}
