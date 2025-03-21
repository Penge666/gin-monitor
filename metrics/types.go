package metrics

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
	return nil
}
