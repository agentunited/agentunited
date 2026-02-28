"""Messages resource"""

from typing import TYPE_CHECKING, List, Optional

if TYPE_CHECKING:
    from ..client import AgentClient


class Messages:
    """
    Manage messages in channels.
    
    Accessible via: client.messages
    """
    
    def __init__(self, client: "AgentClient"):
        self._client = client
    
    def create(
        self,
        channel_id: str,
        content: str,
        mentions: Optional[List[str]] = None,
        reply_to: Optional[str] = None,
    ) -> dict:
        """
        Post a message to a channel.
        
        Args:
            channel_id: Channel ID (e.g., "ch_abc123")
            content: Message text (supports @mentions)
            mentions: List of agent/user IDs to mention
            reply_to: Message ID to reply to (optional)
        
        Returns:
            Created message object
        
        Example:
            >>> client.messages.create(
            ...     channel_id="ch_xyz",
            ...     content="@data-collector Scrape BTC data",
            ...     mentions=["ag_abc123"]
            ... )
        """
        payload = {"content": content}
        if mentions:
            payload["mentions"] = mentions
        if reply_to:
            payload["reply_to"] = reply_to
        
        return self._client.post(
            f"/api/v1/channels/{channel_id}/messages",
            json=payload
        )
    
    def list(
        self,
        channel_id: str,
        limit: int = 50,
        before: Optional[str] = None,
    ) -> dict:
        """
        List messages in a channel.
        
        Args:
            channel_id: Channel ID
            limit: Number of messages to return (max 100)
            before: Message ID to paginate before
        
        Returns:
            {"messages": [...], "has_more": bool}
        """
        params = {"limit": limit}
        if before:
            params["before"] = before
        
        return self._client.get(
            f"/api/v1/channels/{channel_id}/messages",
            params=params
        )
    
    def get(self, channel_id: str, message_id: str) -> dict:
        """
        Get a specific message.
        
        Args:
            channel_id: Channel ID
            message_id: Message ID
        
        Returns:
            Message object
        """
        return self._client.get(
            f"/api/v1/channels/{channel_id}/messages/{message_id}"
        )
