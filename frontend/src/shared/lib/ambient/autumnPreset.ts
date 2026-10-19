import type { AmbientPreset, PresetContext, AmbientIntensity } from './types';

interface LeafParticle {
  imageIndex: number;
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
  isGrounded: boolean;
  groundY: number;
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

const LEAF_PATHS = [
  '/assets/themes/leaf_momiji_1.png',
  '/assets/themes/leaf_momiji_2.png',
  '/assets/themes/leaf_momiji_3.png',
  '/assets/themes/leaf_ginkgo_1.png',
  '/assets/themes/leaf_ginkgo_2.png',
  '/assets/themes/leaf_ginkgo_3.png',
];

export class AutumnPreset implements AmbientPreset {
  private leaves: LeafParticle[] = [];
  private embers: EmberParticle[] = [];
  private images: HTMLImageElement[] = [];
  private isLoaded = false;
  private width = 0;
  private height = 0;
  private currentIntensity: AmbientIntensity = 'subtle';

  init(_ctx: CanvasRenderingContext2D, width: number, height: number): void {
    this.width = width;
    this.height = height;
    this.loadImages();
    this.rebuildParticles();
  }

  resize(width: number, height: number): void {
    this.width = width;
    this.height = height;
    for (const l of this.leaves) {
      l.groundY = this.height - 12 - Math.random() * 26;
    }
  }

  private loadImages(): void {
    if (typeof window === 'undefined' || this.images.length > 0) return;
    let loaded = 0;
    this.images = LEAF_PATHS.map((src) => {
      const img = new Image();
      img.src = src;
      img.onload = () => {
        if (++loaded === LEAF_PATHS.length) this.isLoaded = true;
      };
      return img;
    });
  }

  private getTargetCounts(intensity: AmbientIntensity): { leaves: number; embers: number } {
    if (intensity === 'vivid') return { leaves: 32, embers: 14 };
    if (intensity === 'subtle') return { leaves: 18, embers: 8 };
    return { leaves: 0, embers: 0 };
  }

  private createLeaf(fromTop = false): LeafParticle {
    const targetSpeedY = 22 + Math.random() * 24;
    // Staggered vertical respawn prevents burst waves and empty pauses
    const y = fromTop
      ? -40 - Math.random() * Math.max(150, this.height * 0.8)
      : Math.random() * (this.height + 60) - 40;

    return {
      imageIndex: Math.floor(Math.random() * LEAF_PATHS.length),
      x: Math.random() * this.width,
      y,
      size: 26 + Math.random() * 12,
      speedX: (Math.random() - 0.5) * 20,
      speedY: targetSpeedY,
      targetSpeedY,
      angle: Math.random() * Math.PI * 2,
      angularVelocity: (Math.random() - 0.5) * 1.5,
      flipAngle: Math.random() * Math.PI * 2,
      flipSpeed: 1.1 + Math.random() * 1.5,
      swayPhase: Math.random() * Math.PI * 2,
      swaySpeed: 0.8 + Math.random() * 0.8,
      alpha: 0.8 + Math.random() * 0.2,
      isGrounded: false,
      groundY: this.height - 12 - Math.random() * 26,
      isFading: false,
    };
  }

  private rebuildParticles(): void {
    const { leaves: targetLeaves, embers: targetEmbers } = this.getTargetCounts(this.currentIntensity);
    this.leaves = [];
    this.embers = [];
    if (this.width === 0 || this.height === 0) return;

    for (let i = 0; i < targetLeaves; i++) this.leaves.push(this.createLeaf(false));
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
    const { mouse } = context;

    let groundedCount = 0;
    for (const l of this.leaves) if (l.isGrounded) groundedCount++;
    const maxGrounded = this.currentIntensity === 'vivid' ? 22 : 12;

    for (const leaf of this.leaves) {
      // Soft fading of oldest leaves on the carpet
      if (leaf.isGrounded && groundedCount > maxGrounded && !leaf.isFading) {
        leaf.isFading = true;
      }
      if (leaf.isFading) {
        leaf.alpha -= clampedDt * 0.45;
        if (leaf.alpha <= 0) {
          Object.assign(leaf, this.createLeaf(true));
          continue;
        }
      }

      // Mouse Wind Scatter / Раздувание листьев курсором
      if (mouse.isOver) {
        const dx = leaf.x - mouse.x;
        const dy = leaf.y - mouse.y;
        const dist = Math.hypot(dx, dy);

        if (dist < 140) {
          const force = 1 - dist / 140;
          leaf.isGrounded = false;
          leaf.isFading = false;
          leaf.alpha = 0.85 + Math.random() * 0.15;
          leaf.speedY = -120 - Math.random() * 100 * force;
          leaf.speedX = (dx / (dist || 1)) * 170 * force + (Math.random() - 0.5) * 70;
          leaf.angularVelocity = (Math.random() - 0.5) * 8;
          leaf.flipSpeed = 2.5 + Math.random() * 2.5;
        }
      }

      if (leaf.isGrounded) {
        // Resting softly on the ground
        continue;
      }

      // Approach ground: smooth progressive dampening (no sudden snapping)
      if (leaf.y >= leaf.groundY - 20) {
        leaf.speedY *= 0.86;
        leaf.speedX *= 0.86;
        leaf.angularVelocity *= 0.85;
        leaf.flipSpeed *= 0.85;

        // Settle flipAngle towards closest flat orientation
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
            continue;
          }
        }
      } else {
        // Airborne motion
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

      // Wrap horizontal bounds
      if (leaf.x > this.width + 50) leaf.x = -40;
      if (leaf.x < -50) leaf.x = this.width + 40;
    }

    // Update rising embers
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

    // Render rising embers
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

    // Unified render pipeline for leaves (both airborne and resting on ground)
    if (!this.isLoaded && this.images.length === 0) return;

    for (const l of this.leaves) {
      const img = this.images[l.imageIndex];
      if (!img || !img.complete) continue;

      ctx.save();
      ctx.globalAlpha = l.alpha;
      ctx.translate(l.x, l.y);
      ctx.rotate(l.angle);
      ctx.scale(Math.cos(l.flipAngle), 1);
      ctx.drawImage(img, -l.size / 2, -l.size / 2, l.size, l.size);
      ctx.restore();
    }
  }

  destroy(): void {
    this.leaves = [];
    this.embers = [];
    this.images = [];
    this.isLoaded = false;
  }
}
