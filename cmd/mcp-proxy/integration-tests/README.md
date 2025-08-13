# MCP Proxy 集成测试

这个目录包含了 MCP Proxy 服务的集成测试，用于验证代理接口的功能。

## 功能概述

集成测试覆盖了以下代理接口功能：

- **连接 (Connect)**: 测试 MCP Server 连接接口 (GET 和 POST)
- **消息 (Message)**: 测试代理消息接口
- **错误处理**: 测试各种错误情况和边界条件

## 快速开始

### 1. 环境设置

```bash
# 复制环境配置文件
cp env.example .env

# 编辑配置文件
vim .env
```

### 2. 运行测试

```bash
# 运行所有测试
make test

# 运行特定测试
make test-connect
make test-message
make test-error

# 查看帮助
make help
```

## 配置说明

### 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| `HOST` | MCP Proxy 服务地址 | `http://localhost:8081` |
| `TOKEN` | 认证令牌 | 空 |
| `TIMEOUT` | 请求超时时间（秒） | `120` |
| `TEST_MCP_NAME` | 测试 MCP 服务器名称 | `test-mcp-server` |
| `TEST_MCP_TAG` | 测试 MCP 服务器标签 | `v1.0.0` |

## 测试结构

```
integration-tests/
├── config/           # 配置管理
│   └── config.go
├── common/           # 通用工具
│   └── client.go
├── proxy_test.go     # 代理接口测试文件
├── Makefile          # 构建和测试脚本
├── go.mod           # Go 模块文件
├── env.example      # 环境变量示例
└── README.md        # 本文档
```

## API 端点

测试覆盖的代理接口端点：

- `GET /proxy/connect/{mcpName}/{mcpTag}` - MCP Server 连接 (GET)
- `POST /proxy/connect/{mcpName}/{mcpTag}` - MCP Server 连接 (POST)
- `POST /proxy/message` - 代理消息接口

## 测试数据

测试使用以下数据结构：

### ProxyConnectRequest
```go
type ProxyConnectRequest struct {
    MCPName string `json:"mcpName"`
    MCPTag  string `json:"mcpTag"`
}
```

### ProxyMessageRequest
```go
type ProxyMessageRequest struct {
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

## 故障排除

### 常见问题

1. **连接失败**: 确保 MCP Proxy 服务正在运行且地址配置正确
2. **认证失败**: 检查 TOKEN 配置是否正确
3. **超时错误**: 增加 TIMEOUT 值或检查网络连接
4. **测试失败**: 检查 MCP Server 是否已注册并可用

### 调试模式

```bash
# 运行详细输出
make test-verbose

# 运行竞态检测
make test-race

# 运行覆盖率测试
make test-coverage
```

## 开发指南

### 添加新测试

1. 在 `proxy_test.go` 中添加新的测试函数
2. 遵循命名约定 `TestProxy{Action}`
3. 在 Makefile 中添加对应的测试目标
4. 更新文档

### 修改配置

1. 更新 `config/config.go` 中的配置结构
2. 在 `setConfigValue` 函数中添加新的配置项处理
3. 更新 `env.example` 文件
4. 更新文档

## 许可证

本项目遵循 Apache License 2.0 许可证。

## 总结

这个集成测试套件为 `cmd/mcp-proxy` 提供了：

1. **代理接口测试** - 专注于测试 MCP Proxy 的核心代理功能
2. **连接测试** - 测试 MCP Server 的连接接口 (GET 和 POST)
3. **消息测试** - 测试代理消息传递功能
4. **错误处理** - 测试各种异常情况和边界条件
5. **易于使用** - 提供了清晰的文档和便捷的 Makefile 命令
6. **可扩展性** - 模块化设计便于添加新的测试用例

这个测试套件专注于验证 MCP Proxy 服务的代理功能，确保代理接口的稳定性和可靠性。 