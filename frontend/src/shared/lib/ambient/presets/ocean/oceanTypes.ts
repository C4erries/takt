export type FishSpecies = 'clownfish' | 'bluetang' | 'angelfish' | 'yellowtang';

export interface TropicalFish {
  species: FishSpecies;
  x: number;
  y: number;
  baseY: number;
  speedX: number;
  direction: 1 | -1;
  size: number;
  time: number;
  wagSpeed: number;
  swimPhase: number;
  swimSpeed: number;
}

export interface KelpBlade {
  type: 1 | 2 | 3 | 4 | 5;
  baseX: number;
  baseYOffset: number;
  width: number;
  height: number;
  swaySpeed: number;
  swayPhase: number;
  swayAmp: number;
  deflection: number;
  velocity: number;
  layer: 'back' | 'mid' | 'front';
}

export type KelpRibbon = KelpBlade;

export interface CrabState {
  active: boolean;
  x: number;
  y: number;
  direction: 1 | -1;
  speed: number;
  timer: number;
  interval: number;
  legPhase: number;
}

export interface LightShaft {
  baseX: number;
  widthTop: number;
  widthBottom: number;
  slant: number;
  swaySpeed: number;
  swayPhase: number;
  alpha: number;
}

export interface Bubble {
  x: number;
  y: number;
  radius: number;
  speedY: number;
  swaySpeed: number;
  swayAmp: number;
  swayPhase: number;
  alpha: number;
  isForeground: boolean;
}
