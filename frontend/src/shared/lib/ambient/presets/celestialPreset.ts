import type { AmbientPreset, PresetContext, AmbientIntensity } from '../types';

interface Star {
  baseX: number;
  baseY: number;
  x: number;
  y: number;
  vx: number;
  vy: number;
  radius: number;
  baseAlpha: number;
  twinklePhase: number;
  twinkleSpeed: number;
  isAnchor: boolean;
  color: 'gold' | 'lapis' | 'white';
}

interface ActiveWave {
  x: number;
  y: number;
  radius: number;
  maxRadius: number;
  speed: number;
  alpha: number;
}

export class CelestialPreset implements AmbientPreset {
  private stars: Star[] = [];
  private activeWaves: ActiveWave[] = [];
  private width = 0;
  private height = 0;
  private currentIntensity: AmbientIntensity = 'subtle';
  private processedClickTimes = new Set<number>();

  init(_ctx: CanvasRenderingContext2D, width: number, height: number): void {
    this.width = width;
    this.height = height;
    this.rebuildStars();
  }

  resize(width: number, height: number): void {
    this.width = width;
    this.height = height;
  }

  private getTargetCount(intensity: AmbientIntensity): number {
    if (intensity === 'vivid') return 130;
    if (intensity === 'subtle') return 65;
    return 0;
  }

  private rebuildStars(): void {
    const count = this.getTargetCount(this.currentIntensity);
    this.stars = [];
    if (count === 0 || this.width === 0 || this.height === 0) return;

    for (let i = 0; i < count; i++) {
      const isAnchor = Math.random() < 0.15;
      const colorRoll = Math.random();
      const color: 'gold' | 'lapis' | 'white' =
        colorRoll < 0.45 ? 'gold' : colorRoll < 0.75 ? 'lapis' : 'white';
      const x = Math.random() * this.width;
      const y = Math.random() * this.height;

      this.stars.push({
        baseX: x,
        baseY: y,
        x,
        y,
        vx: 0,
        vy: 0,
        radius: isAnchor ? 1.8 + Math.random() * 0.8 : 0.8 + Math.random() * 0.8,
        baseAlpha: isAnchor ? 0.85 : 0.4 + Math.random() * 0.4,
        twinklePhase: Math.random() * Math.PI * 2,
        twinkleSpeed: 0.8 + Math.random() * 1.5,
        isAnchor,
        color,
      });
    }
  }

  update(dt: number, context: PresetContext): void {
    if (context.intensity !== this.currentIntensity) {
      this.currentIntensity = context.intensity;
      this.rebuildStars();
    }

    if (this.stars.length === 0) return;
    const clampedDt = Math.min(dt, 0.1);

    for (let i = 0; i < context.clicks.length; i++) {
      const click = context.clicks[i];
      if (!this.processedClickTimes.has(click.time)) {
        this.processedClickTimes.add(click.time);
        this.activeWaves.push({
          x: click.x,
          y: click.y,
          radius: 0,
          maxRadius: click.maxRadius || 240,
          speed: 380,
          alpha: 1,
        });
      }
    }

    if (this.processedClickTimes.size > 30) {
      this.processedClickTimes.clear();
    }

    for (let i = this.activeWaves.length - 1; i >= 0; i--) {
      const wave = this.activeWaves[i];
      wave.radius += wave.speed * clampedDt;
      wave.alpha = Math.max(0, 1 - wave.radius / wave.maxRadius);

      for (let j = 0; j < this.stars.length; j++) {
        const star = this.stars[j];
        const dx = star.x - wave.x;
        const dy = star.y - wave.y;
        const dist = Math.hypot(dx, dy);
        const waveFrontDist = Math.abs(dist - wave.radius);

        if (waveFrontDist < 25 && dist > 1) {
          const pushForce = (1 - waveFrontDist / 25) * wave.alpha * 45;
          star.vx += (dx / dist) * pushForce;
          star.vy += (dy / dist) * pushForce;
        }
      }

      if (wave.radius >= wave.maxRadius || wave.alpha <= 0) {
        this.activeWaves.splice(i, 1);
      }
    }

    for (let i = 0; i < this.stars.length; i++) {
      const star = this.stars[i];
      star.twinklePhase += star.twinkleSpeed * clampedDt;

      const dx = star.baseX - star.x;
      const dy = star.baseY - star.y;
      star.vx += dx * 6 * clampedDt;
      star.vy += dy * 6 * clampedDt;
      star.vx *= 0.88;
      star.vy *= 0.88;

      star.x += star.vx * clampedDt;
      star.y += star.vy * clampedDt;
    }
  }

  render(context: PresetContext): void {
    if (this.stars.length === 0) return;
    const { ctx, mouse } = context;

    const now = performance.now();
    const timeSinceLastMove = (now - mouse.lastMoveTime) / 1000;
    const constellationFade = mouse.isOver
      ? Math.max(0, Math.min(1, 1 - (timeSinceLastMove - 1.5) / 1.5))
      : 0;

    if (constellationFade > 0.01) {
      const cursorRadius = 140;
      const nearbyStars: Star[] = [];

      ctx.save();
      ctx.lineWidth = 0.85;

      for (let i = 0; i < this.stars.length; i++) {
        const s = this.stars[i];
        const dist = Math.hypot(s.x - mouse.x, s.y - mouse.y);
        if (dist < cursorRadius) {
          nearbyStars.push(s);
          const alpha = (1 - dist / cursorRadius) * constellationFade * 0.55;
          ctx.strokeStyle = `rgba(212, 168, 83, ${alpha})`;
          ctx.beginPath();
          ctx.moveTo(s.x, s.y);
          ctx.lineTo(mouse.x, mouse.y);
          ctx.stroke();
        }
      }

      for (let i = 0; i < nearbyStars.length; i++) {
        for (let j = i + 1; j < nearbyStars.length; j++) {
          const s1 = nearbyStars[i];
          const s2 = nearbyStars[j];
          const dist = Math.hypot(s1.x - s2.x, s1.y - s2.y);
          if (dist < 95) {
            const alpha = (1 - dist / 95) * constellationFade * 0.4;
            ctx.strokeStyle = `rgba(212, 168, 83, ${alpha})`;
            ctx.beginPath();
            ctx.moveTo(s1.x, s1.y);
            ctx.lineTo(s2.x, s2.y);
            ctx.stroke();
          }
        }
      }
      ctx.restore();
    }

    for (let i = 0; i < this.activeWaves.length; i++) {
      const w = this.activeWaves[i];
      ctx.save();
      ctx.lineWidth = 1.2;
      ctx.strokeStyle = `rgba(212, 168, 83, ${w.alpha * 0.5})`;
      ctx.beginPath();
      ctx.arc(w.x, w.y, w.radius, 0, Math.PI * 2);
      ctx.stroke();
      ctx.restore();
    }

    for (let i = 0; i < this.stars.length; i++) {
      const s = this.stars[i];
      const brightness = 0.6 + 0.4 * Math.sin(s.twinklePhase);
      const alpha = Math.min(1, s.baseAlpha * brightness);

      ctx.save();
      ctx.globalAlpha = alpha;

      let fill = '#FFFFFF';
      if (s.color === 'gold') fill = '#E6BA65';
      else if (s.color === 'lapis') fill = '#82A7FF';

      ctx.fillStyle = fill;
      ctx.beginPath();
      ctx.arc(s.x, s.y, s.radius, 0, Math.PI * 2);
      ctx.fill();

      if (s.isAnchor && alpha > 0.65) {
        ctx.strokeStyle = fill;
        ctx.lineWidth = 0.5;
        const spike = s.radius * 2.4;
        ctx.beginPath();
        ctx.moveTo(s.x - spike, s.y);
        ctx.lineTo(s.x + spike, s.y);
        ctx.moveTo(s.x, s.y - spike);
        ctx.lineTo(s.x, s.y + spike);
        ctx.stroke();
      }

      ctx.restore();
    }
  }

  destroy(): void {
    this.stars = [];
    this.activeWaves = [];
    this.processedClickTimes.clear();
  }
}

