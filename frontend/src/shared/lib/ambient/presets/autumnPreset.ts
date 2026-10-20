import type { AmbientPreset, PresetContext, AmbientIntensity } from '../types';

interface LeafParticle {
  stampIndex: number;
  x: number;
  y: number;
  size: number;
  speedX: number;
  speedY: number;
  targetSpeedY: number;
  angle: number;
  angularVelocity: number;
  flipAngle: number;
  flipSpeed: number;
  swayPhase: number;
  swaySpeed: number;
  alpha: number;
  initialAlpha: number;
  isGrounded: boolean;
  groundY: number;
  groundTimer: number;
  groundLifetime: number;
  isFading: boolean;
}

interface EmberParticle {
  x: number;
  y: number;
  radius: number;
  speedY: number;
  swayPhase: number;
  swaySpeed: number;
  alpha: number;
  pulsePhase: number;
}

export class AutumnPreset implements AmbientPreset {
  private leaves: LeafParticle[] = [];
  private embers: EmberParticle[] = [];
  private stamps: HTMLCanvasElement[] = [];
  private width = 0;
  private height = 0;
  private currentIntensity: AmbientIntensity = 'subtle';

  init(_ctx: CanvasRenderingContext2D, width: number, height: number): void {
    this.width = width;
    this.height = height;
    this.initStamps();
    this.rebuildParticles();
  }

  resize(width: number, height: number): void {
    this.width = width;
    this.height = height;
    for (const l of this.leaves) {
      l.groundY = this.height - 12 - Math.random() * 26;
    }
  }

  private initStamps(): void {
    if (typeof document === 'undefined' || this.stamps.length > 0) return;
    const momijiPath = new Path2D('M0,18 Q-8,14 -22,8 Q-14,2 -12,-2 Q-18,-8 -22,-16 Q-12,-16 -6,-12 Q-4,-22 0,-28 Q4,-22 6,-12 Q12,-16 22,-16 Q18,-8 12,-2 Q14,2 22,8 Q8,14 0,18 Z');
    const ginkgoPath = new Path2D('M0,8 C-12,4 -26,-6 -24,-18 C-18,-26 -6,-24 0,-17 C6,-24 18,-26 24,-18 C26,-6 12,4 0,8 Z');
    const configs = [
      { isM: true, c1: '#8E1B15', c2: '#D9483B' }, { isM: true, c1: '#BA3E18', c2: '#E87042' },
      { isM: true, c1: '#6E1410', c2: '#A82820' }, { isM: false, c1: '#C67812', c2: '#F2C240' },
      { isM: false, c1: '#B0640A', c2: '#E5A024' }, { isM: false, c1: '#8E7818', c2: '#D4B02E' },
    ];

    this.stamps = configs.map(({ isM, c1, c2 }) => {
      const cvs = document.createElement('canvas');
      cvs.width = 64; cvs.height = 64;
      const sCtx = cvs.getContext('2d');
      if (!sCtx) return cvs;
      sCtx.translate(32, 32);

      const grad = sCtx.createLinearGradient(0, 18, 0, isM ? -28 : -24);
      grad.addColorStop(0, c1); grad.addColorStop(1, c2);
      sCtx.fillStyle = grad;
      sCtx.fill(isM ? momijiPath : ginkgoPath);

      sCtx.strokeStyle = c1; sCtx.lineWidth = 1.2; sCtx.lineCap = 'round';
      sCtx.beginPath();
      if (isM) {
        sCtx.moveTo(0, 18); sCtx.lineTo(0, 26); sCtx.moveTo(0, 18); sCtx.lineTo(0, -20);
        sCtx.moveTo(0, 4); sCtx.lineTo(-16, -12); sCtx.moveTo(0, 4); sCtx.lineTo(16, -12);
      } else {
        sCtx.moveTo(0, 8); sCtx.lineTo(0, 26); sCtx.moveTo(0, 7); sCtx.lineTo(-15, -18);
        sCtx.moveTo(0, 7); sCtx.lineTo(15, -18); sCtx.moveTo(0, 7); sCtx.lineTo(-7, -20);
        sCtx.moveTo(0, 7); sCtx.lineTo(7, -20);
      }
      sCtx.stroke();
      return cvs;
    });
  }

  private createLeaf(fromTop = false, onGround = false): LeafParticle {
    const targetSpeedY = 22 + Math.random() * 24;
    const y = onGround
      ? this.height - 12 - Math.random() * 26
      : fromTop
      ? -40 - Math.random() * Math.max(150, this.height * 0.8)
      : Math.random() * (this.height + 60) - 40;
    const alpha = 0.82 + Math.random() * 0.18;

    return {
      stampIndex: Math.floor(Math.random() * 6),
      x: Math.random() * this.width,
      y,
      size: 26 + Math.random() * 12,
      speedX: (Math.random() - 0.5) * 20,
      speedY: onGround ? 0 : targetSpeedY,
      targetSpeedY,
      angle: Math.random() * Math.PI * 2,
      angularVelocity: onGround ? 0 : (Math.random() - 0.5) * 1.5,
      flipAngle: onGround ? Math.round(Math.random()) * Math.PI : Math.random() * Math.PI * 2,
      flipSpeed: 1.1 + Math.random() * 1.5,
      swayPhase: Math.random() * Math.PI * 2,
      swaySpeed: 0.8 + Math.random() * 0.8,
      alpha,
      initialAlpha: alpha,
      isGrounded: onGround,
      groundY: this.height - 12 - Math.random() * 26,
      groundTimer: onGround ? Math.random() * 8 : 0,
      groundLifetime: 12 + Math.random() * 6,
      isFading: false,
    };
  }

  private rebuildParticles(): void {
    const isVivid = this.currentIntensity === 'vivid';
    const totalLeaves = isVivid ? 30 : 18;
    const groundedStart = isVivid ? 14 : 9;
    const targetEmbers = isVivid ? 14 : 8;

    this.leaves = [];
    this.embers = [];
    if (this.width === 0 || this.height === 0) return;

    for (let i = 0; i < totalLeaves; i++) {
      this.leaves.push(this.createLeaf(false, i < groundedStart));
    }
    for (let i = 0; i < targetEmbers; i++) {
      this.embers.push({
        x: Math.random() * this.width,
        y: Math.random() * this.height,
        radius: 1 + Math.random() * 1.8,
        speedY: 14 + Math.random() * 20,
        swayPhase: Math.random() * Math.PI * 2,
        swaySpeed: 1.2 + Math.random() * 1.5,
        alpha: 0.4 + Math.random() * 0.5,
        pulsePhase: Math.random() * Math.PI * 2,
      });
    }
  }

  update(dt: number, context: PresetContext): void {
    if (context.intensity !== this.currentIntensity) {
      this.currentIntensity = context.intensity;
      this.rebuildParticles();
    }

    const clampedDt = Math.min(dt, 0.1);
    const { mouse, isInteractive } = context;

    for (const leaf of this.leaves) {
      if (leaf.isGrounded) {
        leaf.groundTimer += clampedDt;
        if (leaf.groundTimer >= leaf.groundLifetime) {
          leaf.isFading = true;
        }
      }

      if (leaf.isFading) {
        leaf.alpha -= clampedDt * 0.15;
        if (leaf.alpha <= 0) {
          Object.assign(leaf, this.createLeaf(true, false));
          continue;
        }
      }

      if (isInteractive && mouse.isOver) {
        const dx = leaf.x - mouse.x;
        const dy = leaf.y - mouse.y;
        const dist = Math.hypot(dx, dy);

        if (dist < 140) {
          const force = 1 - dist / 140;
          leaf.isGrounded = false;
          leaf.isFading = false;
          leaf.groundTimer = 0;
          leaf.alpha = leaf.initialAlpha;
          leaf.speedY = -120 - Math.random() * 100 * force;
          leaf.speedX = (dx / (dist || 1)) * 170 * force + (Math.random() - 0.5) * 70;
          leaf.angularVelocity = (Math.random() - 0.5) * 8;
          leaf.flipSpeed = 2.5 + Math.random() * 2.5;
        }
      }

      if (leaf.isGrounded) continue;

      if (leaf.y >= leaf.groundY - 20) {
        leaf.speedY *= 0.86;
        leaf.speedX *= 0.86;
        leaf.angularVelocity *= 0.85;
        leaf.flipSpeed *= 0.85;

        const nearestFlat = Math.round(leaf.flipAngle / Math.PI) * Math.PI;
        leaf.flipAngle += (nearestFlat - leaf.flipAngle) * 0.18;

        if (leaf.y >= leaf.groundY) {
          leaf.y = leaf.groundY;
          if (Math.abs(leaf.speedY) < 1.8) {
            leaf.speedY = 0;
            leaf.speedX = 0;
            leaf.angularVelocity = 0;
            leaf.flipAngle = nearestFlat;
            leaf.isGrounded = true;
            leaf.groundTimer = 0;
            leaf.groundLifetime = 12 + Math.random() * 6;
            continue;
          }
        }
      } else {
        leaf.speedY += 80 * clampedDt;
        if (leaf.speedY > leaf.targetSpeedY) {
          leaf.speedY += (leaf.targetSpeedY - leaf.speedY) * clampedDt * 2;
        }
        leaf.swayPhase += leaf.swaySpeed * clampedDt;
        leaf.speedX += Math.cos(leaf.swayPhase) * 16 * clampedDt;
        leaf.speedX *= Math.pow(0.92, clampedDt * 60);
      }

      leaf.y += leaf.speedY * clampedDt;
      leaf.x += leaf.speedX * clampedDt;
      leaf.angle += leaf.angularVelocity * clampedDt;
      leaf.flipAngle += leaf.flipSpeed * clampedDt;

      if (leaf.x > this.width + 50) leaf.x = -40;
      if (leaf.x < -50) leaf.x = this.width + 40;
    }

    for (const ember of this.embers) {
      ember.y -= ember.speedY * clampedDt;
      ember.swayPhase += ember.swaySpeed * clampedDt;
      ember.pulsePhase += clampedDt * 2.5;
      if (ember.y < -20) {
        ember.y = this.height + 20;
        ember.x = Math.random() * this.width;
      }
    }
  }

  render(context: PresetContext): void {
    const { ctx } = context;

    for (const e of this.embers) {
      const swayX = e.x + Math.sin(e.swayPhase) * 12;
      const pulsingAlpha = Math.max(0.1, e.alpha * (0.6 + 0.4 * Math.sin(e.pulsePhase)));
      ctx.save();
      ctx.globalAlpha = pulsingAlpha;
      ctx.fillStyle = '#E89838';
      ctx.shadowColor = '#D97746';
      ctx.shadowBlur = 8;
      ctx.beginPath();
      ctx.arc(swayX, e.y, e.radius, 0, Math.PI * 2);
      ctx.fill();
      ctx.restore();
    }

    if (this.stamps.length === 0) return;

    for (const l of this.leaves) {
      const stamp = this.stamps[l.stampIndex];
      if (!stamp) continue;

      ctx.save();
      ctx.globalAlpha = l.alpha;
      ctx.translate(l.x, l.y);
      ctx.rotate(l.angle);
      ctx.scale(Math.cos(l.flipAngle), 1);
      ctx.drawImage(stamp, -l.size / 2, -l.size / 2, l.size, l.size);
      ctx.restore();
    }
  }

  destroy(): void {
    this.leaves = [];
    this.embers = [];
    this.stamps = [];
  }
}
