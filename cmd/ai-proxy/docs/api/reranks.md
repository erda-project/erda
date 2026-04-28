# AI-Proxy API Design: `/v1/reranks`

## 1. Purpose

提供统一的重排（Rerank）接口，屏蔽不同服务商的请求结构与鉴权差异。

当前已接入：
- Qwen Rerank（Aliyun Bailian）
- Doubao Seed Rerank（Volcengine Viking）

## 2. Endpoint

- Method: `POST`
- Path: `/v1/reranks`

## 3. Authentication

调用 ai-proxy 本身使用：

```http
Authorization: Bearer <client-access-key-or-token>
```

## 4. Request Schema (Client -> ai-proxy)

```json
{
  "model": "bytedance/doubao-seed-rerank",
  "query": "十字花科是什么",
  "documents": [
    "十字花科植物广泛分布，包含蔬菜油料作物。",
    "机器学习是人工智能的一个分支。"
  ],
  "top_n": 2,
  "rerank_instruction": "Whether the document answers the query"
}
```

### 4.1 Field Rules

- `model` (required): ai-proxy 模型 ID。
- `query` (required for rerank): 查询内容，支持 `string` 或 `object`。
- `documents` (required): 候选文档数组，最多 200 条。
  - 支持 `string`（纯文本）
  - 支持 `object`（字段可包含 `query/content/text/document/image/title`）
- `top_n` (optional): 返回 top-N。
- `rerank_instruction` (optional): 重排指令（当前透传给支持该字段的 provider）。

### 4.2 Validation Baseline

ai-proxy 在请求阶段会做以下校验：

- `documents` 不能为空。
- `documents` 长度不能超过 `200`。
- 任意 `documents[i]` 解析失败会立即返回 `400`（不再静默丢弃）。
- 每条文档必须可确定 query（文档内 `query` 或回落全局 `query`）。
- 至少有一条文档包含 `content` 或 `image`。

> 说明：`documents[i]` 无效时直接报错是为了保证返回结果中的 `index` 与调用方原始输入严格对齐，避免隐式数据丢失导致索引错位。

## 5. Response Schema (ai-proxy -> Client)

```json
{
  "output": {
    "results": [
      {
        "index": 0,
        "relevance_score": 0.5190544657827126
      },
      {
        "index": 1,
        "relevance_score": 0.10216368026616135
      }
    ]
  },
  "usage": {
    "total_tokens": 459
  }
}
```

### 5.1 Field Notes

- `output.results[].index`: 对应原始 `documents` 的下标。
- `output.results[].relevance_score`: 相关性得分。
- `usage.total_tokens`: 上游返回的 token 消耗（可选）。

## 6. Error Examples

### 6.1 Invalid document

```json
{
  "message": "request: volcengine-viking-rerank-converter: documents[0] is invalid"
}
```

### 6.2 Missing query

```json
{
  "message": "request: volcengine-viking-rerank-converter: documents[0].query is required for rerank"
}
```

### 6.3 Provider backend error

```json
{
  "message": "LLM Backend Error",
  "error": {
    "raw_llm_backend_response": {
      "code": 1000001,
      "message": "check sign error, please check your ak, sk and tenant id"
    },
    "raw_llm_backend_status": "403 (Forbidden)",
    "type": "llm-backend-error"
  }
}
```

## 7. Provider Mapping

## 7.1 Aliyun Bailian

- `/v1/reranks` -> `/compatible-api/v1/reranks`
- 鉴权：Bearer API Key（provider 配置）

## 7.2 Volcengine Viking

- `/v1/reranks` -> `/api/knowledge/service/rerank`
- 当前支持两种鉴权路径：
  - Bearer token（如果 provider 只配 `apiKey`）
  - AK/SK HMAC-SHA256 签名（推荐；若配置了 `access_key_id/secret_access_key` 会自动签名）

Viking 侧最终请求结构（由 ai-proxy 转换后）：

```json
{
  "rerank_model": "doubao-seed-rerank",
  "datas": [
    {
      "query": "展示一张系统架构图",
      "title": "系统架构设计文档",
      "image": "https://example.com/arch.png"
    },
    {
      "query": "展示一张系统架构图",
      "title": "会议纪要",
      "content": "讨论了下个季度的 OKR。"
    }
  ],
  "top_k": 2,
  "rerank_instruction": "Whether the document answers the query"
}
```

## 8. Example cURL

```bash
curl -i "https://ai-proxy.erda.cloud/v1/reranks" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-client-ak-or-token>" \
  --data '{
    "model":"bytedance/doubao-seed-rerank",
    "query":"十字花科是什么",
    "documents":[
      "十字花科植物广泛分布，包含蔬菜油料作物。",
      "机器学习是人工智能的一个分支。"
    ],
    "top_n":2
  }'
```

## 9. Runtime Checklist

上线后可用前需确认：

1. 已部署包含 `/v1/reranks` 路由与相关 filter 的 ai-proxy 版本。
2. 已创建并启用 provider：
   - `aliyun-bailian`（用于 qwen rerank）
   - `volcengine-viking`（用于 doubao rerank）
3. 已创建并启用模型：
   - `qwen/qwen3-rerank`
   - `bytedance/doubao-seed-rerank`
4. 已把上述模型分配给目标 client。
5. `GET /v1/models` 可见对应 rerank 模型。
