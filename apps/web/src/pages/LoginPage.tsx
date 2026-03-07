import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useNavigate } from 'react-router-dom';
import { AlertCircle, Eye, EyeOff, Loader2 } from 'lucide-react';

const loginSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(6, 'Password must be at least 6 characters'),
});

type LoginFormData = z.infer<typeof loginSchema>;

export function LoginPage() {
  const [isLoading, setIsLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = async (data: LoginFormData) => {
    setIsLoading(true);
    setError('');

    try {
      const { email, password } = data;
      const response = await fetch(`${import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'}/api/v1/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      });

      if (!response.ok) {
        throw new Error('Login failed');
      }

      const result = await response.json();
      localStorage.setItem('auth-token', result.token);
      localStorage.setItem('user-email', data.email);
      if (result.user_id) {
        localStorage.setItem('user-id', result.user_id);
      }

      navigate('/chat');
    } catch {
      setError('Invalid email or password');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div
      className="relative flex min-h-screen flex-col items-center justify-center px-4"
      style={{
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
        <h2 className="mb-1 text-lg font-semibold text-slate-900">Sign in</h2>
        <p className="mb-6 text-sm text-slate-500">Enter your credentials to continue</p>

        {error && (
          <div className="mb-4 flex items-center gap-2 rounded-xl border border-red-200 bg-red-50 px-4 py-3" role="alert">
            <AlertCircle className="h-4 w-4 shrink-0 text-red-500" />
            <p className="text-sm text-red-700">{error}</p>
          </div>
        )}

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-semibold tracking-[0.1em] text-slate-600 uppercase">Email address</label>
            <input
              type="email"
              autoComplete="email"
              placeholder="you@example.com"
              disabled={isLoading}
              className="rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm text-slate-900 outline-none transition-all placeholder:text-slate-400 focus:border-emerald-400/60 focus:ring-4 focus:ring-emerald-500/10 disabled:opacity-70"
              {...register('email')}
            />
            {errors.email && <p className="text-xs text-red-500">{errors.email.message}</p>}
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-semibold tracking-[0.1em] text-slate-600 uppercase">Password</label>
            <div className="relative">
              <input
                type={showPassword ? 'text' : 'password'}
                autoComplete="current-password"
                placeholder="••••••••"
                disabled={isLoading}
                className="w-full rounded-xl border border-slate-200 bg-white px-4 py-2.5 pr-10 text-sm text-slate-900 outline-none transition-all placeholder:text-slate-400 focus:border-emerald-400/60 focus:ring-4 focus:ring-emerald-500/10 disabled:opacity-70"
                {...register('password')}
              />
              <button
                type="button"
                onClick={() => setShowPassword((v) => !v)}
                className="absolute top-1/2 right-3 -translate-y-1/2 text-slate-400 transition-colors hover:text-slate-600"
                aria-label={showPassword ? 'Hide password' : 'Show password'}
              >
                {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </button>
            </div>
            {errors.password && <p className="text-xs text-red-500">{errors.password.message}</p>}
          </div>

          <button
            type="submit"
            disabled={isLoading}
            className="mt-2 w-full rounded-xl bg-emerald-500 py-2.5 text-sm font-semibold text-white shadow-sm transition-all hover:-translate-y-0.5 hover:bg-emerald-600 disabled:translate-y-0 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {isLoading ? (
              <span className="flex items-center justify-center gap-2">
                <Loader2 className="h-4 w-4 animate-spin" /> Signing in…
              </span>
            ) : (
              'Sign in'
            )}
          </button>
        </form>
      </div>

      <p className="mt-6 text-xs text-slate-400">Self-hosted · Open source · MIT license</p>
    </div>
  );
}
