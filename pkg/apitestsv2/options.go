package apitestsv2

type option struct {
	tryV1RenderJsonBodyFirst bool
}

type OpOption func(*option)

// WithTryV1RenderJsonBodyFirst 尝试先使用 v1 严格模式渲染 json body。不论是否打开开关，都会再使用 v2 逻辑渲染一遍。
// 为手动测试的接口测试提供兼容处理；自动化测试无需打开该开关。
func WithTryV1RenderJsonBodyFirst() OpOption {
	return func(opt *option) {
		opt.tryV1RenderJsonBodyFirst = true
	}
}
