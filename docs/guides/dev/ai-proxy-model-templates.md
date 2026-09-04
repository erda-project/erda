# ai-proxy 模型模板

ai-proxy 在构建时通过 `go:embed` 将 `cmd/ai-proxy/conf/templates` 嵌入二进制。新增或修改模型模板后，需要重新构建 ai-proxy 镜像并滚动更新线上实例，模板才会出现在运行时模板列表中。

本轮新增的百炼模型模板：

- `qwen3.8-flash`
- `qwen3.8-max`
- `qwen3.8-27b`
- `kimi-k3`
- `deepseek-v4-pro-0813`
- `glm-5.2`

六个模板默认推荐服务商均为 `aliyun-bailian`，实际模型名默认等于模板 ID。线上仍需手动创建百炼服务商实例、模型实例及授权关系。

模板校验命令：

```bash
go run ./internal/apps/ai-proxy/common/template/check -path cmd/ai-proxy/conf/templates
```
