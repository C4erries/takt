# ADR-0006: Система из 12 тем Editorial Tech, переключатель палитр и настраиваемый фон холста

* **Дата**: `2026-10-19`
* **Статус**: `Accepted`
* **Инициатор / Автор**: `Pavel (Founder) & Lead AI Architect`
* **Затрагиваемые подсистемы**: `frontend/src/shared/styles`, `frontend/src/features/theme-switcher`, `frontend/src/entities/task`, `frontend/src/widgets/sidebar`

---

## 🎯 1. Контекст и Проблема (Context)

В рамках Фазы 7 (Спринт 0.7.4) проекту Takt требуется расширение визуального опыта до стандарта MVP 1.0. 
Вместо жестко фиксированной одной темы оформления внедряется богатая палитра из **12 выверенных тем** в эстетике *Editorial Tech / Japanese Stationery / Humanist Minimalism* (10 светлых палитр и 2 глубокие темные темы для вечерней работы).
Одновременно требуется устранить архитектурную инверсию слоев FSD (перенос запросов задач из `features/task-navigation` в `entities/task`) и заменить хардкодные HEX-цвета на семантические токены палитры.

---

## 📐 2. Принятое архитектурное решение (Decision)

### 1. Каталог 12 палитр Editorial Tech
Внедряется 12 тем, переключаемых через CSS-атрибут `[data-theme="<id>"]` на элементе `document.documentElement` (`<html>`):
1. **`kyoto`** (Default): *Kyoto Matcha & Hinoki* (`#E8E5DA` Canvas, `#F7F5EE` Surface, `#566B44` Solid Matcha)
2. **`olive`**: *Olive & Parchment* (`#ECE8DF` Canvas, `#F6F4EE` Surface, `#6E7A4E` Olive Solid)
3. **`claude`**: *Anthropic Pampas & Clay* (`#EFECE6` Canvas, `#FAF8F5` Surface, `#C96442` Terracotta Solid)
4. **`ochre`**: *Ochre & Beeswax* (`#EFECE4` Canvas, `#FAF8F2` Surface, `#B8860B` Mustard Solid)
5. **`bauhaus`**: *Bauhaus Sun & Stone* (`#E9E8E4` Canvas, `#F7F6F2` Surface, `#C49A14` Sun Solid)
6. **`marigold`**: *Saffron & Raw Silk* (`#F0EAE1` Canvas, `#FBF8F3` Surface, `#D47A1C` Amber Solid)
7. **`nordic`**: *Nordic Birch & Fjord* (`#E6E8E6` Canvas, `#F4F5F4` Surface, `#4A6B75` Fjord Solid)
8. **`oxford`**: *Oxford Ink & Vellum* (`#EAE8E1` Canvas, `#F7F6F2` Surface, `#344E68` Ink Solid)
9. **`fig`**: *Fig & Linen* (`#EDEAE6` Canvas, `#F9F7F5` Surface, `#7A495B` Fig Solid)
10. **`alba`**: *Gallery Chalk & Cobalt* (`#EDEDEB` Canvas, `#FBFBF9` Surface, `#2B54A3` Cobalt Solid)
11. **`espresso`**: *Evening Espresso & Brass* (Warm Dark: `#1C1917` Canvas, `#25211E` Surface, `#D4A359` Brass Solid)
12. **`cypress`**: *Midnight Cypress & Moss* (Matcha Dark: `#181C18` Canvas, `#202620` Surface, `#84A36E` Moss Solid)

### 2. Управление состоянием темы и хранение
- Создается фича `frontend/src/features/theme-switcher/`:
  * `themeStore.ts`: Zustand стор активной темы, синхронизированный с `localStorage.getItem('takt_theme')` (fallback: `'kyoto'`).
  * `useTheme()` хук для реактивного доступа и смены темы.
  * `ThemePickerModal.tsx`: модальное окно выбора темы с карточками-сэмплами палитр (кружки canvas/surface/accent, типографика темы, индикатор текущей активной темы).
- Точка входа в UI: пункт «🎨 Оформление и темы» в меню сайдбара.
- Защита от FOUC (Flash of Unstyled Content): предварительное чтение `localStorage` в `index.html` скриптом `<head>`.

### 3. Двухуровневый фон холста
- **Уровень 1**: Бумажная тактильная текстура (SVG Paper Grain noise) поверх `--bg-canvas` с авто-адаптацией прозрачности под темные темы (`espresso`, `cypress`).
- **Уровень 2**: Атмосферный градиентный отлив (*CSS Ambient Glow*) на полотне Канбана: деликатный радиальный градиент с цветом `var(--accent-surface)` с низкой непрозрачностью (~8-12%), придающий глубину бумажному листу.

### 4. Архитектурный рефакторинг FSD
- Перенос базовых запросов и мутаций задач (`useTasksQuery`, `useInboxTasks`, `useCreateTaskMutation`, `useToggleTaskMutation`, `useDeleteTaskMutation`, `TASKS_QUERY_KEY`) из `features/task-navigation` в `entities/task`.
- Устранение горизонтальных перекрестных связей между фичами.
- Замена сырых HEX `#4A6B82` и т.п. на семантические токены палитры.

---

## ⚡ 3. Последствия и Преимущества (Consequences)

1. Продукт получает премиальную кастомизацию интерфейса под настроение и время суток (от теплого утреннего кофе Kyoto/Claude до ночной работы в Espresso/Cypress).
2. Архитектура чиста от циклических и восходящих зависимостей FSD.
3. Полная совместимость со всеми анимациями и кинетикой из ADR-0005.

