import type { Transition } from 'motion/react';

/**
 * springTactile: Fast, crisp spring physics for buttons, checkboxes, micro-pops.
 * stiffness: 420, damping: 30, mass: 0.8
 * Reference: ADR-0005, D-073-1
 */
export const springTactile: Transition = {
  type: 'spring',
  stiffness: 420,
  damping: 30,
  mass: 0.8,
};

/**
 * springSheet: Smooth, damped spring physics for large surfaces (Drawer, Modals).
 * stiffness: 300, damping: 32, mass: 1
 * Reference: ADR-0005, D-073-1, D-073-4
 */
export const springSheet: Transition = {
  type: 'spring',
  stiffness: 300,
  damping: 32,
  mass: 1,
};

/**
 * springLayout: Natural, fluid spring physics for list reordering and column expansion.
 * stiffness: 340, damping: 28
 * Reference: ADR-0005, D-073-1, D-073-5
 */
export const springLayout: Transition = {
  type: 'spring',
  stiffness: 340,
  damping: 28,
};

/**
 * Button tap micro-interaction preset (D-073-5, ADR-0005)
 * Applied to buttons, chips, tabs via whileTap={{ scale: 0.97 }}
 */
export const tapScale = {
  scale: 0.97,
} as const;

/**
 * Checkbox micro-pop keyframes for task completion (D-073-3, ADR-0005)
 * [1 -> 0.88 -> 1.08 -> 1]
 */
export const checkboxPopScale = [1, 0.88, 1.08, 1] as const;

