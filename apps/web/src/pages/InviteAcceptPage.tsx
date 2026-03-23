import { useState, useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { getInviteInfo, acceptInvite, InviteApiError } from '../services/inviteApi';
import { initializeFromUrlParams } from '../services/apiConfig';
import type { InviteInfo } from '../types/invite';
import { AppDownloadBanner } from '../components/invite/AppDownloadBanner';

export function InviteAcceptPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const token = searchParams.get('token');

  const [inviteInfo, setInviteInfo] = useState<InviteInfo | null>(null);
  const [displayName, setDisplayName] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [displayNameError, setDisplayNameError] = useState<string | null>(null);
  const [passwordError, setPasswordError] = useState<string | null>(null);
  const [confirmError, setConfirmError] = useState<string | null>(null);

  const isValidDisplayName = displayName.trim().length >= 1 && displayName.trim().length <= 50;
  const isValidPassword = password.length >= 12;
  const passwordsMatch = password === confirmPassword;
  const canSubmit = isValidDisplayName && isValidPassword && passwordsMatch && !submitting;

  useEffect(() => {
    // Initialize instance URL from URL params before any API calls
    // This fixes a race condition where App.tsx's useEffect runs after this one
    initializeFromUrlParams();

    if (!token) {
      setError('Invalid invite link. Check the URL.');
      setLoading(false);
      return;
    }

    const fetchInviteInfo = async () => {
      try {
        const info = await getInviteInfo(token);
        setInviteInfo(info);
        setDisplayName((info.display_name || '').trim());
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

  const handleDisplayNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const next = e.target.value;
    setDisplayName(next);

    const trimmed = next.trim();
    if (!trimmed) {
      setDisplayNameError('Display name is required');
    } else if (trimmed.length > 50) {
      setDisplayNameError('Display name must be 50 characters or less');
    } else {
      setDisplayNameError(null);
    }
  };

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

    const trimmedDisplayName = displayName.trim();
    if (!trimmedDisplayName || trimmedDisplayName.length > 50) {
      setDisplayNameError('Display name is required (1–50 chars)');
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      const response = await acceptInvite({ token, password, display_name: trimmedDisplayName });
      localStorage.setItem('auth-token', response.jwt_token);
      // Decode JWT payload to persist user-id (used by DM modal filter + own-message detection)
      try {
        const payload = JSON.parse(atob(response.jwt_token.split('.')[1]));
        if (payload.user_id) localStorage.setItem('user-id', payload.user_id);
        if (payload.email) localStorage.setItem('user-email', payload.email);
      } catch { /* ignore decode errors */ }
      if (response.dm_channel_id) {
        navigate(`/chat/dm/${response.dm_channel_id}`);
      } else {
        navigate('/chat?first_login=1');
      }
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

        <AppDownloadBanner inviteToken={token ?? ''} instanceUrl={window.location.origin} />

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
                className="mt-5 w-full rounded-xl bg-primary px-4 py-2.5 text-sm font-semibold text-primary-foreground transition-colors hover:bg-primary/90"
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
                    You&apos;ve been invited to join an Agent United workspace.
                  </p>
                  <p className="mt-1 text-slate-600">
                    This invite is for <span className="font-medium text-slate-900">{inviteInfo.email}</span>
                  </p>
                </div>
              )}

              {error && (
                <div className="mt-4 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
              )}

              <form onSubmit={handleSubmit} className="mt-5 space-y-4">
                <div>
                  <label className="mb-1.5 block text-xs font-semibold tracking-[0.1em] text-slate-600 uppercase">Email</label>
                  <div className="w-full rounded-xl border border-slate-200 bg-slate-100 px-4 py-2.5 text-sm text-slate-700">
                    {inviteInfo?.email || '—'}
                  </div>
                </div>

                <div>
                  <label className="mb-1.5 block text-xs font-semibold tracking-[0.1em] text-slate-600 uppercase">Display name</label>
                  <input
                    type="text"
                    value={displayName}
                    onChange={handleDisplayNameChange}
                    className="w-full rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm text-slate-900 outline-none transition-all placeholder:text-slate-400 focus:border-emerald-400/60 focus:ring-4 focus:ring-emerald-500/10"
                    placeholder="How should we call you?"
                    maxLength={50}
                    required
                  />
                  {displayNameError && <p className="mt-1 text-xs text-red-500">{displayNameError}</p>}
                </div>

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
