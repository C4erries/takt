import React from 'react';
import { Check, Sun, Moon, Sparkles, MousePointer } from 'lucide-react';
import { Modal, Badge } from '@/shared/components/ui';
import { cn } from '@/shared/lib/cn';
import { useThemeStore } from './themeStore';
import {
  LIGHT_THEMES,
  DARK_THEMES,
  ANIMATED_THEMES,
  type ThemeDefinition,
} from './themes';

export const ThemePickerModal: React.FC = () => {
  const {
    theme: currentTheme,
    ambientIntensity,
    ambientInteractive,
    isPickerOpen,
    setPickerOpen,
    setTheme,
    setAmbientIntensity,
    setAmbientInteractive,
  } = useThemeStore();

  const renderThemeCard = (t: ThemeDefinition) => {
    const isActive = t.id === currentTheme;

    return (
      <button
        key={t.id}
        type="button"
        onClick={() => setTheme(t.id)}
        className={cn(
          'group relative flex flex-col text-left p-3 rounded-xl border transition-all duration-200 select-none cursor-pointer',
          isActive
            ? 'bg-surface border-accent-solid ring-2 ring-accent-solid/30 shadow-washi-md'
            : 'bg-surface/70 hover:bg-surface border-subtle hover:border-primary/40 shadow-washi-sm hover:shadow-washi'
        )}
      >
        {/* Scoped theme preview swatch */}
        <div
          data-theme={t.id}
          className="w-full h-16 rounded-lg bg-canvas border border-subtle p-2 flex flex-col justify-between overflow-hidden mb-2.5 transition-transform duration-200 group-hover:scale-[1.01]"
        >
          <div className="flex items-center justify-between">
            {/* Mini stationery card */}
            <div className="w-16 h-7 rounded bg-surface border border-subtle px-1.5 py-1 flex items-center justify-between shadow-washi-sm">
              <span className="w-2.5 h-2.5 rounded-full bg-accent-solid shrink-0" />
              <div className="flex flex-col gap-0.5 w-8">
                <span className="h-1 rounded bg-primary/70 w-full" />
                <span className="h-0.5 rounded bg-muted/40 w-2/3" />
              </div>
            </div>
            {/* Palette swatch dots */}
            <div className="flex items-center gap-1">
              <span className="w-2.5 h-2.5 rounded-full bg-tag-persimmon shadow-sm" />
              <span className="w-2.5 h-2.5 rounded-full bg-tag-denim shadow-sm" />
              <span className="w-2.5 h-2.5 rounded-full bg-tag-straw shadow-sm" />
            </div>
          </div>
          {/* Bottom subtle baseline */}
          <div className="flex items-center justify-between text-[9px] font-mono text-muted/80">
            <span className="truncate flex items-center gap-1">
              {t.ambientPreset && <Sparkles className="w-2.5 h-2.5 text-accent-solid" />}
              {t.id}
            </span>
            <span className="w-8 h-1 rounded-full bg-accent-solid/30" />
          </div>
        </div>

        {/* Theme Name & Active Badge */}
        <div className="flex items-center justify-between gap-1.5 mb-1">
          <span className="text-xs font-semibold text-primary truncate font-serif">
            {t.name}
          </span>
          {isActive ? (
            <Badge variant="matcha" size="sm" className="gap-1 px-1.5 py-0 text-[10px] shrink-0">
              <Check className="w-3 h-3 stroke-[2.5]" />
              <span>Активна</span>
            </Badge>
          ) : (
            <span className="w-2 h-2 rounded-full border border-subtle group-hover:border-primary/50 transition-colors" />
          )}
        </div>

        {/* Description */}
        <p className="text-[11px] text-muted leading-snug line-clamp-2">
          {t.description}
        </p>
      </button>
    );
  };

  return (
    <Modal
      isOpen={isPickerOpen}
      onClose={() => setPickerOpen(false)}
      title="Палитра оформления"
      description="16 выверенных тем в стиле японской канцелярии и Editorial Tech"
      className="max-w-3xl"
    >
      <div className="space-y-6">
        {/* Ambient Intensity Selector */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-3 rounded-xl bg-canvas/60 border border-subtle">
          <div className="space-y-0.5">
            <div className="flex items-center gap-1.5 text-xs font-medium text-primary">
              <Sparkles className="w-3.5 h-3.5 text-accent-solid" />
              <span>Интенсивность фоновой анимации</span>
            </div>
            <p className="text-[11px] text-muted">
              Управление снегопадом, осенним листопадом и интерактивным звездным атласом
            </p>
          </div>
          <div className="flex items-center gap-1 p-0.5 rounded-lg bg-surface border border-subtle self-start sm:self-auto">
            {(
              [
                { id: 'off', label: 'Выкл' },
                { id: 'subtle', label: 'Деликатная' },
                { id: 'vivid', label: 'Насыщенная' },
              ] as const
            ).map(({ id, label }) => {
              const isSelected = ambientIntensity === id;
              return (
                <button
                  key={id}
                  type="button"
                  onClick={() => setAmbientIntensity(id)}
                  className={cn(
                    'px-2.5 py-1 text-xs rounded-md transition-all duration-150 font-medium select-none',
                    isSelected
                      ? 'bg-accent-solid text-surface shadow-washi-sm'
                      : 'text-muted hover:text-primary hover:bg-canvas/50'
                  )}
                >
                  {label}
                </button>
              );
            })}
          </div>
        </div>

        {/* Cursor Interactivity Toggle */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-3 rounded-xl bg-canvas/60 border border-subtle">
          <div className="space-y-0.5">
            <div className="flex items-center gap-1.5 text-xs font-medium text-primary">
              <MousePointer className="w-3.5 h-3.5 text-accent-solid" />
              <span>Интерактивность с курсором</span>
            </div>
            <p className="text-[11px] text-muted">
              Реакция частиц, листьев и созвездий на движение и клики мыши
            </p>
          </div>
          <div className="flex items-center gap-1 p-0.5 rounded-lg bg-surface border border-subtle self-start sm:self-auto">
            <button
              type="button"
              onClick={() => setAmbientInteractive(true)}
              className={cn(
                'px-2.5 py-1 text-xs rounded-md transition-all duration-150 font-medium select-none',
                ambientInteractive
                  ? 'bg-accent-solid text-surface shadow-washi-sm'
                  : 'text-muted hover:text-primary hover:bg-canvas/50'
              )}
            >
              Вкл
            </button>
            <button
              type="button"
              onClick={() => setAmbientInteractive(false)}
              className={cn(
                'px-2.5 py-1 text-xs rounded-md transition-all duration-150 font-medium select-none',
                !ambientInteractive
                  ? 'bg-accent-solid text-surface shadow-washi-sm'
                  : 'text-muted hover:text-primary hover:bg-canvas/50'
              )}
            >
              Выкл
            </button>
          </div>
        </div>

        {/* Section 1: Animated & Seasonal Themes */}
        <div className="space-y-3">
          <div className="flex items-center gap-2 pb-1 border-b border-subtle/50">
            <Sparkles className="w-4 h-4 text-accent-solid" />
            <span className="text-xs font-medium uppercase tracking-wider text-muted font-mono">
              Анимированные и сезонные темы ({ANIMATED_THEMES.length})
            </span>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3">
            {ANIMATED_THEMES.map(renderThemeCard)}
          </div>
        </div>

        {/* Section 2: Light Static Themes */}
        <div className="space-y-3">
          <div className="flex items-center gap-2 pb-1 border-b border-subtle/50">
            <Sun className="w-4 h-4 text-accent-straw" />
            <span className="text-xs font-medium uppercase tracking-wider text-muted font-mono">
              Светлая бумага ({LIGHT_THEMES.length})
            </span>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3">
            {LIGHT_THEMES.map(renderThemeCard)}
          </div>
        </div>

        {/* Section 3: Dark Static Themes */}
        <div className="space-y-3">
          <div className="flex items-center gap-2 pb-1 border-b border-subtle/50">
            <Moon className="w-4 h-4 text-accent-solid" />
            <span className="text-xs font-medium uppercase tracking-wider text-muted font-mono">
              Вечерние чернила ({DARK_THEMES.length})
            </span>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3">
            {DARK_THEMES.map(renderThemeCard)}
          </div>
        </div>
      </div>
    </Modal>
  );
};

