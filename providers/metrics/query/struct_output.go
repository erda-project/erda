package query

// 单点数据，对应API: {{scope}}?...
type Point struct {
	// Title string
	Name string
	Data []*PointData
}

type PointData struct {
	Name      string      `mapstructure:"name"`
	AggMethod string      `mapstructure:"agg"`
	Data      interface{} `mapstructure:"data"`
}

// 时序数据, 对应API：{{scope}}/histogram?...
type Series struct {
	Name       string
	Data       []*SeriesData
	TimeSeries []int // 毫秒
}

type SeriesData struct {
	Name      string    `mapstructure:"name"`
	AggMethod string    `mapstructure:"agg"`
	Data      []float64 `mapstructure:"data"`
	Tag       string    `mapstructure:"tag"`
}
