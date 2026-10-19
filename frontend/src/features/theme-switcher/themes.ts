// Takt Editorial Tech — 15 Themes Catalog (ADR-0006 & ADR-0007, Sprint 0.7.4)

export type ThemeId =
  | 'kyoto'
  | 'olive'
  | 'claude'
  | 'ochre'
  | 'bauhaus'
  | 'marigold'
  | 'nordic'
  | 'oxford'
  | 'fig'
  | 'alba'
  | 'espresso'
  | 'cypress'
  | 'kyoto-snow'
  | 'tsukimi'
  | 'celestial'
  | 'frost-pine';

export type ThemeCategory = 'light' | 'dark';

export interface ThemePreview {
  canvas: string;
  surface: string;
  accent: string;
  text: string;
  border: string;
}

export interface ThemeDefinition {
  id: ThemeId;
  name: string;
  category: ThemeCategory;
  description: string;
  preview: ThemePreview;
  ambientPreset?: 'snow' | 'autumn' | 'celestial' | 'blizzard';
}

export const THEMES: readonly ThemeDefinition[] = [
  // 10 Light Static Themes (Светлая бумага)
  {
    id: 'kyoto',
    name: 'Kyoto Matcha & Hinoki',
    category: 'light',
    description: 'Сухой бамбук, чистый чай матча и волокна бумаги васи',
    preview: {
      canvas: '#E8E5DA',
      surface: '#F7F5EE',
      accent: '#566B44',
      text: '#21241E',
      border: '#DCD7C8',
    },
  },
  {
    id: 'olive',
    name: 'Olive & Parchment',
    category: 'light',
    description: 'Оливковая хвоя, мягкий лен и состаренный пергамент',
    preview: {
      canvas: '#ECE8DF',
      surface: '#F6F4EE',
      accent: '#6E7A4E',
      text: '#23221F',
      border: '#DFDBD0',
    },
  },
  {
    id: 'claude',
    name: 'Anthropic Pampas & Clay',
    category: 'light',
    description: 'Теплая глина терракоты и шелест степных пампасов',
    preview: {
      canvas: '#EFECE6',
      surface: '#FAF8F5',
      accent: '#C96442',
      text: '#1F1E1D',
      border: '#E3DEC3',
    },
  },
  {
    id: 'ochre',
    name: 'Ochre & Beeswax',
    category: 'light',
    description: 'Золотистая охра, натуральный пчелиный воск и мед',
    preview: {
      canvas: '#EFECE4',
      surface: '#FAF8F2',
      accent: '#B8860B',
      text: '#211E18',
      border: '#E2DBC8',
    },
  },
  {
    id: 'bauhaus',
    name: 'Bauhaus Sun & Stone',
    category: 'light',
    description: 'Каменный бетон, геометрия и теплое солнце Баухауса',
    preview: {
      canvas: '#E9E8E4',
      surface: '#F7F6F2',
      accent: '#C49A14',
      text: '#1C1B18',
      border: '#D8D6CE',
    },
  },
  {
    id: 'marigold',
    name: 'Saffron & Raw Silk',
    category: 'light',
    description: 'Пряный шафран, календула и необработанный шелк',
    preview: {
      canvas: '#F0EAE1',
      surface: '#FBF8F3',
      accent: '#D47A1C',
      text: '#241D16',
      border: '#E4DACB',
    },
  },
  {
    id: 'nordic',
    name: 'Nordic Birch & Fjord',
    category: 'light',
    description: 'Светлая северная береза и холодная вода фьорда',
    preview: {
      canvas: '#E6E8E6',
      surface: '#F4F5F4',
      accent: '#4A6B75',
      text: '#1C2120',
      border: '#D6DAD7',
    },
  },
  {
    id: 'oxford',
    name: 'Oxford Ink & Vellum',
    category: 'light',
    description: 'Глубокие чернила Оксфорда и гладкая веленевая бумага',
    preview: {
      canvas: '#EAE8E1',
      surface: '#F7F6F2',
      accent: '#344E68',
      text: '#192028',
      border: '#DAD6CA',
    },
  },
  {
    id: 'fig',
    name: 'Fig & Linen',
    category: 'light',
    description: 'Спелый средиземноморский инжир и небеленое льняное полотно',
    preview: {
      canvas: '#EDEAE6',
      surface: '#F9F7F5',
      accent: '#7A495B',
      text: '#241E20',
      border: '#DFD9D2',
    },
  },
  {
    id: 'alba',
    name: 'Gallery Chalk & Cobalt',
    category: 'light',
    description: 'Галерейный белый мел и акцентный глубокий кобальт',
    preview: {
      canvas: '#EDEDEB',
      surface: '#FBFBF9',
      accent: '#2B54A3',
      text: '#18191C',
      border: '#DEDEDC',
    },
  },
  // 2 Dark Static Themes (Вечерние чернила)
  {
    id: 'espresso',
    name: 'Evening Espresso & Brass',
    category: 'dark',
    description: 'Вечерний эспрессо, матовая латунь и полированный орех',
    preview: {
      canvas: '#1C1917',
      surface: '#25211E',
      accent: '#D4A359',
      text: '#EAE4DC',
      border: '#36302B',
    },
  },
  {
    id: 'cypress',
    name: 'Midnight Cypress & Moss',
    category: 'dark',
    description: 'Полуночный хвойный кипарис, ночной туман и лесной мох',
    preview: {
      canvas: '#181C18',
      surface: '#202620',
      accent: '#84A36E',
      text: '#E4E8E1',
      border: '#2E372E',
    },
  },
  // 3 Animated & Seasonal Themes (Анимированные и сезонные)
  {
    id: 'kyoto-snow',
    name: 'Зимний Киото',
    category: 'light',
    description: 'Морозная сизая палитра, кристаллический снег и иней',
    ambientPreset: 'snow',
    preview: {
      canvas: '#CFDDEB',
      surface: '#EBF3FA',
      accent: '#2B624A',
      text: '#152230',
      border: '#B5C8DB',
    },
  },
  {
    id: 'tsukimi',
    name: 'Осенний Киото',
    category: 'dark',
    description: 'Листопад из листьев момидзи и гинкго, ковер у земли и вихрь от мыши',
    ambientPreset: 'autumn',
    preview: {
      canvas: '#1C1A18',
      surface: '#25221F',
      accent: '#D97746',
      text: '#EBE5DC',
      border: '#38332D',
    },
  },
  {
    id: 'celestial',
    name: 'Celestial Atlas & Astrolabe',
    category: 'dark',
    description: 'Полуночные чернила, латунь и живой интерактивный атлас созвездий',
    ambientPreset: 'celestial',
    preview: {
      canvas: '#111622',
      surface: '#171E2D',
      accent: '#D4A853',
      text: '#E9EEF5',
      border: '#252E42',
    },
  },
  {
    id: 'frost-pine',
    name: 'Хвойная метель',
    category: 'dark',
    description: 'Ночной хвойный бор, изумрудная латунь и стремительная метель',
    ambientPreset: 'blizzard',
    preview: {
      canvas: '#1A2420',
      surface: '#222E2A',
      accent: '#4E9B73',
      text: '#E5EFEA',
      border: '#30403A',
    },
  },
];

export const LIGHT_THEMES = THEMES.filter((t) => t.category === 'light' && !t.ambientPreset);
export const DARK_THEMES = THEMES.filter((t) => t.category === 'dark' && !t.ambientPreset);
export const ANIMATED_THEMES = THEMES.filter((t) => Boolean(t.ambientPreset));

export function getTheme(id: ThemeId): ThemeDefinition {
  return THEMES.find((t) => t.id === id) ?? THEMES[0];
}

