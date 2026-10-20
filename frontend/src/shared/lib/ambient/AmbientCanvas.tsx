import React, { useEffect, useRef } from 'react';
import { AmbientEngine } from './ambientEngine';
import type { AmbientPresetName, AmbientIntensity } from './types';

export interface AmbientCanvasProps {
  themeId?: string;
  intensity?: AmbientIntensity;
  isInteractive?: boolean;
  className?: string;
}

const THEME_TO_PRESET: Record<string, AmbientPresetName> = {
  'kyoto-snow': 'snow',
  tsukimi: 'autumn',
  celestial: 'celestial',
  'frost-pine': 'blizzard',
  ocean: 'ocean',
};

export const AmbientCanvas: React.FC<AmbientCanvasProps> = ({
  themeId,
  intensity = 'subtle',
  isInteractive = true,
  className = 'fixed inset-0 pointer-events-none z-0',
}) => {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const engineRef = useRef<AmbientEngine | null>(null);

  // Initialize engine on mount
  useEffect(() => {
    const engine = new AmbientEngine();
    engineRef.current = engine;

    if (canvasRef.current) {
      engine.attach(canvasRef.current);
    }

    return () => {
      engine.detach();
      engineRef.current = null;
    };
  }, []);

  // Sync preset, intensity and interactivity
  useEffect(() => {
    const engine = engineRef.current;
    if (!engine) return;

    const presetName = themeId ? THEME_TO_PRESET[themeId] ?? null : null;
    engine.setPreset(presetName);
    engine.setIntensity(intensity);
    engine.setInteractive(isInteractive);
  }, [themeId, intensity, isInteractive]);

  // If no preset or intensity is off, we still keep canvas mounted with 0 opacity / hidden
  const hasActivePreset = Boolean(themeId && THEME_TO_PRESET[themeId] && intensity !== 'off');

  return (
    <canvas
      ref={canvasRef}
      className={className}
      style={{
        display: hasActivePreset ? 'block' : 'none',
      }}
      aria-hidden="true"
    />
  );
};

