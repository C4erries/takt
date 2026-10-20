import type { CrabState } from './oceanTypes';

export class OceanCrab {
  private crab: CrabState = {
    active: false,
    x: 0,
    y: 0,
    direction: 1,
    speed: 38,
    timer: 0,
    interval: 30,
    legPhase: 0,
  };

  init(_width: number, height: number, isVivid: boolean): void {
    this.crab = {
      active: false,
      x: -40,
      y: height - 14,
      direction: 1,
      speed: 36 + Math.random() * 14,
      timer: 0,
      interval: isVivid ? 26 + Math.random() * 8 : 50 + Math.random() * 10,
      legPhase: 0,
    };
  }

  resize(height: number): void {
    this.crab.y = height - 14;
  }

  update(dt: number, width: number, isVivid: boolean): void {
    const clampedDt = Math.min(dt, 0.1);
    this.crab.timer += clampedDt;

    // Trigger crab run
    if (!this.crab.active && this.crab.timer >= this.crab.interval) {
      this.crab.active = true;
      this.crab.direction = Math.random() > 0.5 ? 1 : -1;
      this.crab.x = this.crab.direction === 1 ? -40 : width + 40;
      this.crab.speed = 36 + Math.random() * 16;
    }

    if (this.crab.active) {
      this.crab.x += this.crab.speed * this.crab.direction * clampedDt;
      this.crab.legPhase += clampedDt * 16;

      if (
        (this.crab.direction === 1 && this.crab.x > width + 50) ||
        (this.crab.direction === -1 && this.crab.x < -50)
      ) {
        this.crab.active = false;
        this.crab.timer = 0;
        this.crab.interval = isVivid ? 26 + Math.random() * 8 : 50 + Math.random() * 10;
      }
    }
  }

  render(ctx: CanvasRenderingContext2D): void {
    if (!this.crab.active) return;

    ctx.save();
    ctx.translate(this.crab.x, this.crab.y);
    ctx.scale(this.crab.direction, 1);

    // 1. 6 Articulated Scuttling Legs (3 on left, 3 on right)
    ctx.strokeStyle = '#EA580C';
    ctx.lineWidth = 1.6;
    ctx.lineCap = 'round';
    for (let j = 0; j < 3; j++) {
      const lx = (j - 1) * 5;
      const legWalk = Math.sin(this.crab.legPhase + j * 1.5) * 4;

      // Backward leg
      ctx.beginPath();
      ctx.moveTo(lx, 2);
      ctx.lineTo(lx - 5, 4 + legWalk);
      ctx.lineTo(lx - 8, 8 + legWalk);
      ctx.stroke();

      // Forward leg
      ctx.beginPath();
      ctx.moveTo(lx, 2);
      ctx.lineTo(lx + 5, 4 - legWalk);
      ctx.lineTo(lx + 8, 8 - legWalk);
      ctx.stroke();
    }

    // 2. Volumetric Carapace with Gradient
    const cGrad = ctx.createLinearGradient(0, -7, 0, 7);
    cGrad.addColorStop(0, '#F97316');
    cGrad.addColorStop(0.5, '#EA580C');
    cGrad.addColorStop(1, '#9A3412');
    ctx.fillStyle = cGrad;
    ctx.beginPath();
    ctx.ellipse(0, 0, 10, 6.5, 0, 0, Math.PI * 2);
    ctx.fill();

    // 3. Eyestalks with Dark Eyes and White Glints
    [-3.2, 3.2].forEach((ex) => {
      ctx.strokeStyle = '#EA580C';
      ctx.lineWidth = 1.3;
      ctx.beginPath();
      ctx.moveTo(ex, -2);
      ctx.lineTo(ex, -8);
      ctx.stroke();

      ctx.fillStyle = '#090D16';
      ctx.beginPath();
      ctx.arc(ex, -8, 1.5, 0, Math.PI * 2);
      ctx.fill();

      ctx.fillStyle = '#FFFFFF';
      ctx.beginPath();
      ctx.arc(ex + 0.45, -8.45, 0.55, 0, Math.PI * 2);
      ctx.fill();
    });

    // 4. Raised Chelae (Claws) with Clacking Animation
    const clawClack = Math.sin(this.crab.legPhase * 0.35) * 0.3;
    [[-8.5, -1], [8.5, 1]].forEach(([cx, sign]) => {
      ctx.save();
      ctx.translate(cx, -4);
      ctx.rotate(sign * (0.35 + clawClack));
      ctx.fillStyle = '#FB923C';
      ctx.beginPath();
      ctx.ellipse(0, 0, 4.5, 2.8, 0, 0, Math.PI * 2);
      ctx.fill();

      // Pincer tip
      ctx.fillStyle = '#EA580C';
      ctx.beginPath();
      ctx.moveTo(sign * 2.5, -1);
      ctx.lineTo(sign * 6, -3);
      ctx.lineTo(sign * 4, 1);
      ctx.closePath();
      ctx.fill();
      ctx.restore();
    });

    ctx.restore();
  }
}
