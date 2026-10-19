import type { AmbientPreset, PresetContext, AmbientIntensity } from './types';

interface Snowflake {
  x: number;
  y: number;
  radius: number;
  speedX: number;
  speedY: number;
  swaySpeed: number;
  swayAmp: number;
  swayPhase: number;
  alpha: number;
  layer: 0 | 1 | 2; // 0: far, 1: mid, 2: near
  isCrystal: boolean;
  crystalAngle: number;
  crystalRotSpeed: number;
  crystalSize: number;
}

export class SnowPreset implements AmbientPreset {
  private flakes: Snowflake[] = [];
  private width = 0;
  private height = 0;
  private dpr = 1;
  private currentIntensity: AmbientIntensity = 'subtle';
  private bokehCanvases: HTMLCanvasElement[] = [];
  private windPhase = 0;

  constructor(private isBlizzard = false) {}

  init(_ctx: CanvasRenderingContext2D, width: number, height: number, dpr: number): void {
    this.width = width;
    this.height = height;
    this.dpr = dpr;
    this.initBokehCanvases();
    this.rebuildFlakes();
  }

  resize(width: number, height: number, dpr: number): void {
    this.width = width;
    this.height = height;
    this.dpr = dpr;
  }

  private initBokehCanvases(): void {
    if (typeof document === 'undefined') return;
    this.bokehCanvases = [2.5, 4.5, 8.5].map((baseRadius) => {
      const canvas = document.createElement('canvas');
      const size = (baseRadius * 4 + 8) * this.dpr;
      canvas.width = size;
      canvas.height = size;
      const bCtx = canvas.getContext('2d');
      if (bCtx) {
        const center = size / 2;
        const grad = bCtx.createRadialGradient(
          center,
          center,
          0,
          center,
          center,
          baseRadius * this.dpr * 1.8
        );
        grad.addColorStop(0, 'rgba(255, 255, 255, 1)');
        grad.addColorStop(0.4, 'rgba(242, 248, 255, 0.95)');
        grad.addColorStop(0.75, 'rgba(175, 205, 235, 0.45)');
        grad.addColorStop(1, 'rgba(140, 175, 210, 0)');
        bCtx.fillStyle = grad;
        bCtx.beginPath();
        bCtx.arc(center, center, baseRadius * this.dpr * 1.8, 0, Math.PI * 2);
        bCtx.fill();
      }
      return canvas;
    });
  }

  private getTargetCount(intensity: AmbientIntensity): number {
    if (this.isBlizzard) {
      if (intensity === 'vivid') return 160;
      if (intensity === 'subtle') return 75;
      return 0;
    }
    if (intensity === 'vivid') return 110;
    if (intensity === 'subtle') return 45;
    return 0;
  }

  private rebuildFlakes(): void {
    const count = this.getTargetCount(this.currentIntensity);
    this.flakes = [];
    if (count === 0 || this.width === 0 || this.height === 0) return;

    for (let i = 0; i < count; i++) {
      const rand = Math.random();
      const layer: 0 | 1 | 2 = rand < 0.45 ? 0 : rand < 0.8 ? 1 : 2;
      const isCrystal = rand > 0.65;

      let radius = 2;
      let speedX = this.isBlizzard ? 160 + Math.random() * 80 : 0;
      let speedY = this.isBlizzard ? 70 + Math.random() * 50 : 20;
      let alpha = 0.65;
      let crystalSize = 5;

      if (layer === 1) {
        radius = 3.5;
        speedX = this.isBlizzard ? 220 + Math.random() * 100 : 0;
        speedY = this.isBlizzard ? 95 + Math.random() * 60 : 34;
        alpha = 0.85;
        crystalSize = 8;
      } else if (layer === 2) {
        radius = 6;
        speedX = this.isBlizzard ? 300 + Math.random() * 120 : 0;
        speedY = this.isBlizzard ? 125 + Math.random() * 70 : 52;
        alpha = 0.98;
        crystalSize = 12;
      }

      this.flakes.push({
        x: Math.random() * this.width,
        y: Math.random() * this.height,
        radius,
        speedX,
        speedY: speedY * (0.85 + Math.random() * 0.3),
        swaySpeed: 0.8 + Math.random() * 1.2,
        swayAmp: 16 + Math.random() * 26,
        swayPhase: Math.random() * Math.PI * 2,
        alpha,
        layer,
        isCrystal,
        crystalAngle: Math.random() * Math.PI * 2,
        crystalRotSpeed: (Math.random() - 0.5) * 2.5,
        crystalSize,
      });
    }
  }

  update(dt: number, context: PresetContext): void {
    if (context.intensity !== this.currentIntensity) {
      this.currentIntensity = context.intensity;
      this.rebuildFlakes();
    }

    if (this.flakes.length === 0) return;
    const clampedDt = Math.min(dt, 0.1);

    this.windPhase += clampedDt * 1.8;
    const gust = this.isBlizzard ? Math.sin(this.windPhase) * 55 : 0;

    for (let i = 0; i < this.flakes.length; i++) {
      const f = this.flakes[i];
      f.y += f.speedY * clampedDt;

      if (this.isBlizzard) {
        f.x += (f.speedX + gust) * clampedDt;
        if (f.x > this.width + 40) {
          f.x = -40;
          f.y = Math.random() * this.height;
        }
      } else {
        f.swayPhase += f.swaySpeed * clampedDt;
      }

      if (f.isCrystal) {
        f.crystalAngle += f.crystalRotSpeed * clampedDt;
      }

      if (f.y > this.height + 25) {
        f.y = -25;
        f.x = Math.random() * this.width;
      }
    }
  }

  render(context: PresetContext): void {
    if (this.flakes.length === 0 || this.bokehCanvases.length === 0) return;
    const { ctx } = context;

    for (let i = 0; i < this.flakes.length; i++) {
      const f = this.flakes[i];
      const posX = this.isBlizzard ? f.x : f.x + Math.sin(f.swayPhase) * f.swayAmp;

      if (f.isCrystal) {
        ctx.save();
        ctx.globalAlpha = f.alpha;
        ctx.translate(posX, f.y);
        ctx.rotate(f.crystalAngle);

        ctx.shadowColor = 'rgba(18, 40, 70, 0.35)';
        ctx.shadowBlur = 4;
        ctx.strokeStyle = '#FFFFFF';
        ctx.lineWidth = f.layer === 2 ? 1.3 : 0.9;
        ctx.lineCap = 'round';

        const arm = f.crystalSize;
        const branch = arm * 0.35;

        for (let a = 0; a < 3; a++) {
          const angle = (a * Math.PI) / 3;
          const cos = Math.cos(angle);
          const sin = Math.sin(angle);

          ctx.beginPath();
          ctx.moveTo(-arm * cos, -arm * sin);
          ctx.lineTo(arm * cos, arm * sin);
          ctx.stroke();

          if (f.layer > 0) {
            const bDist = arm * 0.65;
            const bAngle1 = angle + Math.PI / 4;
            const bAngle2 = angle - Math.PI / 4;

            ctx.beginPath();
            ctx.moveTo(bDist * cos, bDist * sin);
            ctx.lineTo(bDist * cos + branch * Math.cos(bAngle1), bDist * sin + branch * Math.sin(bAngle1));
            ctx.moveTo(bDist * cos, bDist * sin);
            ctx.lineTo(bDist * cos + branch * Math.cos(bAngle2), bDist * sin + branch * Math.sin(bAngle2));

            ctx.moveTo(-bDist * cos, -bDist * sin);
            ctx.lineTo(-bDist * cos - branch * Math.cos(bAngle1), -bDist * sin - branch * Math.sin(bAngle1));
            ctx.moveTo(-bDist * cos, -bDist * sin);
            ctx.lineTo(-bDist * cos - branch * Math.cos(bAngle2), -bDist * sin - branch * Math.sin(bAngle2));
            ctx.stroke();
          }
        }
        ctx.restore();
      } else {
        const stamp = this.bokehCanvases[f.layer];
        if (!stamp) continue;

        const halfSize = stamp.width / this.dpr / 2;
        ctx.save();
        ctx.globalAlpha = f.alpha;
        ctx.shadowColor = 'rgba(20, 45, 75, 0.28)';
        ctx.shadowBlur = 4;
        ctx.drawImage(
          stamp,
          posX - halfSize,
          f.y - halfSize,
          stamp.width / this.dpr,
          stamp.height / this.dpr
        );
        ctx.restore();
      }
    }
  }

  destroy(): void {
    this.flakes = [];
    this.bokehCanvases = [];
  }
}
