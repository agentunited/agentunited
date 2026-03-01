import { useState, useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { Card } from '../components/ui/Card';
import { getInviteInfo, acceptInvite, InviteApiError } from '../services/inviteApi';
import type { InviteInfo } from '../types/invite';

export function InviteAcceptPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const token = searchParams.get('token');

  const [inviteInfo, setInviteInfo] = useState<InviteInfo | null>(null);
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [passwordError, setPasswordError] = useState<string | null>(null);
  const [confirmError, setConfirmError] = useState<string | null>(null);

  const isValidPassword = password.length >= 12;
  const passwordsMatch = password === confirmPassword;
  const canSubmit = isValidPassword && passwordsMatch && !submitting;

  useEffect(() => {
    if (!token) {
      setError('Invalid invite link. Check the URL.');
      setLoading(false);
      return;
    }

    const fetchInviteInfo = async () => {
      try {
        const info = await getInviteInfo(token);
        setInviteInfo(info);
        setError(null);
      } catch (err) {
        if (err instanceof InviteApiError) {
          setError(err.error.message);
        } else {
          setError('Could not load invite information. Try again.');
        }
      } finally {
        setLoading(false);
      }
    };

    fetchInviteInfo();
  }, [token]);

  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newPassword = e.target.value;
    setPassword(newPassword);
    
    if (newPassword.length > 0 && newPassword.length < 12) {
      setPasswordError('Password must be at least 12 characters');
    } else {
      setPasswordError(null);
    }
  };

  const handleConfirmChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newConfirm = e.target.value;
    setConfirmPassword(newConfirm);
    
    if (newConfirm.length > 0 && newConfirm !== password) {
      setConfirmError('Passwords do not match');
    } else {
      setConfirmError(null);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!canSubmit || !token) {
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      const response = await acceptInvite({ token, password });
      
      // Store JWT token
      localStorage.setItem('auth-token', response.jwt_token);
      
      // Redirect to channels (or use response.redirect)
      navigate('/chat');
    } catch (err) {
      if (err instanceof InviteApiError) {
        setError(err.error.message);
      } else {
        setError('Could not connect. Try again.');
      }
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-4">
        <Card className="max-w-md w-full bg-card border border-border p-6">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
            <p className="mt-4 text-muted-foreground">Loading invite...</p>
          </div>
        </Card>
      </div>
    );
  }

  if (error && !inviteInfo) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-4">
        <Card className="max-w-md w-full bg-card border border-border p-6">
          <div className="text-center">
            <h1 className="text-xl font-semibold text-foreground mb-4">Agent United</h1>
            <div className="bg-destructive/10 border border-destructive/20 rounded-md p-4 text-destructive">
              {error}
            </div>
            <button
              onClick={() => navigate('/login')}
              className="mt-4 w-full rounded-lg bg-secondary px-4 py-2.5 text-sm font-semibold text-secondary-foreground hover:bg-secondary/80 transition-colors"
            >
              Go to Login
            </button>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <Card className="max-w-md w-full bg-card border border-border p-6">
        <div className="text-center mb-6">
          <h1 className="text-xl font-semibold text-foreground mb-2">Agent United</h1>
          <h2 className="text-lg text-foreground/80 mb-4">Welcome to Agent United</h2>
          
          {inviteInfo && (
            <div className="text-left bg-muted rounded-md p-4 mb-6">
              <p className="text-sm text-muted-foreground mb-2">
                You've been invited by <span className="font-medium text-foreground">{inviteInfo.inviter}</span>
              </p>
              
              <div className="mb-2">
                <label className="text-sm font-medium text-foreground">Email:</label>
                <p className="text-sm text-foreground/85">{inviteInfo.email}</p>
              </div>
              
              <div>
                <label className="text-sm font-medium text-foreground">Role:</label>
                <span className={`inline-block ml-2 px-2 py-1 text-xs font-medium rounded ${
                  inviteInfo.role === 'member' 
                    ? 'bg-primary/10 text-primary' 
                    : 'bg-secondary text-secondary-foreground'
                }`}>
                  {inviteInfo.role === 'member' ? 'Member' : 'Observer'}
                </span>
              </div>
            </div>
          )}
        </div>

        {error && (
          <div className="mb-4 bg-destructive/10 border border-destructive/20 rounded-md p-3 text-destructive text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <p className="text-sm text-muted-foreground mb-4">Set your password to join:</p>
          
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">Password</label>
            <input
              type="password"
              value={password}
              onChange={handlePasswordChange}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all"
              placeholder="At least 12 characters"
              required
            />
            {passwordError && (
              <p className="mt-1 text-xs text-destructive">{passwordError}</p>
            )}
          </div>
          
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">Confirm Password</label>
            <input
              type="password"
              value={confirmPassword}
              onChange={handleConfirmChange}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all"
              placeholder="Confirm your password"
              required
            />
            {confirmError && (
              <p className="mt-1 text-xs text-destructive">{confirmError}</p>
            )}
          </div>
          
          <button
            type="submit"
            disabled={!canSubmit}
            className="w-full rounded-lg bg-primary px-4 py-2.5 text-sm font-semibold text-primary-foreground hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {submitting ? 'Joining...' : 'Join Workspace'}
          </button>
        </form>
      </Card>
    </div>
  );
}