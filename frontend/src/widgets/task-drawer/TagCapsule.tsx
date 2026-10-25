import React, { useRef, useEffect } from 'react';
import { X, Check } from 'lucide-react';
import { cn } from '@/shared/lib/cn';
import { TAG_PALETTE, DEFAULT_TAG_COLOR } from '@/entities/task';
import type { Tag } from '@/shared/api';

export interface TagCapsuleProps {
  tag: Tag;
  isOpen: boolean;
  onToggleOpen: () => void;
  onClose: () => void;
  onUpdateColor: (hex: string) => void;
  onRemove: () => void;
}

export const TagCapsule: React.FC<TagCapsuleProps> = ({
  tag,
  isOpen,
  onToggleOpen,
  onClose,
  onUpdateColor,
  onRemove,
}) => {
  const popoverRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (!isOpen) return;
    const handleClick = (e: MouseEvent) => {
      if (popoverRef.current && !popoverRef.current.contains(e.target as Node)) {
        onClose();
      }
    };
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, [isOpen, onClose]);

  return (
    <div className="relative inline-block">
      <span className="inline-flex items-center gap-1.5 bg-accent-surface/80 text-primary text-[11px] font-medium px-2.5 py-0.5 rounded-full border border-subtle transition-colors shadow-washi-sm">
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation();
            onToggleOpen();
          }}
          className="w-2.5 h-2.5 rounded-full shrink-0 shadow-xs cursor-pointer hover:scale-125 transition-transform"
          style={{ backgroundColor: tag.colorHex || DEFAULT_TAG_COLOR }}
          title="Изменить цвет метки"
          aria-label={`Изменить цвет метки ${tag.name}`}
        />
        <span
          onClick={(e) => {
            e.stopPropagation();
            onToggleOpen();
          }}
          className="cursor-pointer hover:text-accent-solid transition-colors"
          title="Изменить цвет метки"
        >
          {tag.name}
        </span>
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation();
            onRemove();
          }}
          className="text-muted hover:text-primary transition-colors ml-0.5 rounded-full p-0.5 hover:bg-canvas/50"
          title="Удалить тег"
        >
          <X className="w-2.5 h-2.5" />
        </button>
      </span>

      {/* Color Palette Popover (12 Colors) */}
      {isOpen && (
        <div
          ref={popoverRef}
          className="absolute left-0 top-full mt-1.5 z-50 bg-surface border border-subtle rounded-xl shadow-washi-modal p-2 w-48 text-xs space-y-1.5 animate-in fade-in zoom-in-95 duration-100"
          onClick={(e) => e.stopPropagation()}
        >
          <div className="text-[10px] font-mono uppercase tracking-wider text-muted px-1">
            Цвет «{tag.name}»
          </div>
          <div className="grid grid-cols-6 gap-1.5 p-1">
            {TAG_PALETTE.map((c) => {
              const isSelected = (tag.colorHex || '').toLowerCase() === c.hex.toLowerCase();
              return (
                <button
                  key={c.hex}
                  type="button"
                  title={c.name}
                  onClick={() => {
                    onUpdateColor(c.hex);
                    onClose();
                  }}
                  style={{ backgroundColor: c.hex }}
                  className={cn(
                    'w-5 h-5 rounded-full flex items-center justify-center transition-all hover:scale-115 cursor-pointer',
                    isSelected
                      ? 'ring-2 ring-primary ring-offset-2 ring-offset-surface scale-105'
                      : 'opacity-85 hover:opacity-100'
                  )}
                >
                  {isSelected && <Check className="w-3 h-3 text-white drop-shadow-xs" />}
                </button>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
};
