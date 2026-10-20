import type { AmbientPreset, PresetContext, AmbientIntensity } from '../../types';
import type { TropicalFish, LightShaft, Bubble, FishSpecies } from './oceanTypes';
import { renderFish } from './oceanFish';
import { OceanReef } from './oceanReef';
import { OceanCrab } from './oceanCrab';

export class OceanPreset implements AmbientPreset {
  private shafts: LightShaft[] = [];
  private bubbles: Bubble[] = [];
  private fishes: TropicalFish[] = [];
  private reef = new OceanReef();
  private crab = new OceanCrab();
  private width = 0;
  private height = 0;
  private totalTime = 0;
  private currentIntensity: AmbientIntensity = 'subtle';

  init(_ctx: CanvasRenderingContext2D, width: number, height: number): void {
    this.width = width;
    this.height = height;
    this.rebuildElements();
  }

  resize(width: number, height: number): void {
    this.width = width;
    this.height = height;
    const isVivid = this.currentIntensity === 'vivid';
    this.reef.resize(width, height, isVivid);
    this.crab.resize(height);
  }

  setIntensity(intensity: AmbientIntensity): void {
    if (this.currentIntensity === intensity) return;
    this.currentIntensity = intensity;
    this.rebuildElements();
  }

  private rebuildElements(): void {
    const isVivid = this.currentIntensity === 'vivid';
    this.shafts = [];
    this.bubbles = [];
    this.fishes = [];
    if (this.width === 0 || this.height === 0 || this.currentIntensity === 'off') return;

    // 1. Light shafts
    const shaftCount = isVivid ? 5 : 2;
    for (let i = 0; i < shaftCount; i++) {
      this.shafts.push({
        baseX: (this.width / (shaftCount + 1)) * (i + 1) + (Math.random() - 0.5) * 60,
        widthTop: 55 + Math.random() * 45,
        widthBottom: 150 + Math.random() * 85,
        slant: (Math.random() - 0.5) * 120,
        swaySpeed: 0.6 + Math.random() * 0.6,
        swayPhase: Math.random() * Math.PI * 2,
        alpha: isVivid ? 0.12 : 0.08,
      });
    }

    // 2. Pearlescent bubbles (both far and near)
    const bubbleCount = isVivid ? 30 : 10;
    for (let i = 0; i < bubbleCount; i++) {
      const isForeground = i < bubbleCount * 0.35;
      this.bubbles.push({
        x: Math.random() * this.width,
        y: Math.random() * (this.height + 40),
        radius: isForeground ? 2.5 + Math.random() * 3.0 : 1.2 + Math.random() * 1.8,
        speedY: (isForeground ? 32 : 22) + Math.random() * 30,
        swaySpeed: 1.1 + Math.random() * 1.4,
        swayAmp: 8 + Math.random() * 14,
        swayPhase: Math.random() * Math.PI * 2,
        alpha: isForeground ? 0.55 + Math.random() * 0.35 : 0.3 + Math.random() * 0.3,
        isForeground,
      });
    }

    // 3. Tropical Fishes (4 anatomical species, scaled by +35%)
    const speciesCatalog: FishSpecies[] = ['clownfish', 'bluetang', 'angelfish', 'yellowtang'];
    const fishCount = isVivid ? 12 : 4;
    for (let i = 0; i < fishCount; i++) {
      const species = speciesCatalog[i % speciesCatalog.length];
      const dir: 1 | -1 = Math.random() > 0.5 ? 1 : -1;
      const baseY = 75 + Math.random() * Math.max(100, this.height - 230);
      const size =
        species === 'clownfish' ? 24 : species === 'bluetang' ? 26 : species === 'angelfish' ? 29 : 25;

      this.fishes.push({
        species,
        x: Math.random() * this.width,
        y: baseY,
        baseY,
        speedX: 25 + Math.random() * 26,
        direction: dir,
        size,
        time: Math.random() * 10,
        wagSpeed: 4.5 + Math.random() * 2.5,
        swimPhase: Math.random() * Math.PI * 2,
        swimSpeed: 0.8 + Math.random() * 0.8,
      });
    }

    // 4. Modules
    this.reef.init(this.width, this.height, isVivid);
    this.crab.init(this.width, this.height, isVivid);
  }

  update(dt: number, context: PresetContext): void {
    if (context.intensity !== this.currentIntensity) {
      this.currentIntensity = context.intensity;
      this.rebuildElements();
    }
    const clampedDt = Math.min(dt, 0.1);
    this.totalTime += clampedDt;

    // Update light shafts
    for (const s of this.shafts) s.swayPhase += s.swaySpeed * clampedDt;

    // Update bubbles
    for (const b of this.bubbles) {
      b.y -= b.speedY * clampedDt;
      b.swayPhase += b.swaySpeed * clampedDt;
      if (b.y < -25) {
        b.y = this.height + 25;
        b.x = Math.random() * this.width;
      }
    }

    // Update fishes
    for (const f of this.fishes) {
      f.time += clampedDt;
      f.swimPhase += f.swimSpeed * clampedDt;
      f.x += f.speedX * f.direction * clampedDt;
      f.y = f.baseY + Math.sin(f.swimPhase) * 14;

      if (f.direction === 1 && f.x > this.width + 70) {
        f.direction = -1;
        f.baseY = 75 + Math.random() * Math.max(100, this.height - 230);
      } else if (f.direction === -1 && f.x < -70) {
        f.direction = 1;
        f.baseY = 75 + Math.random() * Math.max(100, this.height - 230);
      }
    }

    // Update reef
    this.reef.update(clampedDt, this.totalTime, context.mouse, context.isInteractive);
  }

  render(context: PresetContext): void {
    const { ctx } = context;

    // Layer 1: Caustic Light Shafts
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

    // Layer 2: Background Bubbles
    this.renderBubbleLayer(ctx, false);

    // Layer 3: Seabed Backdrop Ridge
    this.reef.renderBackdrop(ctx, this.width, this.height);

    // Layer 4: Tropical Fishes (4 species with anatomical clipping)
    for (const f of this.fishes) {
      renderFish(ctx, f);
    }

    // Layer 5: Foreground Volumetric Corals, Kelp Ribbons & Starfish
    this.reef.renderForeground(ctx, this.width, this.height, this.totalTime);

    // Layer 6: Scurrying Crab now rendered via OceanCrabDOM in DOM layer

    // Layer 7: Foreground Pearlescent Bubbles with Specular Glints
    this.renderBubbleLayer(ctx, true);
  }

  private renderBubbleLayer(ctx: CanvasRenderingContext2D, foreground: boolean): void {
    for (const b of this.bubbles) {
      if (b.isForeground !== foreground) continue;
      const posX = b.x + Math.sin(b.swayPhase) * b.swayAmp;

      ctx.save();
      ctx.globalAlpha = b.alpha;
      ctx.strokeStyle = 'rgba(224, 238, 248, 0.65)';
      ctx.lineWidth = 0.9;
      ctx.fillStyle = 'rgba(45, 212, 191, 0.12)';
      ctx.beginPath();
      ctx.arc(posX, b.y, b.radius, 0, Math.PI * 2);
      ctx.fill();
      ctx.stroke();

      // Bright specular glint
      ctx.fillStyle = '#FFFFFF';
      ctx.beginPath();
      ctx.arc(posX - b.radius * 0.35, b.y - b.radius * 0.35, b.radius * 0.28, 0, Math.PI * 2);
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
