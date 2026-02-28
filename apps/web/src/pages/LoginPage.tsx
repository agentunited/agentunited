import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Link, useNavigate } from 'react-router-dom';
import { Header } from '../components/layout/Header';
import { Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';

const loginSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(6, 'Password must be at least 6 characters'),
});

type LoginFormData = z.infer<typeof loginSchema>;

export function LoginPage() {
  const [isLoading, setIsLoading] = useState(false);
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
      // Real API call
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
      
      // Navigate to chat
      navigate('/chat');
    } catch (err) {
      setError('Invalid email or password');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <Header />
      <main className="container mx-auto px-4 py-12">
        <div className="max-w-md mx-auto">
          <Card>
            <h1 className="text-2xl font-bold text-gray-900 mb-6">Log In</h1>
            
            {error && (
              <div className="bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded-lg mb-4" role="alert">
                {error}
              </div>
            )}

            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              <Input
                label="Email"
                type="email"
                autoComplete="email"
                error={errors.email?.message}
                {...register('email')}
              />

              <Input
                label="Password"
                type="password"
                autoComplete="current-password"
                error={errors.password?.message}
                {...register('password')}
              />

              <Button
                type="submit"
                variant="primary"
                className="w-full"
                isLoading={isLoading}
              >
                Log In
              </Button>
            </form>

            <p className="mt-4 text-center text-sm text-gray-600">
              Don't have an account?{' '}
              <Link to="/signup" className="text-blue-600 hover:text-blue-700 font-medium">
                Sign Up
              </Link>
            </p>
          </Card>
        </div>
      </main>
    </div>
  );
}
