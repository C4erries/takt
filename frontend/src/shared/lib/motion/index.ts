// Centralized Motion & Kinetics API for Takt (ADR-0005, Sprint 0.7.3)

// Re-export core motion components and types from motion/react
export { motion, AnimatePresence, m } from 'motion/react';
export type { Transition, Variants, MotionProps } from 'motion/react';

// Springs physics profiles
export {
  springTactile,
  springSheet,
  springLayout,
  tapScale,
  checkboxPopScale,
} from './springs';

// Accessibility & Reduced Motion utilities
export {
  defaultFadeFallback,
  getPrefersReducedMotion,
  useReducedMotionPreference,
  withReducedMotion,
} from './reducedMotion';

// Unified motion variants presets
export {
  drawerVariants,
  getDrawerVariants,
  modalVariants,
  getModalVariants,
  backdropVariants,
  cardExitVariants,
  getCardExitVariants,
  strikeThroughVariants,
  getStrikeThroughVariants,
  columnExpandVariants,
} from './variants';

