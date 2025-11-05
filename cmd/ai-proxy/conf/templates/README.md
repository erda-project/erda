# Template

Template for model and service-provider.

Each JSON file is a dictionary of templates. The file name is only a convenient reference;
the **top-level key inside the JSON map is the true template name**. Multiple templates could
reside in a single file, though we currently keep one template per file for clarity.

## Available Templates

### Service-Provider

- [azure-ai-foundry](./service_provider/azure-ai-foundry.json)
- [aliyun-bailian](./service_provider/aliyun-bailian.json)
- [volcengine-ark](./service_provider/volcengine-ark.json)
- [aws-bedrock](./service_provider/aws-bedrock.json)
- [openai-compatible](./service_provider/openai-compatible.json)

### Model

- `type` reflects the output modality in ai-proxy. If a model only returns text—e.g., GPT-5 with image input but text-only output—keep it as `text_generation`. List the true input/output modalities under `metadata.public.abilities.{input_modalities, output_modalities}`.
