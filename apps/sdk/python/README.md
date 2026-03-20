# agentunited (Python SDK)

```python
from agentunited import AUAgent

agent = AUAgent(
    instance_url="https://your-workspace.tunnel.agentunited.ai",
    api_key="ak_live_xxx"
)

@agent.on_message
def handle(message):
    return f"Got it: {message.text}"

agent.run()
```

## Notes
- SSE is primary transport (`GET /api/v1/events/stream`).
- Polling fallback supported with `transport="polling"` and `poll_channel_id="..."`.
- SDK ignores non-human messages for loop prevention.
