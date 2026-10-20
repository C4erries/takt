import type { KelpBlade } from './oceanTypes';
import type { MouseState } from '../../types';

interface CoralAssets {
  staghorn: HTMLImageElement | null;
  staghornGold: HTMLImageElement | null;
  bushPurple: HTMLImageElement | null;
  branchPink: HTMLImageElement | null;
  anemone: HTMLImageElement | null;
  shelf: HTMLImageElement | null;
  cups: HTMLImageElement | null;
  porous: HTMLImageElement | null;
  seashell: HTMLImageElement | null;
}

export class OceanReef {
  private kelps: KelpBlade[] = [];
  private width = 0;
  private height = 0;
  private bladeImgs: (HTMLImageElement | null)[] = [null, null, null, null, null, null];
  private corals: CoralAssets = {
    staghorn: null, staghornGold: null, bushPurple: null, branchPink: null,
    anemone: null, shelf: null, cups: null, porous: null, seashell: null,
  };

  constructor() {
    if (typeof window !== 'undefined') {
      const load = (src: string) => {
        const img = new Image();
        img.src = src;
        return img;
      };
      for (let i = 1; i <= 5; i++) {
        this.bladeImgs[i] = load(`/assets/ambient/svg/kelp-blade-${i}.svg`);
      }
      this.corals = {
        staghorn: load('/assets/ambient/svg/coral-staghorn.svg'),
        staghornGold: load('/assets/ambient/svg/coral-staghorn-gold.svg'),
        bushPurple: load('/assets/ambient/svg/coral-bush-purple.svg'),
        branchPink: load('/assets/ambient/svg/coral-branch-pink.svg'),
        anemone: load('/assets/ambient/svg/coral-anemone.svg'),
        shelf: load('/assets/ambient/svg/coral-shelf.svg'),
        cups: load('/assets/ambient/svg/coral-cups.svg'),
        porous: load('/assets/ambient/svg/coral-porous.svg'),
        seashell: load('/assets/ambient/svg/seashell.svg'),
      };
    }
  }

  init(width: number, height: number, isVivid: boolean): void {
    this.width = width;
    this.height = height;
    this.rebuildKelps(isVivid);
  }

  resize(width: number, height: number, isVivid: boolean): void {
    this.width = width;
    this.height = height;
    this.rebuildKelps(isVivid);
  }

  private rebuildKelps(isVivid: boolean): void {
    if (!this.width || !this.height) return;
    this.kelps = [];

    // Curated kelp clusters across wings and central passage (32%–68%)
    type BladeDef = [1 | 2 | 3 | 4 | 5, number, number, number, number, number, 'back' | 'mid' | 'front'];
    const clusters: { cx: number; blades: BladeDef[] }[] = [
      { cx: this.width * 0.04, blades: [[4, -8, 220, 0.05, 0.9, 0.2, 'back'], [1, 0, 200, 0.06, 1.1, 1.4, 'mid'], [2, 7, 180, 0.07, 1.2, 2.8, 'front']] },
      { cx: this.width * 0.15, blades: [[3, -7, 205, 0.05, 0.95, 1.1, 'back'], [1, 0, 190, 0.06, 1.05, 2.3, 'mid'], [5, 6, 170, 0.08, 1.15, 3.5, 'front']] },
      { cx: this.width * 0.35, blades: [[4, -6, 190, 0.05, 0.92, 0.6, 'back'], [2, 0, 170, 0.06, 1.1, 1.9, 'mid'], [3, 6, 150, 0.07, 1.25, 3.8, 'front']] },
      { cx: this.width * 0.48, blades: [[1, -7, 215, 0.05, 0.88, 1.5, 'back'], [3, 0, 195, 0.06, 1.08, 2.7, 'mid'], [5, 7, 175, 0.08, 1.2, 4.1, 'front']] },
      { cx: this.width * 0.53, blades: [[2, -7, 225, 0.05, 0.9, 0.9, 'back'], [4, 0, 205, 0.06, 1.05, 2.2, 'mid'], [1, 8, 185, 0.07, 1.18, 4.6, 'front']] },
      { cx: this.width * 0.64, blades: [[3, -6, 185, 0.05, 0.94, 1.3, 'back'], [5, 0, 165, 0.06, 1.12, 2.5, 'mid'], [2, 6, 145, 0.07, 1.22, 4.0, 'front']] },
      { cx: this.width * 0.84, blades: [[4, -7, 215, 0.05, 0.92, 0.7, 'back'], [2, 1, 190, 0.06, 1.05, 2.9, 'mid'], [5, 8, 165, 0.08, 1.3, 4.3, 'front']] },
      { cx: this.width * 0.96, blades: [[1, -8, 230, 0.05, 0.88, 1.8, 'back'], [3, 0, 205, 0.06, 1.08, 3.2, 'mid'], [4, 8, 180, 0.07, 1.2, 5.0, 'front']] },
    ];

    for (const c of clusters) {
      const blades = isVivid ? c.blades : c.blades.slice(0, 2);
      for (const [type, off, h, amp, spd, ph, layer] of blades) {
        this.kelps.push({
          type,
          baseX: c.cx + off,
          baseYOffset: layer === 'back' ? 0 : layer === 'mid' ? 4 : 8,
          width: h * (200 / 300),
          height: h,
          swaySpeed: spd,
          swayPhase: ph,
          swayAmp: amp,
          deflection: 0,
          velocity: 0,
          layer,
        });
      }
    }
  }

  update(dt: number, totalTime: number, mouse: MouseState, isInteractive: boolean): void {
    const cdt = Math.min(dt, 0.1);

    for (let i = 0; i < this.kelps.length; i++) {
      const k = this.kelps[i];
      const naturalSway = Math.sin(totalTime * k.swaySpeed + k.swayPhase) * k.swayAmp;
      const maxPush = k.layer === 'front' ? 0.38 : k.layer === 'mid' ? 0.24 : 0.14;

      if (isInteractive && mouse.isOver) {
        const totalAngle = naturalSway + k.deflection;
        const tipX = k.baseX + Math.sin(totalAngle) * k.height;
        const tipY = (this.height - k.baseYOffset) - Math.cos(totalAngle) * k.height;
        const midX = (k.baseX + tipX) * 0.5;
        const midY = ((this.height - k.baseYOffset) + tipY) * 0.5;
        const dist = Math.min(
          Math.hypot(tipX - mouse.x, tipY - mouse.y),
          Math.hypot(midX - mouse.x, midY - mouse.y)
        );

        if (dist < 220) {
          const factor = 1 - dist / 220;
          const dir = Math.sign(tipX - mouse.x) || 1;
          k.velocity += factor * maxPush * dir * 16.0 * cdt;
        }
      }

      // Independent spring rebound physics
      k.velocity += (-k.deflection * 7.0 - k.velocity * 3.8) * cdt;
      k.deflection += k.velocity * cdt;
    }
  }

  renderBackdrop(ctx: CanvasRenderingContext2D, width: number, height: number): void {
    ctx.save();
    const bgGrad = ctx.createLinearGradient(0, height - 200, 0, height);
    bgGrad.addColorStop(0, '#0D1B2A');
    bgGrad.addColorStop(0.4, '#091522');
    bgGrad.addColorStop(1, '#050D15');
    ctx.fillStyle = bgGrad;

    ctx.beginPath();
    ctx.moveTo(0, height);
    ctx.lineTo(0, height - 170);
    ctx.bezierCurveTo(width * 0.1, height - 195, width * 0.22, height - 150, width * 0.35, height - 65);
    ctx.bezierCurveTo(width * 0.48, height - 42, width * 0.56, height - 44, width * 0.68, height - 75);
    ctx.bezierCurveTo(width * 0.8, height - 160, width * 0.92, height - 200, width, height - 175);
    ctx.lineTo(width, height);
    ctx.closePath();
    ctx.fill();

    const sandGrad = ctx.createLinearGradient(0, height - 40, 0, height);
    sandGrad.addColorStop(0, 'rgba(15, 30, 45, 0.4)');
    sandGrad.addColorStop(1, 'rgba(8, 16, 25, 0.9)');
    ctx.fillStyle = sandGrad;
    ctx.beginPath();
    ctx.moveTo(0, height);
    ctx.lineTo(0, height - 40);
    ctx.bezierCurveTo(width * 0.3, height - 42, width * 0.7, height - 44, width, height - 40);
    ctx.lineTo(width, height);
    ctx.closePath();
    ctx.fill();
    ctx.restore();
  }

  renderForeground(ctx: CanvasRenderingContext2D, width: number, height: number, totalTime: number): void {
    // 1. Back Layer: background kelp blades + background corals
    this.renderKelpLayer(ctx, height, totalTime, 'back');

    // Left Back:
    this.drawCoral(ctx, this.corals.staghorn, width * 0.06, height - 6, 250, 1.125, Math.sin(totalTime * 0.8) * 0.02);
    this.drawCoral(ctx, this.corals.bushPurple, width * 0.15, height - 8, 190, 1.285);

    // Right Back (anchored to the far right edge):
    this.drawCoral(ctx, this.corals.staghornGold, width * 0.98, height - 6, 270, 1.125, Math.cos(totalTime * 0.7) * 0.02);
    this.drawCoral(ctx, this.corals.branchPink, width * 0.91, height - 8, 230, 1.133);
    this.drawCoral(ctx, this.corals.bushPurple, width * 0.84, height - 8, 190, 1.285);

    // 2. Mid Layer: mid kelp blades + mid corals
    this.renderKelpLayer(ctx, height, totalTime, 'mid');

    // Left Mid:
    this.drawCoral(ctx, this.corals.porous, width * 0.19, height - 10, 215, 0.696, Math.sin(totalTime * 0.8) * 0.015);
    const pulseLeft = 1 + Math.sin(totalTime * 2.0) * 0.02;
    this.drawCoral(ctx, this.corals.anemone, width * 0.22, height - 12, 175, 1.428, 0, pulseLeft);
    this.drawCoral(ctx, this.corals.cups, width * 0.11, height - 10, 160, 1.266);

    // Right Mid (flipped for organic asymmetry):
    this.drawCoral(ctx, this.corals.porous, width * 0.78, height - 8, 210, 0.696, Math.cos(totalTime * 0.85) * 0.015, 1, true);
    this.drawCoral(ctx, this.corals.cups, width * 0.94, height - 10, 215, 1.266, Math.cos(totalTime * 0.75) * 0.02, 1, true);
    this.drawCoral(ctx, this.corals.shelf, width * 0.87, height - 8, 165, 1.307, 0, 1, true);
    const pulseRight = 1 + Math.cos(totalTime * 2.2) * 0.02;
    this.drawCoral(ctx, this.corals.anemone, width * 0.74, height - 10, 155, 1.428, 0, pulseRight);

    // 3. Front Layer:
    this.drawCoral(ctx, this.corals.shelf, width * 0.14, height - 6, 135, 1.307);

    // Central Passage: seashells with soft shadows & starfish on sandy seabed
    this.renderShellWithShadow(ctx, width * 0.38, height - 14, 82, -0.488, false);
    this.renderStarfish(ctx, width * 0.44, height - 38);
    this.renderShellWithShadow(ctx, width * 0.58, height - 12, 108, 0.733, true);
    this.renderShellWithShadow(ctx, width * 0.67, height - 10, 58, 0.314, false);

    // Front kelp blades (reactive to mouse!)
    this.renderKelpLayer(ctx, height, totalTime, 'front');
  }

  private drawCoral(
    ctx: CanvasRenderingContext2D,
    img: HTMLImageElement | null,
    cx: number,
    baseY: number,
    h: number,
    aspect: number,
    rot = 0,
    scale = 1,
    flipX = false
  ): void {
    if (!img || !img.complete || img.naturalWidth === 0) return;
    const w = h * aspect;
    ctx.save();
    ctx.translate(cx, baseY);
    if (rot !== 0) ctx.rotate(rot);
    const sx = flipX ? -scale : scale;
    if (sx !== 1 || scale !== 1) ctx.scale(sx, scale);
    ctx.drawImage(img, -w / 2, -h, w, h);
    ctx.restore();
  }

  private renderShellWithShadow(
    ctx: CanvasRenderingContext2D,
    cx: number,
    cy: number,
    h: number,
    rot: number,
    flipX = false
  ): void {
    const w = h * 0.565;
    ctx.save();
    ctx.translate(cx, cy + 4);
    ctx.rotate(rot * 0.5);
    const rx = w * 0.6;
    const ry = h * 0.18;
    ctx.scale(1, ry / rx);
    const shadowGrad = ctx.createRadialGradient(0, 0, 0, 0, 0, rx);
    shadowGrad.addColorStop(0, 'rgba(2, 6, 12, 0.75)');
    shadowGrad.addColorStop(0.5, 'rgba(2, 6, 12, 0.35)');
    shadowGrad.addColorStop(1, 'rgba(2, 6, 12, 0)');
    ctx.fillStyle = shadowGrad;
    ctx.beginPath();
    ctx.arc(0, 0, rx, 0, Math.PI * 2);
    ctx.fill();
    ctx.restore();

    this.drawCoral(ctx, this.corals.seashell, cx, cy, h, 0.565, rot, 1, flipX);
  }

  private renderKelpLayer(
    ctx: CanvasRenderingContext2D,
    height: number,
    totalTime: number,
    layer: 'back' | 'mid' | 'front'
  ): void {
    if (layer === 'back') ctx.globalAlpha = 0.55;

    for (const k of this.kelps) {
      if (k.layer !== layer) continue;
      const img = this.bladeImgs[k.type];
      if (!img || !img.complete || img.naturalWidth === 0) continue;

      const naturalSway = Math.sin(totalTime * k.swaySpeed + k.swayPhase) * k.swayAmp;
      const totalAngle = naturalSway + k.deflection;
      const baseY = height - k.baseYOffset;

      ctx.save();
      ctx.translate(k.baseX, baseY);
      ctx.rotate(totalAngle);
      ctx.drawImage(img, -k.width / 2, -k.height, k.width, k.height);
      ctx.restore();
    }

    if (layer === 'back') ctx.globalAlpha = 1.0;
  }

  private renderStarfish(ctx: CanvasRenderingContext2D, sx: number, sy: number): void {
    ctx.save();
    ctx.fillStyle = '#FB7185';
    ctx.beginPath();
    for (let s = 0; s < 10; s++) {
      const r = s % 2 === 0 ? 11 : 4.5;
      const a = (s * Math.PI) / 5 - Math.PI / 2;
      const px = sx + Math.cos(a) * r;
      const py = sy + Math.sin(a) * r;
      if (s === 0) ctx.moveTo(px, py);
      else ctx.lineTo(px, py);
    }
    ctx.closePath();
    ctx.fill();

    ctx.fillStyle = '#F43F5E';
    for (let s = 0; s < 5; s++) {
      const a = (s * Math.PI * 2) / 5 - Math.PI / 2;
      ctx.beginPath();
      ctx.arc(sx + Math.cos(a) * 6, sy + Math.sin(a) * 6, 1.2, 0, Math.PI * 2);
      ctx.fill();
    }
    ctx.restore();
  }
}

