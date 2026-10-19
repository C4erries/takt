import type { Variants, TargetAndTransition } from 'motion/react';
import { springSheet, springLayout } from './springs';
import { defaultFadeFallback, getPrefersReducedMotion } from './reducedMotion';

/**
 * drawerVariants: Slide-in drawer animation from the right (x: '100%' -> x: 0)
 * with adaptive dampening fallback for prefers-reduced-motion.
 * Reference: ADR-0005, D-073-1, D-073-4
 */
export const drawerVariants: Variants = {
  hidden: (isReduced: boolean = getPrefersReducedMotion()): TargetAndTransition => ({
    x: isReduced ? 0 : '100%',
    opacity: isReduced ? 0 : 1,
    transition: isReduced ? defaultFadeFallback : springSheet,
  }),
  visible: (isReduced: boolean = getPrefersReducedMotion()): TargetAndTransition => ({
    x: 0,
    opacity: 1,
    transition: isReduced ? defaultFadeFallback : springSheet,
  }),
  exit: (isReduced: boolean = getPrefersReducedMotion()): TargetAndTransition => ({
    x: isReduced ? 0 : '100%',
    opacity: isReduced ? 0 : 1,
    transition: isReduced ? defaultFadeFallback : springSheet,
  }),
  closed: (isReduced: boolean = getPrefersReducedMotion()): TargetAndTransition => ({
    x: isReduced ? 0 : '100%',
    opacity: isReduced ? 0 : 1,
    transition: isReduced ? defaultFadeFallback : springSheet,
  }),
  open: (isReduced: boolean = getPrefersReducedMotion()): TargetAndTransition => ({
    x: 0,
    opacity: 1,
    transition: isReduced ? defaultFadeFallback : springSheet,
  }),
};

export function getDrawerVariants(isReduced: boolean = getPrefersReducedMotion()): Variants {
  return {
    hidden: (drawerVariants.hidden as (r: boolean) => TargetAndTransition)(isReduced),
    visible: (drawerVariants.visible as (r: boolean) => TargetAndTransition)(isReduced),
    exit: (drawerVariants.exit as (r: boolean) => TargetAndTransition)(isReduced),
  };
}

/**
 * modalVariants: Modal entrance scaling & fade (scale: 0.96, opacity: 0 -> scale: 1, opacity: 1)
 * with reduced-motion dampening (suppresses scaling).
 * Reference: ADR-0005, D-073-1, D-073-4
 */
export const modalVariants: Variants = {
  hidden: (isReduced: boolean = getPrefersReducedMotion()): TargetAndTransition => ({
    scale: isReduced ? 1 : 0.96,
    opacity: 0,
    transition: isReduced ? defaultFadeFallback : springSheet,
  }),
  visible: (isReduced: boolean = getPrefersReducedMotion()): TargetAndTransition => ({
    scale: 1,
    opacity: 1,
    transition: isReduced ? defaultFadeFallback : springSheet,
  }),
  exit: (isReduced: boolean = getPrefersReducedMotion()): TargetAndTransition => ({
    scale: isReduced ? 1 : 0.96,
    opacity: 0,
    transition: isReduced ? defaultFadeFallback : { duration: 0.15, ease: 'easeIn' },
  }),
  closed: (isReduced: boolean = getPrefersReducedMotion()): TargetAndTransition => ({
    scale: isReduced ? 1 : 0.96,
    opacity: 0,
    transition: isReduced ? defaultFadeFallback : springSheet,
  }),
  open: (isReduced: boolean = getPrefersReducedMotion()): TargetAndTransition => ({
    scale: 1,
    opacity: 1,
    transition: isReduced ? defaultFadeFallback : springSheet,
  }),
};

export function getModalVariants(isReduced: boolean = getPrefersReducedMotion()): Variants {
  return {
    hidden: (modalVariants.hidden as (r: boolean) => TargetAndTransition)(isReduced),
    visible: (modalVariants.visible as (r: boolean) => TargetAndTransition)(isReduced),
    exit: (modalVariants.exit as (r: boolean) => TargetAndTransition)(isReduced),
  };
}

/**
 * backdropVariants: Soft overlay appearance without heavy blur (opacity: 0 -> opacity: 1, duration: 0.2).
 * Reference: ADR-0005, D-073-4
 */
export const backdropVariants: Variants = {
  hidden: {
    opacity: 0,
    transition: { duration: 0.2, ease: 'easeOut' },
  },
  visible: {
    opacity: 1,
    transition: { duration: 0.2, ease: 'easeOut' },
  },
  exit: {
    opacity: 0,
    transition: { duration: 0.2, ease: 'easeIn' },
  },
  closed: {
    opacity: 0,
    transition: { duration: 0.2, ease: 'easeOut' },
  },
  open: {
    opacity: 1,
    transition: { duration: 0.2, ease: 'easeOut' },
  },
};

/**
 * cardExitVariants: Smooth collapse of card upon completion/deletion
 * (opacity: 0, height: 0, marginBottom: 0).
 * Reference: ADR-0005, D-073-1, D-073-3
 */
export const cardExitVariants: Variants = {
  initial: {
    opacity: 1,
    height: 'auto',
  },
  visible: {
    opacity: 1,
    height: 'auto',
  },
  animate: {
    opacity: 1,
    height: 'auto',
  },
  exit: (isReduced: boolean = getPrefersReducedMotion()): TargetAndTransition => ({
    opacity: 0,
    height: 0,
    marginBottom: 0,
    overflow: 'hidden',
    transition: isReduced
      ? defaultFadeFallback
      : {
          opacity: { duration: 0.15, ease: 'easeOut' },
          height: springLayout,
          marginBottom: springLayout,
        },
  }),
  collapsed: (isReduced: boolean = getPrefersReducedMotion()): TargetAndTransition => ({
    opacity: 0,
    height: 0,
    marginBottom: 0,
    overflow: 'hidden',
    transition: isReduced
      ? defaultFadeFallback
      : {
          opacity: { duration: 0.15, ease: 'easeOut' },
          height: springLayout,
          marginBottom: springLayout,
        },
  }),
};

export function getCardExitVariants(isReduced: boolean = getPrefersReducedMotion()): Variants {
  return {
    initial: { opacity: 1, height: 'auto' },
    visible: { opacity: 1, height: 'auto' },
    exit: (cardExitVariants.exit as (r: boolean) => TargetAndTransition)(isReduced),
  };
}

/**
 * strikeThroughVariants: Animated progressive ink strike-through line
 * drawn from left to right (scaleX: 0 -> scaleX: 1, originX: 0, duration: 0.16).
 * Reference: ADR-0005, D-073-1, D-073-3
 */
export const strikeThroughVariants: Variants = {
  unchecked: {
    scaleX: 0,
    originX: 0,
    opacity: 0,
    transition: { duration: 0.1 },
  },
  checked: (isReduced: boolean = getPrefersReducedMotion()): TargetAndTransition => ({
    scaleX: 1,
    originX: 0,
    opacity: 1,
    transition: isReduced
      ? { duration: 0.05, ease: 'linear' }
      : { duration: 0.16, ease: [0.22, 1, 0.36, 1] },
  }),
  hidden: {
    scaleX: 0,
    originX: 0,
    opacity: 0,
    transition: { duration: 0.1 },
  },
  visible: (isReduced: boolean = getPrefersReducedMotion()): TargetAndTransition => ({
    scaleX: 1,
    originX: 0,
    opacity: 1,
    transition: isReduced
      ? { duration: 0.05, ease: 'linear' }
      : { duration: 0.16, ease: [0.22, 1, 0.36, 1] },
  }),
};

export function getStrikeThroughVariants(isReduced: boolean = getPrefersReducedMotion()): Variants {
  return {
    unchecked: { scaleX: 0, originX: 0, opacity: 0, transition: { duration: 0.1 } },
    checked: (strikeThroughVariants.checked as (r: boolean) => TargetAndTransition)(isReduced),
  };
}

/**
 * columnExpandVariants: Smooth column width expansion for Kanban board (D-073-5).
 */
export const columnExpandVariants: Variants = {
  hidden: (isReduced: boolean = getPrefersReducedMotion()): TargetAndTransition => ({
    width: isReduced ? 280 : 0,
    opacity: 0,
    overflow: 'hidden',
    transition: isReduced ? defaultFadeFallback : springLayout,
  }),
  visible: {
    width: 280,
    opacity: 1,
    overflow: 'visible',
    transition: springLayout,
  },
  exit: (isReduced: boolean = getPrefersReducedMotion()): TargetAndTransition => ({
    width: isReduced ? 280 : 0,
    opacity: 0,
    overflow: 'hidden',
    transition: isReduced ? defaultFadeFallback : springLayout,
  }),
};
