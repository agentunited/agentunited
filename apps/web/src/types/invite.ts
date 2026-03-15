export interface InviteInfo {
  email: string;
  display_name: string;
  role: 'observer' | 'member';
  inviter: string;
}

export interface InviteAcceptRequest {
  token: string;
  password: string;
  display_name?: string;
}

export interface InviteAcceptResponse {
  user_id: string;
  jwt_token: string;
  redirect: string;
  dm_channel_id?: string;
}

export interface InviteError {
  type: 'invalid_token' | 'expired_token' | 'already_used' | 'weak_password' | 'network_error';
  message: string;
}