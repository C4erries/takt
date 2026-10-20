import type {
  AmbientPreset,
  AmbientPresetName,
  AmbientIntensity,
  PresetContext,
  MouseState,
  ClickPulse,
} from './types';
import { SnowPreset, AutumnPreset, CelestialPreset, OceanPreset } from './presets';

export class AmbientEngine {
  private canvas: HTMLCanvasElement | null = null;
  private ctx: CanvasRenderingContext2D | null = null;
  private activePreset: AmbientPreset | null = null;
  private activePresetName: AmbientPresetName | null = null;
  private intensity: AmbientIntensity = 'subtle';
  private isInteractive = true;

  private rafId: number | null = null;
  private lastTime = 0;
  private width = 0;
  private height = 0;
  private dpr = 1;

  private mouse: MouseState = {
    x: -9999,
    y: -9999,
    isOver: false,
    lastMoveTime: 0,
  };
  private clicks: ClickPulse[] = [];

  private isPaused = false;
  private boundOnVisibilityChange: () => void;
  private boundOnPointerMove: (e: PointerEvent) => void;
  private boundOnPointerDown: (e: PointerEvent) => void;
  private boundOnPointerLeave: () => void;
  private boundOnResize: () => void;

  constructor() {
    this.boundOnVisibilityChange = this.handleVisibilityChange.bind(this);
    this.boundOnPointerMove = this.handlePointerMove.bind(this);
    this.boundOnPointerDown = this.handlePointerDown.bind(this);
    this.boundOnPointerLeave = this.handlePointerLeave.bind(this);
    this.boundOnResize = this.handleResize.bind(this);
  }

  attach(canvas: HTMLCanvasElement): void {
    if (this.canvas === canvas) return;
    this.detach();

    this.canvas = canvas;
    this.ctx = canvas.getContext('2d');
    if (!this.ctx) return;

    this.updateDimensions();
    this.attachEventListeners();
    this.startLoop();
  }

  detach(): void {
    this.stopLoop();
    this.detachEventListeners();
    if (this.activePreset) {
      this.activePreset.destroy();
      this.activePreset = null;
      this.activePresetName = null;
    }
    this.canvas = null;
    this.ctx = null;
  }

  setPreset(name: AmbientPresetName | null): void {
    if (this.activePresetName === name) return;

    if (this.activePreset) {
      this.activePreset.destroy();
      this.activePreset = null;
    }
    this.activePresetName = name;

    if (!name || !this.ctx || !this.canvas) return;

    switch (name) {
      case 'snow':
        this.activePreset = new SnowPreset(false);
        break;
      case 'blizzard':
        this.activePreset = new SnowPreset(true);
        break;
      case 'autumn':
        this.activePreset = new AutumnPreset();
        break;
      case 'celestial':
        this.activePreset = new CelestialPreset();
        break;
      case 'ocean':
        this.activePreset = new OceanPreset();
        break;
    }

    if (this.activePreset && this.ctx) {
      this.activePreset.init(this.ctx, this.width, this.height, this.dpr);
    }
  }

  setIntensity(intensity: AmbientIntensity): void {
    this.intensity = intensity;
    if (intensity === 'off' && this.ctx && this.canvas) {
      this.ctx.clearRect(0, 0, this.width, this.height);
    }
  }

  setInteractive(interactive: boolean): void {
    this.isInteractive = interactive;
  }

  private updateDimensions(): void {
    if (!this.canvas || typeof window === 'undefined') return;

    this.width = window.innerWidth;
    this.height = window.innerHeight;
    this.dpr = Math.min(window.devicePixelRatio || 1, 2);

    this.canvas.width = this.width * this.dpr;
    this.canvas.height = this.height * this.dpr;
    this.canvas.style.width = `${this.width}px`;
    this.canvas.style.height = `${this.height}px`;

    if (this.ctx) {
      this.ctx.scale(this.dpr, this.dpr);
    }

    if (this.activePreset) {
      this.activePreset.resize(this.width, this.height, this.dpr);
    }
  }

  private attachEventListeners(): void {
    if (typeof window === 'undefined') return;

    window.addEventListener('resize', this.boundOnResize, { passive: true });
    window.addEventListener('pointermove', this.boundOnPointerMove, { passive: true });
    window.addEventListener('pointerdown', this.boundOnPointerDown, { passive: true });
    document.addEventListener('pointerleave', this.boundOnPointerLeave, { passive: true });
    document.addEventListener('visibilitychange', this.boundOnVisibilityChange);
  }

  private detachEventListeners(): void {
    if (typeof window === 'undefined') return;

    window.removeEventListener('resize', this.boundOnResize);
    window.removeEventListener('pointermove', this.boundOnPointerMove);
    window.removeEventListener('pointerdown', this.boundOnPointerDown);
    document.removeEventListener('pointerleave', this.boundOnPointerLeave);
    document.removeEventListener('visibilitychange', this.boundOnVisibilityChange);
  }

  private handleResize(): void {
    this.updateDimensions();
  }

  private handlePointerMove(e: PointerEvent): void {
    this.mouse.x = e.clientX;
    this.mouse.y = e.clientY;
    this.mouse.isOver = true;
    this.mouse.lastMoveTime = performance.now();
  }

  private handlePointerDown(e: PointerEvent): void {
    this.clicks.push({
      x: e.clientX,
      y: e.clientY,
      time: performance.now(),
      maxRadius: 260,
    });
    if (this.clicks.length > 20) {
      this.clicks.shift();
    }
  }

  private handlePointerLeave(): void {
    this.mouse.isOver = false;
  }

  private handleVisibilityChange(): void {
    if (document.hidden) {
      this.isPaused = true;
    } else {
      this.isPaused = false;
      this.lastTime = performance.now();
    }
  }

  private startLoop(): void {
    this.lastTime = performance.now();
    const loop = (currentTime: number) => {
      this.rafId = requestAnimationFrame(loop);

      if (this.isPaused || !this.ctx || !this.canvas) return;
      if (this.intensity === 'off' || !this.activePreset) {
        this.ctx.clearRect(0, 0, this.width, this.height);
        return;
      }

      const dt = Math.min((currentTime - this.lastTime) / 1000, 0.1);
      this.lastTime = currentTime;

      const context: PresetContext = {
        ctx: this.ctx,
        width: this.width,
        height: this.height,
        dpr: this.dpr,
        intensity: this.intensity,
        isInteractive: this.isInteractive,
        mouse: this.isInteractive ? this.mouse : { x: -9999, y: -9999, isOver: false, lastMoveTime: 0 },
        clicks: this.isInteractive ? this.clicks : [],
      };

      this.ctx.clearRect(0, 0, this.width, this.height);
      this.activePreset.update(dt, context);
      this.activePreset.render(context);
    };

    this.rafId = requestAnimationFrame(loop);
  }

  private stopLoop(): void {
    if (this.rafId !== null) {
      cancelAnimationFrame(this.rafId);
      this.rafId = null;
    }
  }
}

