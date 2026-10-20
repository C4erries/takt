import type { AmbientPreset, PresetContext, AmbientIntensity } from '../types';

interface LightShaft {
  baseX: number;
  widthTop: number;
  widthBottom: number;
  slant: number;
  swaySpeed: number;
  swayPhase: number;
  alpha: number;
}

interface Bubble {
  x: number;
  y: number;
  radius: number;
  speedY: number;
  swaySpeed: number;
  swayAmp: number;
  swayPhase: number;
  alpha: number;
}

interface TropicalFish {
  x: number;
  y: number;
  baseY: number;
  speedX: number;
  direction: 1 | -1;
  size: number;
  colorBody1: string;
  colorBody2: string;
  colorFin: string;
  time: number;
  wagSpeed: number;
  swimPhase: number;
  swimSpeed: number;
}

export class OceanPreset implements AmbientPreset {
  private shafts: LightShaft[] = [];
  private bubbles: Bubble[] = [];
  private fishes: TropicalFish[] = [];
  private width = 0;
  private height = 0;
  private currentIntensity: AmbientIntensity = 'subtle';

  init(_ctx: CanvasRenderingContext2D, width: number, height: number): void {
    this.width = width;
    this.height = height;
    this.rebuildElements();
  }

  resize(width: number, height: number): void {
    this.width = width;
    this.height = height;
  }

  private rebuildElements(): void {
    const isVivid = this.currentIntensity === 'vivid';
    const shaftCount = isVivid ? 5 : 4;
    const bubbleCount = isVivid ? 32 : 18;
    const fishCount = isVivid ? 6 : 4;

    this.shafts = [];
    this.bubbles = [];
    this.fishes = [];
    if (this.width === 0 || this.height === 0 || this.currentIntensity === 'off') return;

    for (let i = 0; i < shaftCount; i++) {
      this.shafts.push({
        baseX: (this.width / (shaftCount + 1)) * (i + 1) + (Math.random() - 0.5) * 60,
        widthTop: 50 + Math.random() * 40,
        widthBottom: 140 + Math.random() * 80,
        slant: (Math.random() - 0.5) * 120,
        swaySpeed: 0.6 + Math.random() * 0.6,
        swayPhase: Math.random() * Math.PI * 2,
        alpha: isVivid ? 0.12 : 0.08,
      });
    }

    for (let i = 0; i < bubbleCount; i++) {
      this.bubbles.push({
        x: Math.random() * this.width,
        y: Math.random() * (this.height + 40),
        radius: 1.5 + Math.random() * 2.8,
        speedY: 28 + Math.random() * 32,
        swaySpeed: 1.2 + Math.random() * 1.5,
        swayAmp: 8 + Math.random() * 12,
        swayPhase: Math.random() * Math.PI * 2,
        alpha: 0.4 + Math.random() * 0.45,
      });
    }

    const fishPalette = [
      { c1: '#2DD4BF', c2: '#0284C7', fin: '#38BDF8' }, // Bioluminescent Azure
      { c1: '#FB923C', c2: '#E11D48', fin: '#F43F5E' }, // Coral Flame
      { c1: '#FBBF24', c2: '#D97706', fin: '#F59E0B' }, // Golden Tang
      { c1: '#34D399', c2: '#0D9488', fin: '#2DD4BF' }, // Emerald Teal
      { c1: '#A78BFA', c2: '#4F46E5', fin: '#818CF8' }, // Deep Violet Orchid
      { c1: '#38BDF8', c2: '#1D4ED8', fin: '#60A5FA' }, // Lapis Lazuli
    ];

    for (let i = 0; i < fishCount; i++) {
      const p = fishPalette[i % fishPalette.length];
      const dir: 1 | -1 = Math.random() > 0.5 ? 1 : -1;
      const baseY = 80 + Math.random() * Math.max(100, this.height - 200);

      this.fishes.push({
        x: Math.random() * this.width,
        y: baseY,
        baseY,
        speedX: 24 + Math.random() * 26,
        direction: dir,
        size: 16 + Math.random() * 8,
        colorBody1: p.c1,
        colorBody2: p.c2,
        colorFin: p.fin,
        time: Math.random() * 10,
        wagSpeed: 4.5 + Math.random() * 2.5,
        swimPhase: Math.random() * Math.PI * 2,
        swimSpeed: 0.8 + Math.random() * 0.8,
      });
    }
  }

  update(dt: number, context: PresetContext): void {
    if (context.intensity !== this.currentIntensity) {
      this.currentIntensity = context.intensity;
      this.rebuildElements();
    }
    const clampedDt = Math.min(dt, 0.1);

    for (const s of this.shafts) {
      s.swayPhase += s.swaySpeed * clampedDt;
    }

    for (const b of this.bubbles) {
      b.y -= b.speedY * clampedDt;
      b.swayPhase += b.swaySpeed * clampedDt;
      if (b.y < -20) {
        b.y = this.height + 20;
        b.x = Math.random() * this.width;
      }
    }

    for (const f of this.fishes) {
      f.time += clampedDt;
      f.swimPhase += f.swimSpeed * clampedDt;
      f.x += f.speedX * f.direction * clampedDt;
      f.y = f.baseY + Math.sin(f.swimPhase) * 14;

      if (f.direction === 1 && f.x > this.width + 60) {
        f.direction = -1;
        f.baseY = 80 + Math.random() * Math.max(100, this.height - 200);
      } else if (f.direction === -1 && f.x < -60) {
        f.direction = 1;
        f.baseY = 80 + Math.random() * Math.max(100, this.height - 200);
      }
    }
  }

  render(context: PresetContext): void {
    const { ctx } = context;

    // 1. Light shafts
    for (const s of this.shafts) {
      const topX = s.baseX + Math.sin(s.swayPhase) * 25;
      const bottomX = topX + s.slant + Math.cos(s.swayPhase * 0.7) * 45;
      const grad = ctx.createLinearGradient(topX, 0, bottomX, this.height * 0.85);
      grad.addColorStop(0, `rgba(45, 212, 191, ${s.alpha})`);
      grad.addColorStop(0.4, `rgba(56, 189, 248, ${s.alpha * 0.5})`);
      grad.addColorStop(1, 'rgba(10, 20, 30, 0)');

      ctx.save();
      ctx.fillStyle = grad;
      ctx.beginPath();
      ctx.moveTo(topX - s.widthTop / 2, 0);
      ctx.lineTo(topX + s.widthTop / 2, 0);
      ctx.lineTo(bottomX + s.widthBottom / 2, this.height * 0.9);
      ctx.lineTo(bottomX - s.widthBottom / 2, this.height * 0.9);
      ctx.closePath();
      ctx.fill();
      ctx.restore();
    }

    // 2. Coral silhouettes on seabed
    ctx.save();
    ctx.fillStyle = 'rgba(16, 36, 50, 0.35)';
    const coralSteps = 6;
    const stepW = this.width / coralSteps;
    for (let i = 0; i <= coralSteps; i++) {
      const cx = i * stepW;
      const h = 28 + ((i * 17) % 24);
      ctx.beginPath();
      ctx.moveTo(cx - 24, this.height);
      ctx.quadraticCurveTo(cx - 16, this.height - h, cx, this.height - h - 6);
      ctx.quadraticCurveTo(cx + 16, this.height - h, cx + 24, this.height);
      ctx.fill();
    }
    ctx.restore();

    // 3. Pearlescent bubbles
    for (const b of this.bubbles) {
      const posX = b.x + Math.sin(b.swayPhase) * b.swayAmp;
      ctx.save();
      ctx.globalAlpha = b.alpha;
      ctx.strokeStyle = 'rgba(224, 238, 248, 0.6)';
      ctx.lineWidth = 0.8;
      ctx.fillStyle = 'rgba(45, 212, 191, 0.12)';
      ctx.beginPath();
      ctx.arc(posX, b.y, b.radius, 0, Math.PI * 2);
      ctx.fill();
      ctx.stroke();

      // Specular highlight
      ctx.fillStyle = '#FFFFFF';
      ctx.beginPath();
      ctx.arc(posX - b.radius * 0.35, b.y - b.radius * 0.35, b.radius * 0.28, 0, Math.PI * 2);
      ctx.fill();
      ctx.restore();
    }

    // 4. Peaceful tropical fish with tail and fin wagging
    for (const f of this.fishes) {
      const tailAngle = Math.sin(f.time * f.wagSpeed) * 0.35;
      const finAngle = Math.cos(f.time * (f.wagSpeed * 1.1)) * 0.25;
      const len = f.size * 1.8;
      const h = f.size * 0.65;

      ctx.save();
      ctx.translate(f.x, f.y);
      ctx.scale(f.direction, 1);

      // Dorsal Fin
      ctx.save();
      ctx.translate(-len * 0.05, -h * 0.7);
      ctx.rotate(finAngle);
      ctx.fillStyle = f.colorFin;
      ctx.globalAlpha = 0.8;
      ctx.beginPath();
      ctx.moveTo(0, 0);
      ctx.quadraticCurveTo(-len * 0.2, -h * 0.8, -len * 0.35, 0);
      ctx.fill();
      ctx.restore();

      // Wagging Tail Fin
      ctx.save();
      ctx.translate(-len * 0.45, 0);
      ctx.rotate(tailAngle);
      ctx.fillStyle = f.colorFin;
      ctx.globalAlpha = 0.85;
      ctx.beginPath();
      ctx.moveTo(0, 0);
      ctx.lineTo(-len * 0.45, -h * 0.75);
      ctx.quadraticCurveTo(-len * 0.35, 0, -len * 0.45, h * 0.75);
      ctx.closePath();
      ctx.fill();
      ctx.restore();

      // Fish Body
      const bodyGrad = ctx.createLinearGradient(-len * 0.45, 0, len * 0.55, 0);
      bodyGrad.addColorStop(0, f.colorBody2);
      bodyGrad.addColorStop(1, f.colorBody1);
      ctx.fillStyle = bodyGrad;
      ctx.beginPath();
      ctx.moveTo(len * 0.55, 0);
      ctx.quadraticCurveTo(0, -h, -len * 0.45, 0);
      ctx.quadraticCurveTo(0, h, len * 0.55, 0);
      ctx.fill();

      // Eye with glint
      ctx.fillStyle = '#0A141E';
      ctx.beginPath();
      ctx.arc(len * 0.32, -h * 0.18, f.size * 0.12, 0, Math.PI * 2);
      ctx.fill();
      ctx.fillStyle = '#FFFFFF';
      ctx.beginPath();
      ctx.arc(len * 0.34, -h * 0.22, f.size * 0.05, 0, Math.PI * 2);
      ctx.fill();

      ctx.restore();
    }
  }

  destroy(): void {
    this.shafts = [];
    this.bubbles = [];
    this.fishes = [];
  }
}
