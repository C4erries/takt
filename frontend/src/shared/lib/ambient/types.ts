// Takt Canvas 2D Ambient Engine — Types (ADR-0007, Sprint 0.7.4)

export type AmbientPresetName = 'snow' | 'autumn' | 'celestial' | 'blizzard';

export type AmbientIntensity = 'off' | 'subtle' | 'vivid';

export interface MouseState {
  x: number;
  y: number;
  isOver: boolean;
  lastMoveTime: number;
}

export interface ClickPulse {
  x: number;
  y: number;
  time: number;
  maxRadius: number;
}

export interface PresetContext {
  ctx: CanvasRenderingContext2D;
  width: number;
  height: number;
  dpr: number;
  intensity: AmbientIntensity;
  mouse: MouseState;
  clicks: ClickPulse[];
}

export interface AmbientPreset {
  init(ctx: CanvasRenderingContext2D, width: number, height: number, dpr: number): void;
  resize(width: number, height: number, dpr: number): void;
  update(dt: number, context: PresetContext): void;
  render(context: PresetContext): void;
  destroy(): void;
}

