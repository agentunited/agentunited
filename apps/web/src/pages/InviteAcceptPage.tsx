import { useState, useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
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
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <Card className="max-w-md w-full p-6">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
            <p className="mt-4 text-gray-600">Loading invite...</p>
          </div>
        </Card>
      </div>
    );
  }

  if (error && !inviteInfo) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <Card className="max-w-md w-full p-6">
          <div className="text-center">
            <h1 className="text-xl font-semibold text-gray-900 mb-4">Agent United</h1>
            <div className="bg-red-50 border border-red-200 rounded-md p-4 text-red-700">
              {error}
            </div>
            <Button
              onClick={() => navigate('/login')}
              className="mt-4"
              variant="secondary"
            >
              Go to Login
            </Button>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <Card className="max-w-md w-full p-6">
        <div className="text-center mb-6">
          <h1 className="text-xl font-semibold text-gray-900 mb-2">Agent United</h1>
          <h2 className="text-lg text-gray-700 mb-4">Welcome to Agent United</h2>
          
          {inviteInfo && (
            <div className="text-left bg-gray-50 rounded-md p-4 mb-6">
              <p className="text-sm text-gray-600 mb-2">
                You've been invited by <span className="font-medium">{inviteInfo.inviter}</span>
              </p>
              
              <div className="mb-2">
                <label className="text-sm font-medium text-gray-700">Email:</label>
                <p className="text-sm text-gray-900">{inviteInfo.email}</p>
              </div>
              
              <div>
                <label className="text-sm font-medium text-gray-700">Role:</label>
                <span className={`inline-block ml-2 px-2 py-1 text-xs font-medium rounded ${
                  inviteInfo.role === 'member' 
                    ? 'bg-blue-100 text-blue-800' 
                    : 'bg-gray-100 text-gray-800'
                }`}>
                  {inviteInfo.role === 'member' ? 'Member' : 'Observer'}
                </span>
              </div>
            </div>
          )}
        </div>

        {error && (
          <div className="mb-4 bg-red-50 border border-red-200 rounded-md p-3 text-red-700 text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <p className="text-sm text-gray-700 mb-4">Set your password to join:</p>
          
          <div>
            <Input
              label="Password"
              type="password"
              value={password}
              onChange={handlePasswordChange}
              error={passwordError || undefined}
              placeholder="At least 12 characters"
              required
            />
          </div>
          
          <div>
            <Input
              label="Confirm Password"
              type="password"
              value={confirmPassword}
              onChange={handleConfirmChange}
              error={confirmError || undefined}
              placeholder="Confirm your password"
              required
            />
          </div>
          
          <Button
            type="submit"
            disabled={!canSubmit}
            className="w-full"
          >
            {submitting ? 'Joining...' : 'Join Workspace'}
          </Button>
        </form>
      </Card>
    </div>
  );
}