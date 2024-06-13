# stream response chunk

## qwen-vl resp stream format: output.choices.[]message.content.[]text

```plain text
id:1
event:result
:HTTP_STATUS/200
data:{"output":{"choices":[{"message":{"content":[{"text":"Hello"}],"role":"assistant"},"finish_reason":"null"}]},"usage":{"input_tokens":92,"output_tokens":1},"request_id":"e2edfd5d-f5c3-958b-b058-f8341b31af05"}

id:5
event:result
:HTTP_STATUS/200
data:{"output":{"choices":[{"message":{"content":[{"text":"Hello! How can I assist you today?"}],"role":"assistant"},"finish_reason":"stop"}]},"usage":{"input_tokens":92,"output_tokens":10},"request_id":"e2edfd5d-f5c3-958b-b058-f8341b31af05"}
```

## qwen2-72b resp stream format: output.text

```text
id:1
event:result
:HTTP_STATUS/200
data:{"output":{"finish_reason":"null","text":"我是"},"usage":{"total_tokens":11,"input_tokens":10,"output_tokens":1},"request_id":"755310ab-4b61-9f11-acad-d5921c3ba239"}

id:2
event:result
:HTTP_STATUS/200
data:{"output":{"finish_reason":"null","text":"我是阿里"},"usage":{"total_tokens":12,"input_tokens":10,"output_tokens":2},"request_id":"c338a5f1-e8e7-9908-8000-5628cf65a1a2"}
```

## kimi resp stream format: output.choices.[]message.content.string (but string content is a json array string)

这里有一个很奇怪的返回值，返回的 choice.message.content 是一个 string，但是 string 的内容是一个 json
数组字符串。流式处理，必须将它存储在内存中进行片段判断

```text
id:9
event:result
:HTTP_STATUS/200
data:{"output":{"choices":[{"message":{"content":"[{\"text\": \"Okay","role":"assistant"},"index":0}]},"request_id":"6fba8399-3011-9126-ad1a-9833aba19f18"}

id:34
event:result
:HTTP_STATUS/200
data:{"output":{"choices":[{"message":{"content":"[{\"text\": \"Okay! I'm here to help. If you have any questions or need assistance, just let me know.\"}]","role":"assistant"},"index":0,"finish_reason":"stop"}]},"usage":{"total_tokens":87,"input_tokens":56,"output_tokens":31},"request_id":"6fba8399-3011-9126-ad1a-9833aba19f18"}
```

# non-stream response

```json
{
  "output": {
    "choices": [
      {
        "finish_reason": "stop",
        "message": {
          "role": "assistant",
          "content": [
            {
              "text": "Hello! How can I assist you today?"
            }
          ]
        }
      }
    ]
  },
  "usage": {
    "output_tokens": 10,
    "input_tokens": 92
  },
  "request_id": "cc7bcb16-0a9d-9ca4-ae78-b792393eb840"
}
```
