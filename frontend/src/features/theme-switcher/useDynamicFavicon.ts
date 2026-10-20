import { useEffect } from 'react';
import { useThemeStore } from './themeStore';
import { getTheme } from './themes';

function generateFaviconSvg(accent: string, isDark: boolean): string {
  const bg = isDark ? '#0C1117' : '#181C19';
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32" fill="none">
  <rect width="32" height="32" rx="8" fill="${bg}"/>
  <rect x="0.5" y="0.5" width="31" height="31" rx="7.5" stroke="${accent}" stroke-opacity="0.3"/>
  <rect x="7" y="13" width="3.5" height="9" rx="1.75" fill="${accent}" opacity="0.6"/>
  <rect x="14" y="7" width="3.5" height="18" rx="1.75" fill="${accent}"/>
  <rect x="21" y="11" width="3.5" height="13" rx="1.75" fill="${accent}" opacity="0.85"/>
  <circle cx="22.75" cy="6.5" r="1.75" fill="${accent}"/>
</svg>`;
}

export function useDynamicFavicon(): void {
  const themeId = useThemeStore((s) => s.theme);

  useEffect(() => {
    if (typeof document === 'undefined') return;

    const themeDef = getTheme(themeId);
    const accent = themeDef.preview.accent;
    const isDark = themeDef.category === 'dark';
    const svg = generateFaviconSvg(accent, isDark);
    const dataUri = `data:image/svg+xml;utf8,${encodeURIComponent(svg)}`;

    let link = document.querySelector<HTMLLinkElement>("link[rel~='icon']");
    if (!link) {
      link = document.createElement('link');
      link.rel = 'icon';
      document.head.appendChild(link);
    }
    link.type = 'image/svg+xml';
    link.href = dataUri;
  }, [themeId]);
}
