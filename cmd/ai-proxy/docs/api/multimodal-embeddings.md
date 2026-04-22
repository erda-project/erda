# AI-Proxy API Design: `/v1/multimodal/embeddings`

## 1. Purpose

Define a unified, provider-agnostic API for multimodal embeddings (text/image/video), while keeping `/v1/embeddings` on OpenAI-compatible semantics for text-centric clients.

## 2. Scope

- This API is for multimodal embedding models (e.g. Qwen multimodal embedding, Doubao embedding vision).
- `/v1/embeddings` remains unchanged and OpenAI-compatible.
- This document defines the **canonical contract** between ai-proxy clients and ai-proxy.

## 3. Endpoint

- Method: `POST`
- Path: `/v1/multimodal/embeddings`

## 4. Request Schema (Canonical)

```json
{
  "model": "string",
  "input": [
    {
      "type": "text | image | video",
      "text": "string",
      "image_url": "string",
      "video_url": "string"
    }
  ],
  "encoding_format": "float | base64",
  "dimensions": 1024,
  "instruction": "string",
  "output": {
    "dense": true,
    "multi": false,
    "sparse": false,
    "fusion": false
  },
  "options": {
    "fps": 1.0,
    "provider": {}
  }
}
```

### 4.1 Field Rules

- `model` (required): ai-proxy model id.
- `input` (required): ordered multimodal items.
- `input[].type` (required): one of `text`, `image`, `video`.
- `input[].text` required when type=`text`.
- `input[].image_url` required when type=`image`.
- `input[].video_url` required when type=`video`.
- `encoding_format` (optional): default `float`.
- `dimensions` (optional): requested output dimension. If model does not support dimension override, proxy may ignore and return model default.
- `instruction` (optional): task hint (maps to provider-specific `instruct`/`instructions`).
- `output` (optional): controls output vector families.
  - `dense`: default `true`.
  - `multi`: default `false`.
  - `sparse`: default `false`.
  - `fusion`: default `false`.
- `options` (optional): advanced knobs, provider specific values go to `options.provider`.

### 4.2 `dimensions` Configurable Values (Documented)

`dimensions` is accepted in canonical request, but the allowed values are model-specific:

| Provider | Model | Supported dimensions | Notes |
|---|---|---|---|
| Alibaba Cloud | `qwen3-vl-embedding` | `2560, 2048, 1536, 1024, 768, 512, 256` | Default `2560` |
| Alibaba Cloud | `multimodal-embedding-v1` | fixed `1024` | `dimension` parameter is not supported |
| Alibaba Cloud | `tongyi-embedding-vision-plus` | fixed `1152` | `dimension` parameter is not supported in latest multimodal API doc |
| Alibaba Cloud | `tongyi-embedding-vision-flash` | fixed `768` | `dimension` parameter is not supported in latest multimodal API doc |
| Volcengine Ark | `doubao-embedding-vision-*` | `1024, 2048` | Default `2048` |

Validation policy in ai-proxy:
- If model has known allowlist, reject invalid values with `400`.
- For `doubao-embedding-vision-*`, enforce allowlist `{1024, 2048}` and default to `2048` when omitted.

## 5. Response Schema (Canonical)

```json
{
  "object": "list",
  "model": "string",
  "created": 1776838589,
  "data": [
    {
      "object": "embedding",
      "index": 0,
      "type": "text",
      "embedding": [0.1, 0.2],
      "multi_embedding": [[0.1, 0.2]],
      "sparse_embedding": [{"index": 12, "value": 0.34}]
    }
  ],
  "usage": {
    "prompt_tokens": 54,
    "total_tokens": 54,
    "input_tokens": 54,
    "image_tokens": 0,
    "video_tokens": 0,
    "input_tokens_details": {
      "text_tokens": 54,
      "image_tokens": 0
    }
  },
  "id": "request-id",
  "provider_response": {
    "request_id": "provider-request-id"
  }
}
```

### 5.1 Field Rules

- `data[]` is always an array in canonical response.
- `embedding` is the default dense vector.
- `multi_embedding` appears only when requested and provider supports it.
- `sparse_embedding` appears only when requested and provider supports it.
- `type` reflects source modality for independent outputs. If provider returns fused output, type may be `vl`.
- `provider_response` carries non-standard trace fields to avoid data loss.

## 6. Error Schema

```json
{
  "error": {
    "code": "invalid_request_error",
    "message": "input[1].video_url is required when type=video",
    "param": "input[1].video_url",
    "type": "validation_error"
  }
}
```

HTTP status:
- `400` invalid request
- `401` auth failed
- `403` forbidden
- `404` model or endpoint not found
- `429` rate limited
- `500/502/503` upstream/proxy internal errors

## 7. Provider Mapping

### 7.1 Qwen (Alibaba Cloud Model Studio)

Canonical -> Qwen mapping:
- `input` -> `input.contents`
- `instruction` -> `parameters.instruct`
- `dimensions` -> `parameters.dimension` (only for models that support it)
- `output.fusion` -> `parameters.enable_fusion`
- `options.fps` -> `parameters.fps`

Qwen response -> Canonical:
- `output.embeddings[]` -> `data[]`
- `usage` passthrough with field normalization
- `request_id` -> `provider_response.request_id`

### 7.2 Doubao (Volcengine Ark)

Canonical -> Doubao mapping:
- Path: `/api/v3/embeddings/multimodal`
- `instruction` -> `instructions`
- `input[]` -> `input[]` with provider item schema
- `dimensions` passthrough
- `output.multi=true` -> `multi_embedding.type=enabled`
- `output.sparse=true` -> `sparse_embedding.type=enabled`
- `encoding_format` passthrough

Doubao response -> Canonical:
- `data.embedding` -> `data[0].embedding`
- `data.multi_embedding` -> `data[0].multi_embedding`
- `data.sparse_embedding` -> `data[0].sparse_embedding`
- top-level `usage/id/model/created` passthrough

## 8. Compatibility Strategy

- Keep `/v1/embeddings` for OpenAI-compatible text embedding contract.
- Introduce `/v1/multimodal/embeddings` as the canonical multimodal contract.
- Do not silently switch `/v1/embeddings` to multimodal provider paths.

## 9. Validation Baseline in ai-proxy

Proxy MUST validate:
- required fields (`model`, `input`, per-item modality fields)
- supported modality type
- `output` booleans
- reject impossible combinations early when known (for example provider/model does not support sparse + image/video)

Proxy SHOULD:
- keep provider-specific strict validation in provider layer
- normalize response shape to canonical array-based `data`

## 10. Non-Goals (Current Design Stage)

- No streaming multimodal embedding API.
- No batch multimodal embedding API in this phase.
- No custom binary upload protocol in this endpoint.

## 11. References

- Alibaba Cloud multimodal embedding API reference: https://www.alibabacloud.com/help/zh/model-studio/multimodal-embedding-api-reference
- Volcengine Ark vectorization API (Doubao embedding vision): https://www.volcengine.com/docs/82379/1523520?lang=zh
- Alibaba Cloud (EN, last updated Mar 30, 2026, includes `dimension` constraints): https://www.alibabacloud.com/help/en/model-studio/multimodal-embedding-api-reference
