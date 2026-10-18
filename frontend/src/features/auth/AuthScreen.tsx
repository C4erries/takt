import React, { useState } from 'react';
import { Input, Button, ErrorBanner } from '@/shared/components/ui';
import type { LoginRequest, RegisterRequest } from '@/shared/api';

export interface AuthScreenProps {
  onLogin: (data: LoginRequest) => Promise<void>;
  onRegister: (data: RegisterRequest) => Promise<void>;
  defaultMode?: 'login' | 'register';
}

export const AuthScreen: React.FC<AuthScreenProps> = ({
  onLogin,
  onRegister,
  defaultMode = 'login',
}) => {
  const [mode, setMode] = useState<'login' | 'register'>(defaultMode);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [displayName, setDisplayName] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsSubmitting(true);

    try {
      if (mode === 'login') {
        await onLogin({ email, password });
      } else {
        await onRegister({
          email,
          password,
          displayName: displayName || email.split('@')[0],
        });
      }
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Произошла ошибка авторизации');
    } finally {
      setIsSubmitting(false);
    }
  };

  const switchMode = (newMode: 'login' | 'register') => {
    setMode(newMode);
    setError(null);
  };

  return (
    <div className="h-screen w-screen overflow-hidden bg-canvas flex flex-col items-center justify-center p-4">
      {/* Editorial Card */}
      <div className="w-full max-w-sm bg-surface border border-subtle rounded-sheet shadow-washi p-8 space-y-6">
        {/* Brand Header */}
        <div className="text-center space-y-1.5">
          <div className="flex items-center justify-center gap-2">
            <span className="w-2.5 h-2.5 rounded-full bg-accent-solid inline-block" />
            <h1 className="font-serif text-3xl font-normal tracking-tight text-primary">
              Takt
            </h1>
          </div>
          <p className="text-xs text-muted font-sans">
            Тактильный таск-трекер и пространство фокуса
          </p>
        </div>

        {/* Mode Tabs */}
        <div className="flex border-b border-subtle/70 pb-2 gap-2">
          <Button
            type="button"
            size="sm"
            variant={mode === 'login' ? 'solid' : 'ghost'}
            onClick={() => switchMode('login')}
            className="flex-1 text-xs"
          >
            Вход
          </Button>
          <Button
            type="button"
            size="sm"
            variant={mode === 'register' ? 'solid' : 'ghost'}
            onClick={() => switchMode('register')}
            className="flex-1 text-xs"
          >
            Регистрация
          </Button>
        </div>

        {/* Error Alert */}
        {error && (
          <ErrorBanner
            title={mode === 'login' ? 'Ошибка входа' : 'Ошибка регистрации'}
            message={error}
            className="my-1"
          />
        )}

        {/* Credentials Form */}
        <form onSubmit={handleSubmit} className="space-y-4">
          {mode === 'register' && (
            <div>
              <label className="block text-[11px] uppercase tracking-wider text-muted font-mono mb-1">
                Ваше имя
              </label>
              <Input
                type="text"
                placeholder="Как к вам обращаться?"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                required
              />
            </div>
          )}

          <div>
            <label className="block text-[11px] uppercase tracking-wider text-muted font-mono mb-1">
              Электронная почта
            </label>
            <Input
              type="email"
              placeholder="you@domain.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              autoFocus
            />
          </div>

          <div>
            <label className="block text-[11px] uppercase tracking-wider text-muted font-mono mb-1">
              Пароль
            </label>
            <Input
              type="password"
              placeholder="••••••••"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={6}
            />
          </div>

          <div className="pt-2">
            <Button
              type="submit"
              variant="solid"
              isLoading={isSubmitting}
              className="w-full justify-center py-2 text-sm"
            >
              {mode === 'login' ? 'Войти в Takt' : 'Создать аккаунт'}
            </Button>
          </div>
        </form>
      </div>

      {/* Footer minimal info */}
      <div className="mt-6 text-center text-[11px] text-muted font-mono">
        Editorial GTD · Палитра Washi & Matcha
      </div>
    </div>
  );
};
