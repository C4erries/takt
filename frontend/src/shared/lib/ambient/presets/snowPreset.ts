import type { AmbientPreset, PresetContext, AmbientIntensity } from '../types';

interface Snowflake {
  x: number;
  y: number;
  radius: number;
  speedX: number;
  speedY: number;
  swaySpeed: number;
  swayAmp: number;
  swayPhase: number;
  swirlPhase: number;
  swirlSpeed: number;
  swirlAmp: number;
  alpha: number;
  layer: 0 | 1 | 2;
  isCrystal: boolean;
  crystalAngle: number;
  crystalRotSpeed: number;
  crystalSize: number;
}

interface PineNeedle {
  x: number;
  y: number;
  length: number;
  speedX: number;
  speedY: number;
  angle: number;
  rotSpeed: number;
  swirlPhase: number;
  alpha: number;
}

export class SnowPreset implements AmbientPreset {
  private flakes: Snowflake[] = [];
  private needles: PineNeedle[] = [];
  private width = 0;
  private height = 0;
  private dpr = 1;
  private currentIntensity: AmbientIntensity = 'subtle';
  private bokehCanvases: HTMLCanvasElement[] = [];
  private windPhase = 0;
  private windTime = 0;

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
        const grad = bCtx.createRadialGradient(center, center, 0, center, center, baseRadius * this.dpr * 1.8);
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

  private getTargetCount(intensity: AmbientIntensity): { flakes: number; needles: number } {
    if (this.isBlizzard) {
      if (intensity === 'vivid') return { flakes: 150, needles: 16 };
      if (intensity === 'subtle') return { flakes: 75, needles: 10 };
      return { flakes: 0, needles: 0 };
    }
    if (intensity === 'vivid') return { flakes: 110, needles: 0 };
    if (intensity === 'subtle') return { flakes: 45, needles: 0 };
    return { flakes: 0, needles: 0 };
  }

  private rebuildFlakes(): void {
    const { flakes: count, needles: needleCount } = this.getTargetCount(this.currentIntensity);
    this.flakes = [];
    this.needles = [];
    if (count === 0 || this.width === 0 || this.height === 0) return;

    for (let i = 0; i < count; i++) {
      const rand = Math.random();
      const layer: 0 | 1 | 2 = rand < 0.45 ? 0 : rand < 0.8 ? 1 : 2;
      const isCrystal = rand > 0.65;
      const layerStats = [
        { radius: 2, spdX: 160, spdY: 60, a: 0.65, sz: 5 },
        { radius: 3.5, spdX: 220, spdY: 85, a: 0.85, sz: 8 },
        { radius: 6, spdX: 290, spdY: 115, a: 0.98, sz: 12 },
      ][layer];

      this.flakes.push({
        x: Math.random() * this.width,
        y: Math.random() * this.height,
        radius: layerStats.radius,
        speedX: this.isBlizzard ? layerStats.spdX + Math.random() * 80 : 0,
        speedY: (this.isBlizzard ? layerStats.spdY : 18 + layer * 16) * (0.85 + Math.random() * 0.3),
        swaySpeed: 0.8 + Math.random() * 1.2,
        swayAmp: 16 + Math.random() * 26,
        swayPhase: Math.random() * Math.PI * 2,
        swirlPhase: Math.random() * Math.PI * 2,
        swirlSpeed: 1.8 + Math.random() * 2.4,
        swirlAmp: this.isBlizzard ? 12 + Math.random() * 22 : 0,
        alpha: layerStats.a,
        layer,
        isCrystal,
        crystalAngle: Math.random() * Math.PI * 2,
        crystalRotSpeed: (Math.random() - 0.5) * (this.isBlizzard ? 5.5 : 2.5),
        crystalSize: layerStats.sz,
      });
    }

    for (let i = 0; i < needleCount; i++) {
      this.needles.push({
        x: Math.random() * this.width,
        y: Math.random() * this.height,
        length: 8 + Math.random() * 7,
        speedX: 200 + Math.random() * 120,
        speedY: 70 + Math.random() * 60,
        angle: Math.random() * Math.PI * 2,
        rotSpeed: (Math.random() - 0.5) * 6,
        swirlPhase: Math.random() * Math.PI * 2,
        alpha: 0.7 + Math.random() * 0.25,
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
    this.windTime += clampedDt;
    this.windPhase += clampedDt * 2.2;

    const gust = this.isBlizzard ? Math.sin(this.windPhase) * Math.cos(this.windTime * 0.7) * 95 : 0;
    const { mouse, isInteractive } = context;

    for (let i = 0; i < this.flakes.length; i++) {
      const f = this.flakes[i];
      f.y += f.speedY * clampedDt;

      if (this.isBlizzard) {
        f.swirlPhase += f.swirlSpeed * clampedDt;
        f.x += (f.speedX + gust) * clampedDt;

        if (isInteractive && mouse.isOver) {
          const dx = f.x - mouse.x, dy = f.y - mouse.y;
          const dist = Math.hypot(dx, dy);
          if (dist < 120) {
            const force = (1 - dist / 120) * 110;
            f.x += (dx / (dist || 1)) * force * clampedDt;
            f.y += (dy / (dist || 1)) * force * clampedDt;
          }
        }
        if (f.x > this.width + 40) {
          f.x = -40;
          f.y = Math.random() * this.height;
        }
      } else {
        f.swayPhase += f.swaySpeed * clampedDt;
      }

      if (f.isCrystal) f.crystalAngle += f.crystalRotSpeed * clampedDt;
      if (f.y > this.height + 25) {
        f.y = -25;
        f.x = Math.random() * this.width;
      }
    }

    for (let i = 0; i < this.needles.length; i++) {
      const n = this.needles[i];
      n.swirlPhase += clampedDt * 3;
      n.x += (n.speedX + gust) * clampedDt;
      n.y += (n.speedY + Math.sin(n.swirlPhase) * 20) * clampedDt;
      n.angle += n.rotSpeed * clampedDt;
      if (n.x > this.width + 30) {
        n.x = -30;
        n.y = Math.random() * this.height;
      }
      if (n.y > this.height + 20) {
        n.y = -20;
        n.x = Math.random() * this.width;
      }
    }
  }

  render(context: PresetContext): void {
    if (this.flakes.length === 0 || this.bokehCanvases.length === 0) return;
    const { ctx } = context;

    for (let i = 0; i < this.flakes.length; i++) {
      const f = this.flakes[i];
      const swirlX = this.isBlizzard ? Math.sin(f.swirlPhase) * f.swirlAmp : 0;
      const swirlY = this.isBlizzard ? Math.cos(f.swirlPhase) * (f.swirlAmp * 0.4) : 0;
      const posX = this.isBlizzard ? f.x + swirlX : f.x + Math.sin(f.swayPhase) * f.swayAmp;
      const posY = f.y + swirlY;

      if (f.isCrystal) {
        ctx.save();
        ctx.globalAlpha = f.alpha;
        ctx.translate(posX, posY);
        ctx.rotate(f.crystalAngle);
        ctx.shadowColor = 'rgba(18, 40, 70, 0.35)';
        ctx.shadowBlur = 4;
        ctx.strokeStyle = '#FFFFFF';
        ctx.lineWidth = f.layer === 2 ? 1.3 : 0.9;
        ctx.lineCap = 'round';

        const arm = f.crystalSize;
        ctx.beginPath();
        for (let a = 0; a < 3; a++) {
          const ang = (a * Math.PI) / 3, cos = Math.cos(ang), sin = Math.sin(ang);
          ctx.moveTo(-arm * cos, -arm * sin);
          ctx.lineTo(arm * cos, arm * sin);
          if (f.layer > 0) {
            const bd = arm * 0.65, br = arm * 0.35;
            for (const sign of [-1, 1]) {
              const cx = sign * bd * cos, cy = sign * bd * sin;
              for (const da of [-Math.PI / 4, Math.PI / 4]) {
                ctx.moveTo(cx, cy);
                ctx.lineTo(cx + sign * br * Math.cos(ang + da), cy + sign * br * Math.sin(ang + da));
              }
            }
          }
        }
        ctx.stroke();
        ctx.restore();
      } else {
        const stamp = this.bokehCanvases[f.layer];
        if (!stamp) continue;
        const halfSize = stamp.width / this.dpr / 2;
        ctx.save();
        ctx.globalAlpha = f.alpha;
        ctx.shadowColor = 'rgba(20, 45, 75, 0.28)';
        ctx.shadowBlur = 4;
        ctx.drawImage(stamp, posX - halfSize, posY - halfSize, stamp.width / this.dpr, stamp.height / this.dpr);
        ctx.restore();
      }
    }

    if (this.isBlizzard && this.needles.length > 0) {
      for (let i = 0; i < this.needles.length; i++) {
        const n = this.needles[i];
        ctx.save();
        ctx.globalAlpha = n.alpha;
        ctx.translate(n.x, n.y);
        ctx.rotate(n.angle);
        ctx.strokeStyle = '#274435';
        ctx.lineWidth = 1.3;
        ctx.lineCap = 'round';
        ctx.beginPath();
        ctx.moveTo(-n.length / 2, 0);
        ctx.quadraticCurveTo(0, 1.5, n.length / 2, 0);
        ctx.stroke();
        ctx.restore();
      }
    }
  }

  destroy(): void {
    this.flakes = [];
    this.needles = [];
    this.bokehCanvases = [];
  }
}

