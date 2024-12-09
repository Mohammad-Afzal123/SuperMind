# Langflow Proxy

## Description

This is a proxy service that allow you to interact with Langflow workflows using OPENAI API.

## Curl Example

```bash
curl http://langflow-proxy/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer [LANGFLOW-API-KEY]" \
  -d '{
  "model": "workflow-id",
  "messages": [
    {
      "role": "user",
      "content": "Who are you ?"
    }
  ]
}'
```
