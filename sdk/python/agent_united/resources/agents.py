"""Agents resource"""

from typing import TYPE_CHECKING, Optional

if TYPE_CHECKING:
    from ..client import AgentClient


class Agents:
    """
    Manage agents and API keys.
    
    Accessible via: client.agents
    """
    
    def __init__(self, client: "AgentClient"):
        self._client = client
    
    def list(self) -> dict:
        """
        List all agents.
        
        Returns:
            {"agents": [...]}
        """
        return self._client.get("/api/v1/agents")
    
    def get(self, agent_id: str) -> dict:
        """
        Get agent details.
        
        Args:
            agent_id: Agent ID
        
        Returns:
            Agent object
        """
        return self._client.get(f"/api/v1/agents/{agent_id}")
    
    def create_api_key(
        self,
        agent_id: str,
        name: str,
        expires_at: Optional[str] = None,
    ) -> dict:
        """
        Create API key for agent.
        
        Args:
            agent_id: Agent ID
            name: Key name/label
            expires_at: ISO 8601 timestamp (optional)
        
        Returns:
            {"api_key": "au_...", "api_key_id": "key_..."}
        
        Warning:
            API key is returned ONCE. Store it securely.
        """
        payload = {"name": name}
        if expires_at:
            payload["expires_at"] = expires_at
        
        return self._client.post(
            f"/api/v1/agents/{agent_id}/keys",
            json=payload
        )
    
    def revoke_api_key(self, agent_id: str, key_id: str) -> dict:
        """
        Revoke API key.
        
        Args:
            agent_id: Agent ID
            key_id: API key ID
        
        Returns:
            Empty dict on success
        """
        return self._client.delete(
            f"/api/v1/agents/{agent_id}/keys/{key_id}"
        )
