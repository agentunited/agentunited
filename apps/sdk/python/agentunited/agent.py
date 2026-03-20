from __future__ import annotations

import json
import time
from dataclasses import dataclass
from datetime import datetime
from typing import Callable, Optional

import requests


@dataclass
class MessageEvent:
    channel_id: str
    message_id: str
    author_id: str
    author_type: str
    text: str
    created_at: datetime


class AUAgent:
    def __init__(self, instance_url: str, api_key: str, transport: str = "sse", poll_interval: float = 2.0, poll_channel_id: Optional[str] = None):
        self.instance_url = instance_url.rstrip("/")
        self.api_key = api_key
        self.transport = transport
        self.poll_interval = poll_interval
        self._handler: Optional[Callable[[MessageEvent], Optional[str]]] = None
        self._last_event_id: Optional[str] = None
        self._last_since: Optional[str] = None
        self._poll_channel_id = poll_channel_id
        self._seen_ids: set[str] = set()

    def on_message(self, func: Callable[[MessageEvent], Optional[str]]) -> Callable[[MessageEvent], Optional[str]]:
        self._handler = func
        return func

    def reply(self, channel_id: str, text: str) -> None:
        self.send(channel_id, text)

    def send(self, channel_id: str, text: str) -> None:
        requests.post(
            f"{self.instance_url}/api/v1/channels/{channel_id}/messages",
            headers={"Authorization": f"Bearer {self.api_key}", "Content-Type": "application/json"},
            json={"text": text},
            timeout=15,
        ).raise_for_status()

    def run(self) -> None:
        if not self._handler:
            raise RuntimeError("register a handler with @agent.on_message")
        if self.transport == "polling":
            self._run_polling()
        else:
            self._run_sse_with_fallback()

    def _run_sse_with_fallback(self) -> None:
        backoff = 1
        while True:
            try:
                self._run_sse_once()
                backoff = 1
            except Exception:
                time.sleep(backoff)
                backoff = min(backoff * 2, 30)
                self._run_polling_once()

    def _run_sse_once(self) -> None:
        headers = {"Authorization": f"Bearer {self.api_key}"}
        if self._last_event_id:
            headers["Last-Event-ID"] = self._last_event_id

        with requests.get(f"{self.instance_url}/api/v1/events/stream", headers=headers, stream=True, timeout=65) as resp:
            resp.raise_for_status()
            event_id = None
            event_data = None
            for raw in resp.iter_lines(decode_unicode=True):
                if raw is None:
                    continue
                line = raw.strip()
                if line.startswith(":"):
                    continue
                if line.startswith("id:"):
                    event_id = line[3:].strip()
                elif line.startswith("data:"):
                    event_data = line[5:].strip()
                elif line == "":
                    if event_data:
                        payload = json.loads(event_data)
                        self._handle_payload(payload)
                        if event_id:
                            self._last_event_id = event_id
                    event_id = None
                    event_data = None

    def _run_polling(self) -> None:
        while True:
            self._run_polling_once()
            time.sleep(self.poll_interval)

    def _run_polling_once(self) -> None:
        if not self._poll_channel_id:
            return
        if not self._last_since:
            self._last_since = datetime.utcnow().isoformat() + "Z"
        resp = requests.get(
            f"{self.instance_url}/api/v1/messages",
            headers={"Authorization": f"Bearer {self.api_key}"},
            params={"channel_id": self._poll_channel_id, "since": self._last_since},
            timeout=15,
        )
        resp.raise_for_status()
        data = resp.json()
        for payload in data.get("messages", []):
            if payload.get("id") in self._seen_ids:
                continue
            normalized = {
                "channel_id": payload.get("channel_id"),
                "message_id": payload.get("id"),
                "author_id": payload.get("author_id"),
                "author_type": payload.get("author_type"),
                "text": payload.get("text", ""),
                "created_at": payload.get("created_at"),
            }
            self._handle_payload(normalized)
            if payload.get("id"):
                self._seen_ids.add(payload["id"])
                if len(self._seen_ids) > 5000:
                    self._seen_ids = set(list(self._seen_ids)[-2500:])
        self._last_since = datetime.utcnow().isoformat() + "Z"

    def _handle_payload(self, payload: dict) -> None:
        if payload.get("author_type") != "human":
            return
        ev = MessageEvent(
            channel_id=payload["channel_id"],
            message_id=payload["message_id"],
            author_id=payload["author_id"],
            author_type=payload.get("author_type", ""),
            text=payload.get("text", ""),
            created_at=datetime.fromisoformat(payload["created_at"].replace("Z", "+00:00")),
        )
        result = self._handler(ev) if self._handler else None
        if isinstance(result, str) and result.strip():
            self.reply(ev.channel_id, result)
