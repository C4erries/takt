import React, { useState } from 'react';
import { Modal, Input, Button, ErrorBanner } from '@/shared/components/ui';
import type { LoginRequest, RegisterRequest } from '@/shared/api';

export interface AuthModalProps {
  isOpen: boolean;
  onClose: () => void;
  onLogin: (data: LoginRequest) => Promise<void>;
  onRegister: (data: RegisterRequest) => Promise<void>;
  defaultMode?: 'login' | 'register';
}

export const AuthModal: React.FC<AuthModalProps> = ({
  isOpen,
  onClose,
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
        await onRegister({ email, password, displayName: displayName || email.split('@')[0] });
      }
      onClose();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Произошла непредвиденная ошибка');
    } finally {
      setIsSubmitting(false);
    }
  };

  const switchMode = (newMode: 'login' | 'register') => {
    setMode(newMode);
    setError(null);
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={mode === 'login' ? 'Вход в Takt' : 'Создание аккаунта'}
      description={
        mode === 'login'
          ? 'Введите ваши учетные данные для доступа к личному пространству'
          : 'Создайте личное рабочее пространство с тактильной организацией задач'
      }
    >
      <div className="flex border-b border-subtle/70 pb-3 mb-2 gap-2">
        <Button
          type="button"
          size="sm"
          variant={mode === 'login' ? 'solid' : 'ghost'}
          onClick={() => switchMode('login')}
          className="flex-1"
        >
          Вход
        </Button>
        <Button
          type="button"
          size="sm"
          variant={mode === 'register' ? 'solid' : 'ghost'}
          onClick={() => switchMode('register')}
          className="flex-1"
        >
          Регистрация
        </Button>
      </div>

      {error && (
        <ErrorBanner
          title="Ошибка входа"
          message={error}
          className="my-1"
        />
      )}

      <form onSubmit={handleSubmit} className="space-y-3.5">
        {mode === 'register' && (
          <div>
            <label className="block text-xs uppercase tracking-wider text-muted font-mono mb-1">
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
          <label className="block text-xs uppercase tracking-wider text-muted font-mono mb-1">
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
          <label className="block text-xs uppercase tracking-wider text-muted font-mono mb-1">
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

        <div className="pt-2 flex items-center justify-end gap-2">
          <Button type="button" variant="ghost" onClick={onClose} disabled={isSubmitting}>
            Отмена
          </Button>
          <Button type="submit" variant="solid" isLoading={isSubmitting}>
            {mode === 'login' ? 'Войти в систему' : 'Зарегистрироваться'}
          </Button>
        </div>
      </form>
    </Modal>
  );
};
