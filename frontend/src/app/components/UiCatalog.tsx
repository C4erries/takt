import React, { useState } from 'react';
import {
  Sparkles,
  Search,
  Plus,
  Clock,
  Calendar,
} from 'lucide-react';
import {
  Button,
  Input,
  Checkbox,
  Badge,
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from '@/shared/components/ui';

interface SampleTask {
  id: string;
  title: string;
  tag: 'persimmon' | 'denim' | 'straw' | 'matcha';
  tagLabel: string;
  estimate: string;
  completed: boolean;
}

interface UiCatalogProps {
  onOpenModal: () => void;
}

export const UiCatalog: React.FC<UiCatalogProps> = ({ onOpenModal }) => {
  const [tasks, setTasks] = useState<SampleTask[]>([
    {
      id: '1',
      title: 'Сверить архитектурные инварианты Clean Architecture',
      tag: 'denim',
      tagLabel: 'Архитектура',
      estimate: '25 мин',
      completed: false,
    },
    {
      id: '2',
      title: 'Настроить тактильную палитру Washi и шрифты Newsreader',
      tag: 'matcha',
      tagLabel: 'UI-Kit',
      estimate: '15 мин',
      completed: true,
    },
    {
      id: '3',
      title: 'Оформить карточку ежедневного фокуса Focus of the Day',
      tag: 'persimmon',
      tagLabel: '@фокус',
      estimate: '40 мин',
      completed: false,
    },
  ]);

  const [searchQuery, setSearchQuery] = useState('');
  const [btnLoading, setBtnLoading] = useState(false);

  const toggleTask = (id: string) => {
    setTasks((prev) =>
      prev.map((t) => (t.id === id ? { ...t, completed: !t.completed } : t))
    );
  };

  const handleSimulateAction = () => {
    setBtnLoading(true);
    setTimeout(() => setBtnLoading(false), 900);
  };

  return (
    <div className="space-y-8">
      {/* 01: Color Palette */}
      <section className="space-y-3">
        <h2 className="text-xs font-mono uppercase tracking-wider text-muted">
          01 · Семантическая палитра
        </h2>
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-6 gap-2.5 text-xs">
          <div className="p-3 rounded-lg border border-subtle bg-canvas space-y-1">
            <div className="w-4 h-4 rounded-full border border-subtle bg-canvas" />
            <p className="font-medium">Canvas</p>
            <p className="font-mono text-[10px] text-muted">#E8E5DA</p>
          </div>
          <div className="p-3 rounded-lg border border-subtle bg-surface space-y-1">
            <div className="w-4 h-4 rounded-full border border-subtle bg-surface" />
            <p className="font-medium">Washi Surface</p>
            <p className="font-mono text-[10px] text-muted">#F7F5EE</p>
          </div>
          <div className="p-3 rounded-lg border border-subtle bg-accent-surface space-y-1">
            <div className="w-4 h-4 rounded-full bg-accent-surface" />
            <p className="font-medium">Milky Matcha</p>
            <p className="font-mono text-[10px] text-muted">#D6DEC8</p>
          </div>
          <div className="p-3 rounded-lg border border-subtle bg-accent-solid text-surface space-y-1">
            <div className="w-4 h-4 rounded-full bg-surface" />
            <p className="font-medium">Matcha Solid</p>
            <p className="font-mono text-[10px] opacity-80">#566B44</p>
          </div>
          <div className="p-3 rounded-lg border border-subtle bg-tag-persimmon/50 space-y-1">
            <div className="w-4 h-4 rounded-full bg-tag-persimmon" />
            <p className="font-medium">Persimmon</p>
            <p className="font-mono text-[10px] text-muted">#EBD5C3</p>
          </div>
          <div className="p-3 rounded-lg border border-subtle bg-tag-denim/50 space-y-1">
            <div className="w-4 h-4 rounded-full bg-tag-denim" />
            <p className="font-medium">Denim Indigo</p>
            <p className="font-mono text-[10px] text-muted">#CFD8E0</p>
          </div>
        </div>
      </section>

      {/* 02: Buttons */}
      <section className="space-y-3">
        <h2 className="text-xs font-mono uppercase tracking-wider text-muted">
          02 · Кнопки (Button Variants & Sizes)
        </h2>
        <div className="flex flex-wrap items-center gap-3 p-4 rounded-xl border border-subtle bg-canvas/30">
          <Button
            variant="solid"
            onClick={handleSimulateAction}
            isLoading={btnLoading}
            leftIcon={<Plus className="w-4 h-4" />}
          >
            Матча Solid
          </Button>
          <Button variant="outline" leftIcon={<Sparkles className="w-4 h-4" />}>
            1px Outline
          </Button>
          <Button variant="secondary">Secondary</Button>
          <Button variant="ghost">Ghost</Button>
          <Button variant="solid" size="sm">
            Размер SM
          </Button>
          <Button variant="outline" size="lg">
            Размер LG
          </Button>
          <Button variant="solid" disabled>
            Disabled
          </Button>
        </div>
      </section>

      {/* 03 & 04: Inputs and Badges */}
      <section className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="space-y-3">
          <h2 className="text-xs font-mono uppercase tracking-wider text-muted">
            03 · Поля ввода (Input)
          </h2>
          <div className="space-y-3 p-4 rounded-xl border border-subtle bg-canvas/30">
            <Input
              placeholder="Быстрый захват задачи в Inbox (Cmd+K)..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              leftAdornment={<Search className="w-4 h-4" />}
              rightAdornment={
                <kbd className="text-[10px] font-mono px-1.5 py-0.5 rounded border border-subtle bg-surface">
                  ⌘K
                </kbd>
              }
            />
            <Input
              placeholder="Задача с ошибкой валидации"
              defaultValue="Недопустимый формат даты"
              error="Укажите корректный срок выполнения в формате ДД.ММ"
            />
          </div>
        </div>

        <div className="space-y-3">
          <h2 className="text-xs font-mono uppercase tracking-wider text-muted">
            04 · Капсульные теги (Badge Pills)
          </h2>
          <div className="p-4 rounded-xl border border-subtle bg-canvas/30 flex flex-wrap gap-2 items-center min-h-[110px]">
            <Badge variant="persimmon" dot>
              @дом
            </Badge>
            <Badge variant="denim" dot>
              Проект Takt
            </Badge>
            <Badge variant="straw" dot>
              20 мин
            </Badge>
            <Badge variant="matcha" dot>
              Сегодня
            </Badge>
            <Badge variant="subtle">Входящие</Badge>
            <Badge variant="denim" size="sm">
              Компактный
            </Badge>
          </div>
        </div>
      </section>

      {/* 05: Checkboxes & Task Rows */}
      <section className="space-y-3">
        <div className="flex items-center justify-between">
          <h2 className="text-xs font-mono uppercase tracking-wider text-muted">
            05 · Чекбоксы и Строки задач (18px круг + J/K пружинный фокус)
          </h2>
          <span className="text-[11px] font-mono text-muted">cubic-bezier(0.22, 1, 0.36, 1)</span>
        </div>

        <div className="border border-subtle rounded-xl divide-y divide-subtle/70 bg-surface overflow-hidden">
          {tasks.map((task) => (
            <div
              key={task.id}
              className="flex items-center justify-between px-4 py-3 hover:bg-accent-surface/30 transition-all duration-150 ease-spring-jk group cursor-pointer"
              onClick={() => toggleTask(task.id)}
            >
              <div className="flex items-center gap-3">
                <Checkbox
                  checked={task.completed}
                  onChange={() => toggleTask(task.id)}
                  onClick={(e) => e.stopPropagation()}
                />
                <span
                  className={`text-sm transition-all duration-150 ${
                    task.completed ? 'line-through text-muted' : 'text-primary'
                  }`}
                >
                  {task.title}
                </span>
              </div>

              <div className="flex items-center gap-2">
                <Badge variant={task.tag} size="sm">
                  {task.tagLabel}
                </Badge>
                <span className="text-[11px] font-mono text-muted flex items-center gap-1">
                  <Clock className="w-3 h-3" />
                  {task.estimate}
                </span>
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* 06: Cards & Modal Trigger */}
      <section className="space-y-3">
        <h2 className="text-xs font-mono uppercase tracking-wider text-muted">
          06 · Карточки Washi и Модальные окна (Modal)
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Card variant="sheet">
            <CardHeader>
              <CardTitle>План на день</CardTitle>
              <CardDescription>Сфокусируйтесь на 3 главных делах (GTD Rule)</CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-xs text-muted leading-relaxed">
                Карточка на подложке рисовой бумаги Washi со шрифтом заголовка Newsreader и тонкой
                границей 1px.
              </p>
            </CardContent>
            <CardFooter className="justify-between">
              <span className="text-[11px] font-mono text-muted">3 задачи в очереди</span>
              <Button size="sm" variant="outline" onClick={onOpenModal}>
                Открыть детали
              </Button>
            </CardFooter>
          </Card>

          <Card
            variant="subtle"
            className="flex flex-col justify-center items-center text-center p-6 space-y-3"
          >
            <Calendar className="w-8 h-8 text-text-muted stroke-[1.5]" />
            <h3 className="font-serif text-xl text-primary">Диалог быстрого захвата</h3>
            <p className="text-xs text-muted max-w-xs">
              Проверьте модальное окно без тяжелого размытия блюра с мягким затемнением
            </p>
            <Button variant="solid" onClick={onOpenModal}>
              Показать Modal
            </Button>
          </Card>
        </div>
      </section>
    </div>
  );
};
