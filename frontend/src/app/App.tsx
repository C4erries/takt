import React, { useState } from 'react';
import {
  Button,
  Input,
  Badge,
  Modal,
} from '@/shared/components/ui';
import { UiCatalog } from './components/UiCatalog';
import { StatesShowcase } from './components/StatesShowcase';

export const App: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'ui-kit' | 'states'>('ui-kit');
  const [isModalOpen, setIsModalOpen] = useState(false);

  return (
    <div className="min-h-screen bg-canvas text-primary p-4 md:p-8 flex flex-col items-center">
      {/* Container simulating Surface-on-Canvas */}
      <main className="w-full max-w-4xl bg-surface border border-subtle rounded-sheet shadow-washi p-6 md:p-10 space-y-8">
        {/* Header */}
        <header className="border-b border-subtle/60 pb-6 flex flex-col md:flex-row md:items-end justify-between gap-4">
          <div className="space-y-1.5">
            <div className="flex items-center gap-2">
              <span className="w-2.5 h-2.5 rounded-full bg-accent-solid inline-block" />
              <span className="text-[11px] font-mono tracking-widest uppercase text-muted">
                Takt System · Sprint 0.1.1
              </span>
            </div>
            <h1 className="font-serif text-3xl md:text-4xl font-normal tracking-tight text-primary">
              Soft Editorial GTD
            </h1>
            <p className="text-sm text-muted">
              Витрина атомарного UI-Kit в эстетике японской канцелярии Midori MD и Hobonichi
            </p>
          </div>

          <div className="flex items-center gap-2 bg-canvas/70 p-1 rounded-xl border border-subtle/80">
            <Button
              size="sm"
              variant={activeTab === 'ui-kit' ? 'solid' : 'ghost'}
              onClick={() => setActiveTab('ui-kit')}
            >
              Компоненты UI
            </Button>
            <Button
              size="sm"
              variant={activeTab === 'states' ? 'solid' : 'ghost'}
              onClick={() => setActiveTab('states')}
            >
              4 Состояния экрана
            </Button>
          </div>
        </header>

        {/* Tab Content */}
        {activeTab === 'ui-kit' ? (
          <UiCatalog onOpenModal={() => setIsModalOpen(true)} />
        ) : (
          <StatesShowcase />
        )}

        {/* Modal Instance */}
        <Modal
          isOpen={isModalOpen}
          onClose={() => setIsModalOpen(false)}
          title="Быстрый захват задачи"
          description="Добавьте задачу в Inbox без отвлечения от текущего контекста"
          footer={
            <>
              <Button variant="ghost" onClick={() => setIsModalOpen(false)}>
                Отмена
              </Button>
              <Button variant="solid" onClick={() => setIsModalOpen(false)}>
                Сохранить в Inbox
              </Button>
            </>
          }
        >
          <div className="space-y-3">
            <Input placeholder="Что нужно сделать?" autoFocus />
            <div className="flex items-center gap-2">
              <Badge variant="straw">20 мин</Badge>
              <Badge variant="persimmon">@фокус</Badge>
            </div>
          </div>
        </Modal>
      </main>
    </div>
  );
};

export default App;
