/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        canvas: 'var(--bg-canvas)',
        'bg-canvas': 'var(--bg-canvas)',
        'canvas-subtle': 'var(--bg-canvas-subtle)',
        'bg-canvas-subtle': 'var(--bg-canvas-subtle)',
        surface: 'var(--bg-surface)',
        'bg-surface': 'var(--bg-surface)',
        subtle: 'var(--border-subtle)',
        'border-subtle': 'var(--border-subtle)',
        primary: 'var(--text-primary)',
        'text-primary': 'var(--text-primary)',
        muted: 'var(--text-muted)',
        'text-muted': 'var(--text-muted)',
        'accent-surface': 'var(--accent-surface)',
        'accent-solid': 'var(--accent-solid)',
        'accent-solid-hover': 'var(--accent-solid-hover)',
        tag: {
          persimmon: 'var(--tag-persimmon)',
          denim: 'var(--tag-denim)',
          straw: 'var(--tag-straw)',
        },
        'tag-persimmon': 'var(--tag-persimmon)',
        'tag-denim': 'var(--tag-denim)',
        'tag-straw': 'var(--tag-straw)',
        'accent-denim': 'var(--accent-denim)',
        'accent-persimmon': 'var(--accent-persimmon)',
        'accent-straw': 'var(--accent-straw)',
        'accent-bamboo': 'var(--accent-bamboo)',
      },
      fontFamily: {
        serif: ['Newsreader', 'Georgia', 'serif'],
        sans: ['Geist', 'system-ui', 'sans-serif'],
        mono: ['"JetBrains Mono"', 'monospace'],
      },
      transitionTimingFunction: {
        'spring-jk': 'cubic-bezier(0.22, 1, 0.36, 1)',
      },
      transitionDuration: {
        '150': '150ms',
      },
      boxShadow: {
        'washi-sm': '0 1px 2px rgba(33, 36, 30, 0.04)',
        'washi': '0 1px 3px rgba(33, 36, 30, 0.03)',
        'washi-md': '0 4px 14px rgba(33, 36, 30, 0.06)',
        'washi-modal': '0 12px 36px rgba(33, 36, 30, 0.12)',
      },
      borderRadius: {
        'sheet': '16px',
      },
    },
  },
  plugins: [],
};
