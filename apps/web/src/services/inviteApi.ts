import type { InviteInfo, InviteAcceptRequest, InviteAcceptResponse, InviteError } from '../types/invite';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

export class InviteApiError extends Error {
  public error: InviteError;
  
  constructor(error: InviteError) {
    super(error.message);
    this.name = 'InviteApiError';
    this.error = error;
  }
}

export async function getInviteInfo(token: string): Promise<InviteInfo> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/invite?token=${encodeURIComponent(token)}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (response.status === 404) {
      throw new InviteApiError({
        type: 'invalid_token',
        message: 'Invalid invite link. Check the URL.',
      });
    }

    if (response.status === 410) {
      throw new InviteApiError({
        type: 'expired_token',
        message: 'This invite has expired. Contact the agent for a new invite.',
      });
    }

    if (!response.ok) {
      throw new InviteApiError({
        type: 'network_error',
        message: 'Could not connect. Check your connection and try again.',
      });
    }

    return await response.json();
  } catch (error) {
    if (error instanceof InviteApiError) {
      throw error;
    }
    throw new InviteApiError({
      type: 'network_error',
      message: 'Could not connect. Check your connection and try again.',
    });
  }
}

export async function acceptInvite(request: InviteAcceptRequest): Promise<InviteAcceptResponse> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/invite/accept`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (response.status === 404) {
      throw new InviteApiError({
        type: 'invalid_token',
        message: 'Invalid invite link. Check the URL.',
      });
    }

    if (response.status === 410) {
      throw new InviteApiError({
        type: 'expired_token',
        message: 'This invite has expired. Contact the agent for a new invite.',
      });
    }

    if (response.status === 409) {
      throw new InviteApiError({
        type: 'already_used',
        message: 'This invite was already used. Log in instead.',
      });
    }

    if (response.status === 400) {
      const error = await response.json();
      if (error.message?.includes('password')) {
        throw new InviteApiError({
          type: 'weak_password',
          message: 'Password must be at least 12 characters',
        });
      }
    }

    if (!response.ok) {
      throw new InviteApiError({
        type: 'network_error',
        message: 'Could not connect. Check your connection and try again.',
      });
    }

    return await response.json();
  } catch (error) {
    if (error instanceof InviteApiError) {
      throw error;
    }
    throw new InviteApiError({
      type: 'network_error',
      message: 'Could not connect. Check your connection and try again.',
    });
  }
}