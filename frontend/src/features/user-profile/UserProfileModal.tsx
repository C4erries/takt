import React, { useState, useRef, useEffect } from 'react';
import { Camera, Upload, AlertCircle, Check, Loader2 } from 'lucide-react';
import { Modal, Input, Button } from '@/shared/components/ui';
import { useUpdateProfileMutation, useUploadAvatarMutation } from '@/entities/user';
import type { User } from '@/shared/api';

export interface UserProfileModalProps {
  isOpen: boolean;
  onClose: () => void;
  user: User | null;
  onUserUpdated?: (user: User) => void;
}

const MAX_AVATAR_SIZE = 5 * 1024 * 1024; // 5MB
const ALLOWED_TYPES = ['image/png', 'image/jpeg', 'image/jpg', 'image/webp'];

export const UserProfileModal: React.FC<UserProfileModalProps> = ({
  isOpen,
  onClose,
  user,
  onUserUpdated,
}) => {
  const [displayName, setDisplayName] = useState(user?.displayName || '');
  const [avatarPreview, setAvatarPreview] = useState<string | null>(user?.avatarUrl || null);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [fileError, setFileError] = useState<string | null>(null);
  const [isSuccess, setIsSuccess] = useState(false);
  const [isDragging, setIsDragging] = useState(false);

  const fileInputRef = useRef<HTMLInputElement | null>(null);

  const updateProfileMutation = useUpdateProfileMutation();
  const uploadAvatarMutation = useUploadAvatarMutation();

  useEffect(() => {
    if (isOpen && user) {
      setDisplayName(user.displayName || '');
      setAvatarPreview(user.avatarUrl || null);
      setSelectedFile(null);
      setFileError(null);
      setIsSuccess(false);
    }
  }, [isOpen, user]);

  const handleFileValidation = (file: File): boolean => {
    setFileError(null);
    if (!ALLOWED_TYPES.includes(file.type)) {
      setFileError('Поддерживаются только форматы PNG, JPEG и WebP');
      return false;
    }
    if (file.size > MAX_AVATAR_SIZE) {
      setFileError('Размер файла не должен превышать 5 МБ');
      return false;
    }
    return true;
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    if (handleFileValidation(file)) {
      setSelectedFile(file);
      setAvatarPreview(URL.createObjectURL(file));
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    const file = e.dataTransfer.files?.[0];
    if (!file) return;
    if (handleFileValidation(file)) {
      setSelectedFile(file);
      setAvatarPreview(URL.createObjectURL(file));
    }
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!user) return;
    setFileError(null);
    setIsSuccess(false);

    let updatedUser = user;

    try {
      // 1. Upload avatar if selected
      if (selectedFile) {
        updatedUser = await uploadAvatarMutation.mutateAsync(selectedFile);
      }

      // 2. Update display name if changed
      const trimmedName = displayName.trim();
      if (trimmedName !== (user.displayName || '')) {
        updatedUser = await updateProfileMutation.mutateAsync({
          displayName: trimmedName || undefined,
        });
      }

      onUserUpdated?.(updatedUser);
      setIsSuccess(true);
      setTimeout(() => {
        onClose();
      }, 700);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Не удалось обновить профиль';
      setFileError(msg);
    }
  };

  const isPending = updateProfileMutation.isPending || uploadAvatarMutation.isPending;
  const initialLetter = (displayName.trim() || user?.email || 'U').charAt(0).toUpperCase();

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Профиль пользователя"
      description="Настройка отображаемого имени и персонального аватара"
      className="max-w-md"
    >
      <form onSubmit={handleSave} className="space-y-5">
        {/* Avatar Upload with Drag and Drop */}
        <div className="flex flex-col items-center gap-3">
          <div
            onDragOver={(e) => {
              e.preventDefault();
              setIsDragging(true);
            }}
            onDragLeave={() => setIsDragging(false)}
            onDrop={handleDrop}
            onClick={() => fileInputRef.current?.click()}
            className={`relative group w-24 h-24 rounded-full border-2 border-dashed flex items-center justify-center cursor-pointer overflow-hidden transition-all shadow-washi-sm ${
              isDragging
                ? 'border-accent-solid bg-accent-surface/40 scale-105'
                : 'border-subtle hover:border-accent-solid/70 bg-canvas/60'
            }`}
            title="Нажмите или перетащите изображение"
          >
            {avatarPreview ? (
              <img
                src={avatarPreview}
                alt="Аватар"
                className="w-full h-full object-cover rounded-full"
              />
            ) : (
              <span className="text-3xl font-serif font-medium text-primary">
                {initialLetter}
              </span>
            )}

            {/* Hover Camera Overlay */}
            <div className="absolute inset-0 bg-primary/40 text-surface flex flex-col items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity rounded-full">
              <Camera className="w-6 h-6 mb-0.5" />
              <span className="text-[9px] font-sans">Сменить</span>
            </div>
          </div>

          <input
            ref={fileInputRef}
            type="file"
            accept="image/png,image/jpeg,image/webp"
            onChange={handleFileChange}
            className="hidden"
          />

          <div className="text-center">
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              className="text-xs text-accent-solid hover:underline inline-flex items-center gap-1 font-medium"
            >
              <Upload className="w-3.5 h-3.5" />
              <span>Загрузить фото</span>
            </button>
            <p className="text-[10px] text-muted mt-0.5">
              PNG, JPEG или WebP до 5 МБ
            </p>
          </div>
        </div>

        {/* Display Name Input */}
        <div className="space-y-1.5">
          <label className="text-[10px] font-mono uppercase tracking-wider text-muted">
            Отображаемое имя
          </label>
          <Input
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            placeholder="Ваше имя..."
            className="text-xs bg-surface border-subtle h-9"
          />
        </div>

        {/* Email (Readonly) */}
        <div className="space-y-1.5">
          <label className="text-[10px] font-mono uppercase tracking-wider text-muted">
            Email аккаунта
          </label>
          <div className="text-xs text-muted font-mono bg-canvas/60 px-3 py-2 rounded-lg border border-subtle/50 select-text">
            {user?.email}
          </div>
        </div>

        {/* Error message */}
        {fileError && (
          <div className="flex items-center gap-2 text-xs text-destructive bg-destructive/10 border border-destructive/20 p-2.5 rounded-lg">
            <AlertCircle className="w-4 h-4 shrink-0" />
            <span>{fileError}</span>
          </div>
        )}

        {/* Success message */}
        {isSuccess && (
          <div className="flex items-center gap-2 text-xs text-matcha bg-matcha/10 border border-matcha/20 p-2.5 rounded-lg animate-in fade-in">
            <Check className="w-4 h-4 shrink-0" />
            <span>Профиль успешно обновлен!</span>
          </div>
        )}

        {/* Footer Actions */}
        <div className="flex items-center justify-end gap-2 pt-2 border-t border-subtle/60">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={onClose}
            disabled={isPending}
            className="text-xs"
          >
            Отмена
          </Button>
          <Button
            type="submit"
            variant="solid"
            size="sm"
            disabled={isPending || isSuccess}
            className="text-xs px-4"
          >
            {isPending ? (
              <span className="flex items-center gap-1.5">
                <Loader2 className="w-3.5 h-3.5 animate-spin" />
                Сохранение...
              </span>
            ) : isSuccess ? (
              'Сохранено'
            ) : (
              'Сохранить'
            )}
          </Button>
        </div>
      </form>
    </Modal>
  );
};
