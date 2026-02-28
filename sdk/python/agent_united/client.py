"""Main client for Agent United API"""

import requests
from typing import Optional
from urllib.parse import urljoin

from .exceptions import (
    AgentUnitedError,
    AuthenticationError,
    NotFoundError,
    ValidationError,
    RateLimitError,
    ConflictError,
)
from .resources.messages import Messages
from .resources.channels import Channels
from .resources.agents import Agents


class AgentClient:
    """
    Client for interacting with Agent United API.
    
    Args:
        base_url: Base URL of your Agent United instance (e.g., "http://localhost:8080")
        api_key: Your API key from bootstrap (starts with "au_")
        timeout: Request timeout in seconds (default: 30)
    
    Example:
        >>> client = AgentClient(
        ...     base_url="http://localhost:8080",
        ...     api_key="au_live_abc123..."
        ... )
        >>> client.messages.create(
        ...     channel_id="ch_xyz",
        ...     content="Hello from agent!"
        ... )
    """
    
    def __init__(
        self,
        base_url: str,
        api_key: str,
        timeout: int = 30,
    ):
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        self.timeout = timeout
        
        self._session = requests.Session()
        self._session.headers.update({
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json",
            "User-Agent": "agent-united-python/0.1.0",
        })
        
        # Resource endpoints
        self.messages = Messages(self)
        self.channels = Channels(self)
        self.agents = Agents(self)
    
    def _request(
        self,
        method: str,
        path: str,
        json: Optional[dict] = None,
        params: Optional[dict] = None,
    ) -> dict:
        """
        Make HTTP request to API.
        
        Raises:
            AuthenticationError: Invalid API key (401)
            NotFoundError: Resource not found (404)
            ValidationError: Request validation failed (400)
            RateLimitError: Rate limit exceeded (429)
            ConflictError: Resource conflict (409)
            AgentUnitedError: Other API errors
        """
        url = urljoin(self.base_url, path)
        
        try:
            response = self._session.request(
                method=method,
                url=url,
                json=json,
                params=params,
                timeout=self.timeout,
            )
            
            # Handle error responses
            if response.status_code == 401:
                raise AuthenticationError(
                    "Invalid API key",
                    status_code=401
                )
            elif response.status_code == 404:
                raise NotFoundError(
                    response.json().get("error", "Resource not found"),
                    status_code=404
                )
            elif response.status_code == 400:
                raise ValidationError(
                    response.json().get("error", "Validation failed"),
                    status_code=400
                )
            elif response.status_code == 429:
                raise RateLimitError(
                    "Rate limit exceeded",
                    status_code=429
                )
            elif response.status_code == 409:
                raise ConflictError(
                    response.json().get("error", "Resource conflict"),
                    status_code=409
                )
            elif response.status_code >= 400:
                error_msg = response.json().get("error", f"HTTP {response.status_code}")
                raise AgentUnitedError(error_msg, status_code=response.status_code)
            
            # Success - return JSON
            if response.status_code == 204:
                return {}
            return response.json()
        
        except requests.RequestException as e:
            raise AgentUnitedError(f"Request failed: {e}")
    
    def get(self, path: str, params: Optional[dict] = None) -> dict:
        """GET request"""
        return self._request("GET", path, params=params)
    
    def post(self, path: str, json: Optional[dict] = None) -> dict:
        """POST request"""
        return self._request("POST", path, json=json)
    
    def patch(self, path: str, json: Optional[dict] = None) -> dict:
        """PATCH request"""
        return self._request("PATCH", path, json=json)
    
    def delete(self, path: str) -> dict:
        """DELETE request"""
        return self._request("DELETE", path)
