import type { TropicalFish } from './oceanTypes';

export function renderFish(ctx: CanvasRenderingContext2D, f: TropicalFish): void {
  const tailAngle = Math.sin(f.time * f.wagSpeed) * 0.35;
  const finAngle = Math.cos(f.time * (f.wagSpeed * 1.2)) * 0.25;
  const len = f.size * 1.8;
  const h = f.species === 'angelfish' ? f.size * 0.95 : f.size * 0.65;

  ctx.save();
  ctx.translate(f.x, f.y);
  ctx.scale(f.direction, 1);

  // 1. Caudal (Tail) Fin (drawn behind or seamlessly attached)
  ctx.save();
  if (f.species === 'clownfish') {
    ctx.translate(-len * 0.44, 0);
    ctx.rotate(tailAngle);
    ctx.fillStyle = '#FF7528';
    ctx.strokeStyle = '#18181B';
    ctx.lineWidth = 1.2;
    ctx.beginPath();
    ctx.moveTo(0, 0);
    ctx.quadraticCurveTo(-len * 0.28, -h * 0.72, -len * 0.42, -h * 0.58);
    ctx.quadraticCurveTo(-len * 0.34, 0, -len * 0.42, h * 0.58);
    ctx.quadraticCurveTo(-len * 0.28, h * 0.72, 0, 0);
    ctx.fill();
    ctx.stroke();
  } else if (f.species === 'bluetang') {
    ctx.translate(-len * 0.45, 0);
    ctx.rotate(tailAngle);
    ctx.fillStyle = '#FBBF24';
    ctx.strokeStyle = '#D97706';
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(0, 0);
    ctx.lineTo(-len * 0.44, -h * 0.72);
    ctx.lineTo(-len * 0.36, 0);
    ctx.lineTo(-len * 0.44, h * 0.72);
    ctx.closePath();
    ctx.fill();
    ctx.stroke();
  } else if (f.species === 'angelfish') {
    ctx.translate(-len * 0.36, 0);
    ctx.rotate(tailAngle);
    ctx.fillStyle = '#FDE047';
    ctx.beginPath();
    ctx.moveTo(0, 0);
    ctx.lineTo(-len * 0.38, -h * 0.45);
    ctx.lineTo(-len * 0.38, h * 0.45);
    ctx.closePath();
    ctx.fill();
  } else {
    // yellowtang
    ctx.translate(-len * 0.4, 0);
    ctx.rotate(tailAngle);
    ctx.fillStyle = '#FACC15';
    ctx.strokeStyle = '#CA8A04';
    ctx.lineWidth = 0.8;
    ctx.beginPath();
    ctx.moveTo(0, 0);
    ctx.lineTo(-len * 0.4, -h * 0.65);
    ctx.lineTo(-len * 0.33, 0);
    ctx.lineTo(-len * 0.4, h * 0.65);
    ctx.closePath();
    ctx.fill();
    ctx.stroke();
  }
  ctx.restore();

  // 2. Body with strict ctx.clip() to prevent stripe/pattern bleeding
  ctx.save();
  ctx.beginPath();
  if (f.species === 'clownfish') {
    ctx.ellipse(0, 0, len * 0.5, h, 0, 0, Math.PI * 2);
  } else if (f.species === 'bluetang') {
    ctx.moveTo(len * 0.54, 0);
    ctx.quadraticCurveTo(len * 0.1, -h, -len * 0.45, 0);
    ctx.quadraticCurveTo(len * 0.1, h, len * 0.54, 0);
  } else if (f.species === 'angelfish') {
    ctx.moveTo(len * 0.55, 0);
    ctx.quadraticCurveTo(len * 0.05, -h * 1.15, -len * 0.35, -h * 0.65);
    ctx.quadraticCurveTo(-len * 0.4, 0, -len * 0.35, h * 0.65);
    ctx.quadraticCurveTo(len * 0.05, h * 1.15, len * 0.55, 0);
  } else {
    // yellowtang (elongated snout)
    ctx.moveTo(len * 0.65, 0);
    ctx.quadraticCurveTo(len * 0.1, -h * 1.05, -len * 0.4, 0);
    ctx.quadraticCurveTo(len * 0.1, h * 1.05, len * 0.65, 0);
  }
  ctx.clip();

  // Inner Body Rendering inside the clipped boundary
  if (f.species === 'clownfish') {
    const clownGrad = ctx.createLinearGradient(0, -h, 0, h);
    clownGrad.addColorStop(0, '#FF8533');
    clownGrad.addColorStop(1, '#E65A15');
    ctx.fillStyle = clownGrad;
    ctx.fillRect(-len, -h * 2, len * 2, h * 4);

    // 3 white stripes with soft black borders (head, mid, tail)
    const stripes = [len * 0.28, -len * 0.03, -len * 0.32];
    for (const sx of stripes) {
      // Soft black boundary
      ctx.strokeStyle = 'rgba(24, 24, 27, 0.75)';
      ctx.lineWidth = len * 0.15;
      ctx.beginPath();
      ctx.moveTo(sx, -h * 1.5);
      ctx.quadraticCurveTo(sx + 3, 0, sx, h * 1.5);
      ctx.stroke();

      // Soft white core
      ctx.strokeStyle = 'rgba(255, 255, 255, 0.94)';
      ctx.lineWidth = len * 0.10;
      ctx.beginPath();
      ctx.moveTo(sx, -h * 1.5);
      ctx.quadraticCurveTo(sx + 3, 0, sx, h * 1.5);
      ctx.stroke();
    }
  } else if (f.species === 'bluetang') {
    const blueGrad = ctx.createLinearGradient(-len * 0.5, 0, len * 0.5, 0);
    blueGrad.addColorStop(0, '#1D4ED8');
    blueGrad.addColorStop(1, '#2563EB');
    ctx.fillStyle = blueGrad;
    ctx.fillRect(-len, -h * 2, len * 2, h * 4);

    // Signature black curved side loop
    ctx.fillStyle = 'rgba(15, 23, 42, 0.85)';
    ctx.beginPath();
    ctx.ellipse(-len * 0.08, 0, len * 0.25, h * 0.5, 0.2, 0, Math.PI * 2);
    ctx.fill();

    // Inner blue core inside black loop
    ctx.fillStyle = '#2563EB';
    ctx.beginPath();
    ctx.ellipse(-len * 0.06, 0, len * 0.12, h * 0.25, 0, 0, Math.PI * 2);
    ctx.fill();
  } else if (f.species === 'angelfish') {
    const angelGrad = ctx.createLinearGradient(0, -h, 0, h);
    angelGrad.addColorStop(0, '#3730A3');
    angelGrad.addColorStop(1, '#1E1B4B');
    ctx.fillStyle = angelGrad;
    ctx.fillRect(-len, -h * 2, len * 2, h * 4);

    // Curved Bezier gold stripes with soft glow
    ctx.save();
    ctx.strokeStyle = 'rgba(253, 224, 71, 0.72)';
    ctx.lineWidth = 1.1;
    ctx.shadowColor = 'rgba(253, 224, 71, 0.35)';
    ctx.shadowBlur = 2;
    const stripeOffsets = [-len * 0.26, -len * 0.12, len * 0.02, len * 0.16];
    for (const sx of stripeOffsets) {
      ctx.beginPath();
      ctx.moveTo(sx, -h * 1.15);
      ctx.quadraticCurveTo(sx + len * 0.14, 0, sx - len * 0.06, h * 1.15);
      ctx.stroke();
    }
    ctx.restore();

    // Soft eye mask with neon cyan border
    ctx.fillStyle = 'rgba(15, 23, 42, 0.7)';
    ctx.strokeStyle = 'rgba(6, 182, 212, 0.8)';
    ctx.lineWidth = 0.95;
    ctx.beginPath();
    ctx.ellipse(len * 0.28, -h * 0.15, len * 0.09, h * 0.38, 0.2, 0, Math.PI * 2);
    ctx.fill();
    ctx.stroke();
  } else {
    // yellowtang
    const yellowGrad = ctx.createLinearGradient(0, -h, 0, h);
    yellowGrad.addColorStop(0, '#FACC15');
    yellowGrad.addColorStop(1, '#EAB308');
    ctx.fillStyle = yellowGrad;
    ctx.fillRect(-len, -h * 2, len * 2, h * 4);

    // White spine dot at tail base
    ctx.fillStyle = '#FEF08A';
    ctx.strokeStyle = '#CA8A04';
    ctx.lineWidth = 0.8;
    ctx.beginPath();
    ctx.arc(-len * 0.34, 0, len * 0.06, 0, Math.PI * 2);
    ctx.fill();
    ctx.stroke();
  }
  ctx.restore(); // Ends body clipping

  // 3. Expressive Eye
  const eyeX = f.species === 'yellowtang' ? len * 0.4 : len * 0.32;
  const eyeY = -h * 0.2;
  const eyeR = f.size * 0.15;
  ctx.fillStyle = '#FFFFFF';
  ctx.beginPath();
  ctx.arc(eyeX, eyeY, eyeR, 0, Math.PI * 2);
  ctx.fill();
  ctx.fillStyle = '#090D16';
  ctx.beginPath();
  ctx.arc(eyeX + eyeR * 0.35, eyeY, eyeR * 0.55, 0, Math.PI * 2);
  ctx.fill();
  ctx.fillStyle = '#FFFFFF';
  ctx.beginPath();
  ctx.arc(eyeX + eyeR * 0.45, eyeY - eyeR * 0.2, eyeR * 0.22, 0, Math.PI * 2);
  ctx.fill();

  // 4. Fluttering Pectoral Fin on Side
  ctx.save();
  ctx.translate(0, h * 0.12);
  ctx.rotate(finAngle);
  ctx.fillStyle =
    f.species === 'clownfish'
      ? '#FF7528'
      : f.species === 'bluetang'
      ? '#1D4ED8'
      : f.species === 'angelfish'
      ? '#4338CA'
      : '#FACC15';
  ctx.globalAlpha = 0.88;
  ctx.beginPath();
  ctx.ellipse(-len * 0.04, 0, len * 0.19, h * 0.3, 0.4, 0, Math.PI * 2);
  ctx.fill();
  ctx.restore();

  ctx.restore();
}
