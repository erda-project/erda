# AI-Proxy

AI-Proxy 是端点及 Erda 各应用接入 OpenAI 或 Azure 服务的网关。

# AI-Proxy 服务配置

AI-Proxy 服务本质是一个网关中间件，它将客户端的请求反向代理到后端服务。
`provider` 是 AI 能力提供商，如 OpenAI，Azure 等，也即 AI-Proxy 的

## providers

示例:

```yaml
providers:
  - name: openai
    instanceId: default
    host: api.openai.com
    scheme: https
    description: openai 提供的 chatgpt 能力
    docSite: https://platform.openai.com/docs/api-reference
    appKey: ${ env.OPENAI_API_KEY }
    organization: ""
    metadata: {}

  - name: azure
    instanceId: default
    host: codeai.openai.azure.com
    scheme: https
    description: azure 提供的 ai 能力
    docSite: https://learn.microsoft.com/en-us/azure/cognitive-services/openai/reference
    metadata:
      RESOURCE_NAME: "codeai"
      DEVELOPMENT_NAME: "gpt-35-turbo-0301"
```

## routes

