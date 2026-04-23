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
  "dimensions": 1024,
  "instruction": "string",
  "output": {
    "primary": "dense",
    "additional": ["multi"]
  },
  "options": {}
}
```

### 4.1 Field Rules

- `model` (required): ai-proxy model id.
- `input` (required): ordered multimodal items.
- `input[].type` (required): one of `text`, `image`, `video`.
- `input[].text` required when type=`text`.
- `input[].image_url` required when type=`image`.
- `input[].video_url` required when type=`video`.
- `dimensions` (optional): requested output dimension. If model does not support dimension override, proxy may ignore and return model default.
- `instruction` (optional): task hint (maps to provider-specific `instruct`/`instructions`).
- `output` (optional): controls output vector families using explicit mode semantics.
  - `primary`: optional enum, `dense | fusion`, default `dense`.
  - `additional`: optional enum array, allowed values `multi | sparse`.
  - `additional` can be combined with `primary=dense`.
  - `primary=fusion` MUST NOT be combined with `additional` (reject with `400`).
  - unsupported `primary/additional` combinations for a given model MUST return `400` with clear `param` and message.
  - when `output` is omitted, ai-proxy should request provider default dense embedding only.
- `options` (optional): advanced key/value parameters. ai-proxy maps keys by model/provider capabilities (for example `fps`, `encoding_format`).

### 4.2 `output` Compatibility Matrix (Current)

| Output Request | Canonical Meaning | Typical Provider Support |
|---|---|---|
| `primary=dense` | return dense embedding | Qwen + Doubao |
| `primary=dense, additional=[multi]` | return dense + multi embedding | Doubao |
| `primary=dense, additional=[sparse]` | return dense + sparse embedding | Doubao (text-only constraint by model/provider) |
| `primary=dense, additional=[multi,sparse]` | return dense + multi + sparse | Doubao (text-only constraint for sparse) |
| `primary=fusion` | return fused multimodal embedding | Qwen models that support fusion |

### 4.3 `dimensions` Configurable Values (Documented)

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
  "model": "string",
  "created": 1776838589,
  "data": [
    {
      "index": 0,
      "type": "text",
      "embedding": [0.1, 0.2]
    },
    {
      "index": 1,
      "type": "image",
      "embedding": [0.1, 0.2]
    },
    {
      "index": 2,
      "type": "video",
      "embedding": [0.1, 0.2]
    }
  ],
  "usage": {
    "total_tokens": 54,
    "input_tokens": 54,
    "output_tokens": 0,
    "input_tokens_details": {
      "text_tokens": 54,
      "image_tokens": 0,
      "video_tokens": 0
    }
  },
  "request_id": "request-id"
}
```

### 5.1 Field Rules

- `data[]` is always an array in canonical response.
- `index` should be sequential (`0..n-1`) by output item order.
- `embedding` is the canonical dense vector output for each item.
- `type` reflects source modality (`text | image | video`).
- `output_tokens` is optional for embedding models; using `0` is acceptable when upstream does not provide it.

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
- `output.primary=fusion` -> `parameters.enable_fusion=true`
- `options.fps` -> `parameters.fps`

Qwen response -> Canonical:
- `output.embeddings[]` -> `data[]`
- `usage` passthrough with field normalization
- `request_id` -> `request_id`

### 7.2 Doubao (Volcengine Ark)

Canonical -> Doubao mapping:
- Path: `/api/v3/embeddings/multimodal`
- `instruction` -> `instructions`
- `input[]` -> `input[]` with provider item schema
- `dimensions` passthrough
- `output` omitted -> keep provider default dense output
- `output.additional` contains `multi` -> `multi_embedding.type=enabled`
- `output.additional` contains `sparse` -> `sparse_embedding.type=enabled`
- `options.encoding_format` -> `encoding_format`

Doubao response -> Canonical:
- `data.embedding` -> canonical `data[].embedding` (adapter expands/normalizes to array contract)
- top-level `usage/id/model/created` mapped to canonical fields (`request_id` etc.)

## 8. Compatibility Strategy

- Keep `/v1/embeddings` for OpenAI-compatible text embedding contract.
- Introduce `/v1/multimodal/embeddings` as the canonical multimodal contract.
- Do not silently switch `/v1/embeddings` to multimodal provider paths.

## 9. Validation Baseline in ai-proxy

Proxy MUST validate:
- required fields (`model`, `input`, per-item modality fields)
- supported modality type
- `output.primary/output.additional` enum constraints and combination constraints
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
