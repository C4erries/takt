import { useState, useEffect } from 'react';
import type { Transition } from 'motion/react';

/**
 * Standard delicate fade fallback for reduced motion preference.
 * Replaces spatial translations (x/y) and scaling with a soft 100ms fade.
 * Reference: ADR-0005, D-073-2
 */
export const defaultFadeFallback: Transition = {
  duration: 0.1,
  ease: 'easeOut',
};

/**
 * Check if the user agent prefers reduced motion.
 * SSR-safe helper for media query inspection.
 */
export function getPrefersReducedMotion(): boolean {
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
    return false;
  }
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

/**
 * React hook listening to @media (prefers-reduced-motion: reduce).
 * Dynamically responds to OS accessibility settings changes with proper cleanup.
 */
export function useReducedMotionPreference(): boolean {
  const [prefersReduced, setPrefersReduced] = useState<boolean>(() => getPrefersReducedMotion());

  useEffect(() => {
    if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
      return;
    }

    const mql = window.matchMedia('(prefers-reduced-motion: reduce)');
    const onChange = (event: MediaQueryListEvent) => {
      setPrefersReduced(event.matches);
    };

    setPrefersReduced(mql.matches);

    if (typeof mql.addEventListener === 'function') {
      mql.addEventListener('change', onChange);
      return () => {
        mql.removeEventListener('change', onChange);
      };
    } else if (typeof (mql as unknown as { addListener: (cb: (e: MediaQueryListEvent) => void) => void }).addListener === 'function') {
      (mql as unknown as { addListener: (cb: (e: MediaQueryListEvent) => void) => void }).addListener(onChange);
      return () => {
        (mql as unknown as { removeListener: (cb: (e: MediaQueryListEvent) => void) => void }).removeListener(onChange);
      };
    }
  }, []);

  return prefersReduced;
}

/**
 * Adaptive dampening helper: replaces spring physics with a gentle fade fallback
 * if reduced motion preference is active.
 *
 * Usage:
 *   withReducedMotion(springSheet)
 *   withReducedMotion(springSheet, isReduced)
 *   withReducedMotion(springSheet, customFallback)
 *   withReducedMotion(springSheet, customFallback, isReduced)
 */
export function withReducedMotion(
  springTransition: Transition,
  isReduced: boolean
): Transition;
export function withReducedMotion(
  springTransition: Transition,
  fadeFallback?: Transition,
  isReduced?: boolean
): Transition;
export function withReducedMotion(
  springTransition: Transition,
  fadeFallbackOrIsReduced?: Transition | boolean,
  isReduced?: boolean
): Transition {
  let fallback: Transition = defaultFadeFallback;
  let active: boolean;

  if (typeof fadeFallbackOrIsReduced === 'boolean') {
    active = fadeFallbackOrIsReduced;
  } else {
    if (fadeFallbackOrIsReduced) {
      fallback = fadeFallbackOrIsReduced;
    }
    active = isReduced !== undefined ? isReduced : getPrefersReducedMotion();
  }

  return active ? fallback : springTransition;
}

