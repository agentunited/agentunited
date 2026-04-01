import { useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { AlertCircle, ArrowLeft, CheckCircle2, Eye, EyeOff, Loader2 } from 'lucide-react';
import { resetPassword } from '../services/authApi';

type PageState =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'success' }
  | { status: 'error'; message: string };

export function ResetPasswordPage() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get('token') ?? '';
  const navigate = useNavigate();

  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);
  const [state, setState] = useState<PageState>({ status: 'idle' });

  const validationError =
    password.length > 0 && password.length < 8
      ? 'Password must be at least 8 characters'
      : password !== confirm && confirm.length > 0
        ? 'Passwords do not match'
        : null;

  const canSubmit = !!token && password.length >= 8 && password === confirm && state.status !== 'loading';

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!canSubmit) return;
    setState({ status: 'loading' });
    try {
      await resetPassword(token, password);
      setState({ status: 'success' });
      setTimeout(() => navigate('/login?reset=1'), 2000);
    } catch (err) {
      const message =
        err instanceof Error && err.message
          ? err.message
          : 'This reset link has expired or is invalid. Please request a new one.';
      setState({ status: 'error', message });
    }
  };

  if (!token) {
    return <InvalidLink />;
  }

  return (
    <div
      className="relative flex min-h-screen flex-col items-center justify-center px-4"
      style={{
        fontFamily: 'Manrope, system-ui, -apple-system, sans-serif',
        background:
          'radial-gradient(circle at top left, rgba(16, 185, 129, 0.12), transparent 45%), radial-gradient(circle at 80% 20%, rgba(14, 165, 233, 0.08), transparent 40%), #f8fafc',
      }}
    >
      <div className="mb-8 flex flex-col items-center gap-3">
        <div className="flex h-12 w-12 items-center justify-center rounded-full border border-emerald-500/50 shadow-[0_0_20px_rgba(16,185,129,0.35)]">
          <div className="h-3 w-3 rounded-full bg-emerald-500" />
        </div>
        <h1 className="text-xl font-semibold tracking-tight text-slate-900">Agent United</h1>
      </div>

      <div className="w-full max-w-sm rounded-2xl border border-white/70 bg-white/90 p-8 shadow-[0_24px_80px_-48px_rgba(15,23,42,0.3)] backdrop-blur">
        {state.status === 'success' ? (
          <SuccessState />
        ) : (
          <>
            <h2 className="mb-1 text-lg font-semibold text-slate-900">Set a new password</h2>
            <p className="mb-6 text-sm text-slate-500">Choose a strong password for your account.</p>

            {state.status === 'error' && (
              <div
                className="mb-4 flex items-center gap-2 rounded-xl border border-red-200 bg-red-50 px-4 py-3"
                role="alert"
              >
                <AlertCircle className="h-4 w-4 shrink-0 text-red-500" aria-hidden="true" />
                <p className="text-sm text-red-700">{state.message}</p>
              </div>
            )}

            <form onSubmit={handleSubmit} className="space-y-4">
              <PasswordField
                id="rp-password"
                label="New password"
                value={password}
                onChange={setPassword}
                show={showPassword}
                onToggleShow={() => setShowPassword((v) => !v)}
                autoComplete="new-password"
                disabled={state.status === 'loading'}
              />

              <PasswordField
                id="rp-confirm"
                label="Confirm password"
                value={confirm}
                onChange={setConfirm}
                show={showConfirm}
                onToggleShow={() => setShowConfirm((v) => !v)}
                autoComplete="new-password"
                disabled={state.status === 'loading'}
              />

              {validationError && (
                <p className="text-xs text-red-500" role="alert">{validationError}</p>
              )}

              <button
                type="submit"
                disabled={!canSubmit}
                className="mt-2 w-full rounded-xl bg-emerald-500 py-2.5 text-sm font-semibold text-white shadow-sm transition-all hover:-translate-y-0.5 hover:bg-emerald-600 disabled:translate-y-0 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {state.status === 'loading' ? (
                  <span className="flex items-center justify-center gap-2">
                    <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" /> Saving…
                  </span>
                ) : (
                  'Set new password'
                )}
              </button>
            </form>
          </>
        )}
      </div>

      {state.status !== 'success' && (
        <Link
          to="/login"
          className="mt-6 flex items-center gap-1.5 text-xs text-slate-400 transition-colors hover:text-slate-600"
        >
          <ArrowLeft className="h-3.5 w-3.5" />
          Back to sign in
        </Link>
      )}
    </div>
  );
}

interface PasswordFieldProps {
  id: string;
  label: string;
  value: string;
  onChange: (v: string) => void;
  show: boolean;
  onToggleShow: () => void;
  autoComplete: string;
  disabled: boolean;
}

function PasswordField({ id, label, value, onChange, show, onToggleShow, autoComplete, disabled }: PasswordFieldProps) {
  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className="text-xs font-semibold tracking-[0.1em] text-slate-600 uppercase">
        {label}
      </label>
      <div className="relative">
        <input
          id={id}
          type={show ? 'text' : 'password'}
          autoComplete={autoComplete}
          placeholder="••••••••"
          disabled={disabled}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          required
          className="w-full rounded-xl border border-slate-200 bg-white px-4 py-2.5 pr-10 text-sm text-slate-900 outline-none transition-all placeholder:text-slate-400 focus:border-emerald-400/60 focus:ring-4 focus:ring-emerald-500/10 disabled:opacity-70"
        />
        <button
          type="button"
          onClick={onToggleShow}
          aria-label={show ? 'Hide password' : 'Show password'}
          className="absolute top-1/2 right-3 -translate-y-1/2 text-slate-400 transition-colors hover:text-slate-600"
        >
          {show ? <EyeOff className="h-4 w-4" aria-hidden="true" /> : <Eye className="h-4 w-4" aria-hidden="true" />}
        </button>
      </div>
    </div>
  );
}

function SuccessState() {
  return (
    <div className="flex flex-col items-center gap-4 py-2 text-center">
      <CheckCircle2 className="h-10 w-10 text-emerald-500" aria-hidden="true" />
      <div>
        <h2 className="text-lg font-semibold text-slate-900">Password updated</h2>
        <p className="mt-2 text-sm text-slate-500">
          Your password has been reset. Redirecting you to sign in…
        </p>
      </div>
    </div>
  );
}

function InvalidLink() {
  return (
    <div
      className="relative flex min-h-screen flex-col items-center justify-center px-4"
      style={{
        fontFamily: 'Manrope, system-ui, -apple-system, sans-serif',
        background:
          'radial-gradient(circle at top left, rgba(16, 185, 129, 0.12), transparent 45%), radial-gradient(circle at 80% 20%, rgba(14, 165, 233, 0.08), transparent 40%), #f8fafc',
      }}
    >
      <div className="w-full max-w-sm rounded-2xl border border-white/70 bg-white/90 p-8 shadow-[0_24px_80px_-48px_rgba(15,23,42,0.3)] backdrop-blur text-center">
        <AlertCircle className="mx-auto h-10 w-10 text-red-400" aria-hidden="true" />
        <h2 className="mt-4 text-lg font-semibold text-slate-900">Invalid reset link</h2>
        <p className="mt-2 text-sm text-slate-500">
          This link is missing a reset token. Please use the link from your email.
        </p>
        <Link
          to="/forgot-password"
          className="mt-6 block w-full rounded-xl bg-emerald-500 py-2.5 text-sm font-semibold text-white shadow-sm transition-all hover:-translate-y-0.5 hover:bg-emerald-600"
        >
          Request a new link
        </Link>
      </div>
    </div>
  );
}
