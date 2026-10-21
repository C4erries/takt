export interface TagColorOption {
  name: string;
  hex: string;
}

export const TAG_PALETTE: TagColorOption[] = [
  { name: 'Persimmon', hex: '#E25C3D' },
  { name: 'Terracotta', hex: '#C86D51' },
  { name: 'Amber', hex: '#D97706' },
  { name: 'Bamboo', hex: '#5B8C5A' },
  { name: 'Emerald', hex: '#10B981' },
  { name: 'Ocean', hex: '#0284C7' },
  { name: 'Denim', hex: '#3B82F6' },
  { name: 'Indigo', hex: '#6366F1' },
  { name: 'Lavender', hex: '#8B5CF6' },
  { name: 'Plum', hex: '#A855F7' },
  { name: 'Cocoa', hex: '#8D6E63' },
  { name: 'Sand', hex: '#78716C' },
];

export const DEFAULT_TAG_COLOR = TAG_PALETTE[4].hex; // Emerald
