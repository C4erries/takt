import React, { useEffect, useRef, useState } from 'react';

export interface OceanCrabDOMProps {
  isInteractive?: boolean;
  className?: string;
}

type CrabState = 'idle' | 'walking' | 'spooked' | 'fleeing';

export const OceanCrabDOM: React.FC<OceanCrabDOMProps> = ({
  isInteractive = true,
  className = '',
}) => {
  const [status, setStatus] = useState<CrabState>('idle');
  const [facing, setFacing] = useState<1 | -1>(1);
  const containerRef = useRef<HTMLDivElement>(null);

  const stateRef = useRef<{
    status: CrabState;
    xPercent: number;
    targetXPercent: number;
    facing: 1 | -1;
    lastTime: number;
    idleTimer: number;
    spookedTimer: number;
    isInteractive: boolean;
  }>({
    status: 'idle',
    xPercent: 48,
    targetXPercent: 48,
    facing: 1,
    lastTime: performance.now(),
    idleTimer: 3.5,
    spookedTimer: 0,
    isInteractive,
  });

  stateRef.current.isInteractive = isInteractive;

  useEffect(() => {
    let rafId: number;

    const pickNewPatrolTarget = () => {
      const s = stateRef.current;
      // Pick target between 20% and 80%
      const newTarget = 20 + Math.random() * 60;
      s.targetXPercent = newTarget;
      s.facing = newTarget >= s.xPercent ? 1 : -1;
      s.status = 'walking';
      setStatus('walking');
      setFacing(s.facing);
    };

    const loop = (currentTime: number) => {
      rafId = requestAnimationFrame(loop);
      const s = stateRef.current;
      const dt = Math.min((currentTime - s.lastTime) / 1000, 0.1);
      s.lastTime = currentTime;

      if (s.status === 'idle') {
        s.idleTimer -= dt;
        if (s.idleTimer <= 0) {
          s.idleTimer = 3.0 + Math.random() * 3.5;
          pickNewPatrolTarget();
        }
      } else if (s.status === 'walking') {
        const diff = s.targetXPercent - s.xPercent;
        const speed = 7.5; // percent per second
        if (Math.abs(diff) < 0.6) {
          s.xPercent = s.targetXPercent;
          s.status = 'idle';
          s.idleTimer = 3.0 + Math.random() * 3.5;
          setStatus('idle');
        } else {
          const move = Math.sign(diff) * speed * dt;
          s.xPercent += move;
        }
      } else if (s.status === 'fleeing') {
        const diff = s.targetXPercent - s.xPercent;
        const speed = 18.0; // fast sprint
        if (Math.abs(diff) < 0.8) {
          s.xPercent = s.targetXPercent;
          s.status = 'spooked';
          s.spookedTimer = 2.0;
          setStatus('spooked');
        } else {
          s.xPercent += Math.sign(diff) * speed * dt;
        }
      } else if (s.status === 'spooked') {
        s.spookedTimer -= dt;
        if (s.spookedTimer <= 0) {
          s.status = 'idle';
          s.idleTimer = 2.5;
          setStatus('idle');
        }
      }

      if (containerRef.current) {
        containerRef.current.style.left = `${s.xPercent}%`;
        containerRef.current.style.transform = `translateX(-50%) scaleX(${s.facing})`;
      }
    };

    rafId = requestAnimationFrame(loop);
    return () => cancelAnimationFrame(rafId);
  }, []);

  // Cursor reactivity (mouse movement listener)
  useEffect(() => {
    const handlePointerMove = (e: MouseEvent) => {
      const s = stateRef.current;
      if (!s.isInteractive || typeof window === 'undefined') return;

      const crabScreenX = (s.xPercent / 100) * window.innerWidth;
      const crabScreenY = window.innerHeight - 36;
      const dist = Math.hypot(e.clientX - crabScreenX, e.clientY - crabScreenY);

      if (dist < 90) {
        // Flee away from mouse immediately
        const escapeDir: 1 | -1 = e.clientX > crabScreenX ? -1 : 1;
        s.facing = escapeDir;
        s.targetXPercent = escapeDir === -1 ? 18 : 82;
        s.status = 'fleeing';
        setStatus('fleeing');
        setFacing(escapeDir);
      } else if (dist < 200) {
        // Threat alert: raise claws in defense
        if (s.status !== 'fleeing') {
          s.status = 'spooked';
          s.spookedTimer = 2.0;
          setStatus('spooked');
        }
      }
    };

    window.addEventListener('mousemove', handlePointerMove, { passive: true });
    return () => window.removeEventListener('mousemove', handlePointerMove);
  }, []);

  return (
    <div
      ref={containerRef}
      className={`ocean-crab-wrapper is-${status} ${className}`}
      style={{
        position: 'absolute',
        bottom: '12px',
        left: '48%',
        transform: `translateX(-50%) scaleX(${facing})`,
        pointerEvents: 'auto',
        width: '96px',
        height: '64px',
        zIndex: 1,
      }}
      aria-hidden="true"
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 240 160"
        className="ocean-crab"
      >
        <defs>
          <radialGradient id="crab-ground-shadow" cx="50%" cy="50%" r="50%">
            <stop offset="0%" stopColor="#02060B" stopOpacity="0.85" />
            <stop offset="65%" stopColor="#040C14" stopOpacity="0.4" />
            <stop offset="100%" stopColor="#040C14" stopOpacity="0" />
          </radialGradient>
          <linearGradient id="crab-shell-main" x1="0.3" y1="0" x2="0.6" y2="1">
            <stop offset="0%" stopColor="#FDBA74" />
            <stop offset="18%" stopColor="#F97316" />
            <stop offset="58%" stopColor="#BE123C" />
            <stop offset="100%" stopColor="#3B0726" />
          </linearGradient>
          <linearGradient id="crab-shell-rim" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#FEF08A" stopOpacity="0.75" />
            <stop offset="50%" stopColor="#FB923C" stopOpacity="0.2" />
            <stop offset="100%" stopColor="#BE123C" stopOpacity="0" />
          </linearGradient>
          <linearGradient id="crab-claw-grad" x1="0.2" y1="0" x2="0.8" y2="1">
            <stop offset="0%" stopColor="#FED7AA" />
            <stop offset="25%" stopColor="#FB7185" />
            <stop offset="70%" stopColor="#9F1239" />
            <stop offset="100%" stopColor="#2E0518" />
          </linearGradient>
          <linearGradient id="crab-leg-front" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#FB923C" />
            <stop offset="45%" stopColor="#B91C1C" />
            <stop offset="85%" stopColor="#450A0A" />
            <stop offset="100%" stopColor="#1E0510" />
          </linearGradient>
          <linearGradient id="crab-leg-back" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#9F1239" />
            <stop offset="60%" stopColor="#4C0519" />
            <stop offset="100%" stopColor="#14030B" />
          </linearGradient>
          <radialGradient id="crab-eye-ball" cx="35%" cy="35%" r="65%">
            <stop offset="0%" stopColor="#38BDF8" />
            <stop offset="30%" stopColor="#0F172A" />
            <stop offset="100%" stopColor="#020617" />
          </radialGradient>
        </defs>

        {/* 1. Ground shadow */}
        <ellipse className="crab-shadow" cx="120" cy="142" rx="82" ry="13" fill="url(#crab-ground-shadow)" />

        {/* 2. Left claw */}
        <g className="crab-claw-l" style={{ transformOrigin: '78px 72px' }}>
          <path d="M82,76 C62,68 44,54 38,38 C46,32 58,38 66,48 C74,58 84,64 88,70 Z" fill="url(#crab-leg-front)" />
          <path d="M44,44 C26,42 12,28 16,14 C20,4 36,12 42,22 C36,24 28,20 26,24 C24,28 34,34 48,32 C56,30 60,40 44,44 Z" fill="url(#crab-claw-grad)" />
          <g className="crab-pincer-l" style={{ transformOrigin: '48px 30px' }}>
            <path d="M46,32 C54,18 46,4 32,4 C22,4 22,12 28,14 C34,16 40,22 38,28 Z" fill="url(#crab-claw-grad)" />
            <path d="M32,5 C42,5 48,14 45,24 C43,16 38,10 30,9 Z" fill="#FEF08A" opacity={0.5} />
          </g>
          <path d="M18,16 C22,26 34,36 46,38 C34,38 20,30 18,16 Z" fill="#FFE4E6" opacity={0.45} />
        </g>

        {/* 3. Right claw */}
        <g className="crab-claw-r" style={{ transformOrigin: '162px 72px' }}>
          <path d="M158,76 C178,68 196,54 202,38 C194,32 182,38 174,48 C166,58 156,64 152,70 Z" fill="url(#crab-leg-front)" />
          <path d="M196,44 C214,42 228,28 224,14 C220,4 204,12 198,22 C204,24 212,20 214,24 C216,28 206,34 192,32 C184,30 180,40 196,44 Z" fill="url(#crab-claw-grad)" />
          <g className="crab-pincer-r" style={{ transformOrigin: '192px 30px' }}>
            <path d="M194,32 C186,18 194,4 208,4 C218,4 218,12 212,14 C206,16 200,22 202,28 Z" fill="url(#crab-claw-grad)" />
            <path d="M208,5 C198,5 192,14 195,24 C197,16 202,10 210,9 Z" fill="#FEF08A" opacity={0.5} />
          </g>
          <path d="M222,16 C218,26 206,36 194,38 C206,38 220,30 222,16 Z" fill="#FFE4E6" opacity={0.45} />
        </g>

        {/* 4. Walking legs set A & B */}
        <g className="crab-leg leg-set-a leg-l3" style={{ transformOrigin: '76px 98px' }}>
          <path d="M78,96 C54,96 34,106 24,126 C20,134 24,136 28,128 C38,112 56,104 78,104 Z" fill="url(#crab-leg-back)" />
        </g>
        <g className="crab-leg leg-set-b leg-r3" style={{ transformOrigin: '164px 98px' }}>
          <path d="M162,96 C186,96 206,106 216,126 C220,134 216,136 212,128 C202,112 184,104 162,104 Z" fill="url(#crab-leg-back)" />
        </g>

        <g className="crab-leg leg-set-b leg-l2" style={{ transformOrigin: '72px 92px' }}>
          <path d="M76,88 C46,84 20,98 10,128 C8,136 14,138 18,128 C28,104 48,94 76,98 Z" fill="url(#crab-leg-front)" />
          <path d="M74,89 C48,86 26,96 14,120 C28,98 48,90 74,92 Z" fill="#FDE047" opacity={0.35} />
        </g>
        <g className="crab-leg leg-set-a leg-r2" style={{ transformOrigin: '168px 92px' }}>
          <path d="M164,88 C194,84 220,98 230,128 C232,136 226,138 222,128 C212,104 192,94 164,98 Z" fill="url(#crab-leg-front)" />
          <path d="M166,89 C192,86 214,96 226,120 C212,98 192,90 166,92 Z" fill="#FDE047" opacity={0.35} />
        </g>

        <g className="crab-leg leg-set-a leg-l1" style={{ transformOrigin: '78px 86px' }}>
          <path d="M82,82 C54,76 30,92 22,132 C20,140 26,142 30,132 C38,102 56,88 82,92 Z" fill="url(#crab-leg-front)" />
          <path d="M80,83 C56,78 36,90 26,120 C38,94 56,84 80,86 Z" fill="#FEF08A" opacity={0.45} />
        </g>
        <g className="crab-leg leg-set-b leg-r1" style={{ transformOrigin: '162px 86px' }}>
          <path d="M158,82 C186,76 210,92 218,132 C220,140 214,142 210,132 C202,102 184,88 158,92 Z" fill="url(#crab-leg-front)" />
          <path d="M160,83 C184,78 204,90 214,120 C202,94 184,84 160,86 Z" fill="#FEF08A" opacity={0.45} />
        </g>

        {/* 5. Main carapace / body group */}
        <g className="crab-body-group" style={{ transformOrigin: '120px 88px' }}>
          <g className="crab-eyes">
            <g className="crab-eye-l" style={{ transformOrigin: '102px 62px' }}>
              <path d="M98,64 C96,52 92,42 94,38 C98,36 104,44 104,62 Z" fill="url(#crab-leg-front)" />
              <circle cx="95" cy="38" r="7.5" fill="url(#crab-eye-ball)" stroke="#FB7185" strokeWidth={1.5} />
              <circle cx="92.5" cy="35.5" r="2.2" fill="#FFFFFF" opacity={0.9} />
              <circle cx="97.5" cy="40.5" r="1" fill="#38BDF8" opacity={0.7} />
            </g>
            <g className="crab-eye-r" style={{ transformOrigin: '138px 62px' }}>
              <path d="M142,64 C144,52 148,42 146,38 C142,36 136,44 136,62 Z" fill="url(#crab-leg-front)" />
              <circle cx="145" cy="38" r="7.5" fill="url(#crab-eye-ball)" stroke="#FB7185" strokeWidth={1.5} />
              <circle cx="142.5" cy="35.5" r="2.2" fill="#FFFFFF" opacity={0.9} />
              <circle cx="147.5" cy="40.5" r="1" fill="#38BDF8" opacity={0.7} />
            </g>
          </g>

          <path d="M74,94 C86,112 154,112 166,94 C156,104 84,104 74,94 Z" fill="#1A0410" />
          <path d="M62,78 C58,72 68,66 72,62 C74,56 84,54 92,54 C104,50 136,50 148,54 C156,54 166,56 168,62 C172,66 182,72 178,78 C174,96 152,108 120,108 C88,108 66,96 62,78 Z" fill="url(#crab-shell-main)" />
          <path d="M72,64 C86,52 154,52 168,64 C152,56 88,56 72,64 Z" fill="url(#crab-shell-rim)" />

          <path d="M96,68 C108,76 132,76 144,68 C132,73 108,73 96,68 Z" fill="#4C0519" opacity={0.45} />
          <path d="M104,82 C112,86 128,86 136,82 C128,84 112,84 104,82 Z" fill="#4C0519" opacity={0.4} />

          <g fill="#FED7AA" opacity={0.45}>
            <ellipse cx="86" cy="70" rx="4" ry="2.5" transform="rotate(-15 86 70)" />
            <ellipse cx="76" cy="78" rx="3" ry="2" transform="rotate(-25 76 78)" />
            <ellipse cx="96" cy="62" rx="3" ry="1.8" />
            <ellipse cx="154" cy="70" rx="4" ry="2.5" transform="rotate(15 154 70)" />
            <ellipse cx="164" cy="78" rx="3" ry="2" transform="rotate(25 164 78)" />
            <ellipse cx="144" cy="62" rx="3" ry="1.8" />
          </g>

          <path d="M108,92 Q120,96 132,92" stroke="#2E0518" strokeWidth={2} strokeLinecap="round" fill="none" opacity={0.7} />
          <path d="M114,96 L116,100 M126,96 L124,100" stroke="#FDA4AF" strokeWidth={1.8} strokeLinecap="round" opacity={0.6} />
        </g>
      </svg>
    </div>
  );
};
