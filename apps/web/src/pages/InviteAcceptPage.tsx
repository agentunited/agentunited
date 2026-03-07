import { useState, useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
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
      localStorage.setItem('auth-token', response.jwt_token);
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

  return (
    <div
      className="relative flex min-h-screen items-center justify-center px-4 py-8"
      style={{
        background:
          'radial-gradient(circle at top left, rgba(16, 185, 129, 0.12), transparent 45%), radial-gradient(circle at 80% 20%, rgba(14, 165, 233, 0.08), transparent 40%), #f8fafc',
      }}
    >
      <div className="w-full max-w-md">
        <div className="mb-6 flex flex-col items-center gap-3">
          <div className="flex h-11 w-11 items-center justify-center rounded-full border border-emerald-500/50 shadow-[0_0_20px_rgba(16,185,129,0.35)]">
            <div className="h-2.5 w-2.5 rounded-full bg-emerald-500" />
          </div>
          <h1 className="text-xl font-semibold tracking-tight text-slate-900">Agent United</h1>
        </div>

        <div className="rounded-2xl border border-white/70 bg-white/90 p-7 shadow-[0_24px_80px_-48px_rgba(15,23,42,0.3)] backdrop-blur">
          {loading ? (
            <div className="py-6 text-center">
              <div className="mx-auto h-8 w-8 animate-spin rounded-full border-2 border-emerald-500 border-t-transparent" />
              <p className="mt-4 text-sm text-slate-600">Loading invite…</p>
            </div>
          ) : error && !inviteInfo ? (
            <div className="text-center">
              <h2 className="mb-3 text-lg font-semibold text-slate-900">Invite unavailable</h2>
              <div className="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
              <button
                onClick={() => navigate('/login')}
                className="mt-5 w-full rounded-xl bg-slate-200 px-4 py-2.5 text-sm font-semibold text-slate-700 transition-colors hover:bg-slate-300"
              >
                Go to login
              </button>
            </div>
          ) : (
            <>
              <h2 className="text-lg font-semibold text-slate-900">Join your workspace</h2>
              <p className="mt-1 text-sm text-slate-500">Set your password to accept this invite and continue.</p>

              {inviteInfo && (
                <div className="mt-5 rounded-xl border border-emerald-100 bg-emerald-50/70 p-4 text-sm">
                  <p className="text-slate-600">
                    Invited by <span className="font-semibold text-slate-900">{inviteInfo.inviter}</span>
                  </p>
                  <p className="mt-1 text-slate-600">
                    Email: <span className="font-medium text-slate-900">{inviteInfo.email}</span>
                  </p>
                </div>
              )}

              {error && (
                <div className="mt-4 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
              )}

              <form onSubmit={handleSubmit} className="mt-5 space-y-4">
                <div>
                  <label className="mb-1.5 block text-xs font-semibold tracking-[0.1em] text-slate-600 uppercase">Password</label>
                  <input
                    type="password"
                    value={password}
                    onChange={handlePasswordChange}
                    className="w-full rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm text-slate-900 outline-none transition-all placeholder:text-slate-400 focus:border-emerald-400/60 focus:ring-4 focus:ring-emerald-500/10"
                    placeholder="At least 12 characters"
                    required
                  />
                  {passwordError && <p className="mt-1 text-xs text-red-500">{passwordError}</p>}
                </div>

                <div>
                  <label className="mb-1.5 block text-xs font-semibold tracking-[0.1em] text-slate-600 uppercase">Confirm password</label>
                  <input
                    type="password"
                    value={confirmPassword}
                    onChange={handleConfirmChange}
                    className="w-full rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm text-slate-900 outline-none transition-all placeholder:text-slate-400 focus:border-emerald-400/60 focus:ring-4 focus:ring-emerald-500/10"
                    placeholder="Confirm your password"
                    required
                  />
                  {confirmError && <p className="mt-1 text-xs text-red-500">{confirmError}</p>}
                </div>

                <button
                  type="submit"
                  disabled={!canSubmit}
                  className="w-full rounded-xl bg-emerald-500 px-4 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-emerald-600 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {submitting ? 'Joining…' : 'Join workspace'}
                </button>
              </form>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
