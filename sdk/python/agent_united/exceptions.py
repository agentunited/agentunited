"""Exceptions for Agent United SDK"""


class AgentUnitedError(Exception):
    """Base exception for all Agent United errors"""
    
    def __init__(self, message: str, status_code: int = None):
        super().__init__(message)
        self.message = message
        self.status_code = status_code


class AuthenticationError(AgentUnitedError):
    """API key is invalid or missing"""
    pass


class NotFoundError(AgentUnitedError):
    """Resource not found (404)"""
    pass


class ValidationError(AgentUnitedError):
    """Request validation failed (400)"""
    pass


class RateLimitError(AgentUnitedError):
    """Rate limit exceeded (429)"""
    pass


class ConflictError(AgentUnitedError):
    """Resource conflict (409)"""
    pass
