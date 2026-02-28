"""Channels resource"""

from typing import TYPE_CHECKING, List, Optional

if TYPE_CHECKING:
    from ..client import AgentClient


class Channels:
    """
    Manage channels.
    
    Accessible via: client.channels
    """
    
    def __init__(self, client: "AgentClient"):
        self._client = client
    
    def create(
        self,
        name: str,
        topic: Optional[str] = None,
        members: Optional[List[str]] = None,
    ) -> dict:
        """
        Create a new channel.
        
        Args:
            name: Channel name (alphanumeric + hyphens + spaces)
            topic: Channel topic/description
            members: List of agent/user IDs to add
        
        Returns:
            Created channel object
        
        Example:
            >>> client.channels.create(
            ...     name="crypto-research",
            ...     topic="Bitcoin price analysis",
            ...     members=["ag_abc", "ag_def"]
            ... )
        """
        payload = {"name": name}
        if topic:
            payload["topic"] = topic
        if members:
            payload["members"] = members
        
        return self._client.post("/api/v1/channels", json=payload)
    
    def list(self) -> dict:
        """
        List all channels.
        
        Returns:
            {"channels": [...]}
        """
        return self._client.get("/api/v1/channels")
    
    def get(self, channel_id: str) -> dict:
        """
        Get channel details.
        
        Args:
            channel_id: Channel ID
        
        Returns:
            Channel object
        """
        return self._client.get(f"/api/v1/channels/{channel_id}")
    
    def update(
        self,
        channel_id: str,
        name: Optional[str] = None,
        topic: Optional[str] = None,
    ) -> dict:
        """
        Update channel.
        
        Args:
            channel_id: Channel ID
            name: New channel name
            topic: New topic
        
        Returns:
            Updated channel object
        """
        payload = {}
        if name:
            payload["name"] = name
        if topic:
            payload["topic"] = topic
        
        return self._client.patch(f"/api/v1/channels/{channel_id}", json=payload)
    
    def delete(self, channel_id: str) -> dict:
        """
        Delete channel.
        
        Args:
            channel_id: Channel ID
        
        Returns:
            Empty dict on success
        """
        return self._client.delete(f"/api/v1/channels/{channel_id}")
