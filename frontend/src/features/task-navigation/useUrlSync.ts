import { useEffect, useRef } from 'react';
import { useTaskStore } from './taskStore';
import { useProjectsQuery, useProjectBoardsQuery } from '@/entities/project';
import { parseUrl, buildUrl, LAST_ROUTE_KEY } from './urlParser';

export * from './urlParser';

export interface UseUrlSyncOptions {
  workspaceId?: string;
}

export const useUrlSync = ({ workspaceId }: UseUrlSyncOptions = {}) => {
  const {
    activeNav,
    setActiveNav,
    selectedProjectId,
    setSelectedProjectId,
    activeBoardId,
    setActiveBoardId,
    selectedTaskId,
  } = useTaskStore();

  const isPopStateRef = useRef(false);
  const isFirstMountRef = useRef(true);
  const lastPathnameRef = useRef<string>(
    typeof window !== 'undefined' ? window.location.pathname : ''
  );

  const projectsQuery = useProjectsQuery(workspaceId);
  const boardsQuery = useProjectBoardsQuery(workspaceId, selectedProjectId ?? undefined);

  // 1. Initial mount: record initial route in localStorage
  useEffect(() => {
    if (typeof window === 'undefined') return;

    let pathname = window.location.pathname;
    let search = window.location.search;

    if (pathname === '/' || pathname === '') {
      const savedRoute = localStorage.getItem(LAST_ROUTE_KEY);
      if (savedRoute && savedRoute !== '/' && savedRoute !== '') {
        window.history.replaceState(null, '', savedRoute);
        pathname = window.location.pathname;
        search = window.location.search;
        const route = parseUrl(pathname, search);
        useTaskStore.setState({
          activeNav: route.activeNav,
          selectedProjectId: route.selectedProjectId,
          activeBoardId: route.activeBoardId,
          selectedTaskId: route.selectedTaskId,
        });
      }
    }

    lastPathnameRef.current = pathname;
    const fullRoute = pathname + search;
    localStorage.setItem(LAST_ROUTE_KEY, fullRoute);
  }, []);

  // 2. Popstate listener for browser Back / Forward buttons
  useEffect(() => {
    if (typeof window === 'undefined') return;

    const handlePopState = () => {
      isPopStateRef.current = true;
      const parsed = parseUrl(window.location.pathname, window.location.search);
      lastPathnameRef.current = window.location.pathname;

      useTaskStore.setState({
        activeNav: parsed.activeNav,
        selectedProjectId: parsed.selectedProjectId,
        activeBoardId: parsed.activeBoardId,
        selectedTaskId: parsed.selectedTaskId,
      });

      const fullRoute = window.location.pathname + window.location.search;
      localStorage.setItem(LAST_ROUTE_KEY, fullRoute);
    };

    window.addEventListener('popstate', handlePopState);
    return () => window.removeEventListener('popstate', handlePopState);
  }, []);

  // 3a. Fallback if initial deep-linked project does not exist
  const initialProjectCheckedRef = useRef(false);

  useEffect(() => {
    if (initialProjectCheckedRef.current) return;
    if (!selectedProjectId || !projectsQuery.isSuccess || projectsQuery.isFetching || !projectsQuery.data) return;

    const projects = projectsQuery.data;
    if (projects.length === 0) return;

    initialProjectCheckedRef.current = true;
    const exists = projects.some((p) => p.id === selectedProjectId);
    if (!exists) {
      setSelectedProjectId(null);
      setActiveBoardId(null);
      setActiveNav('inbox');
      window.history.replaceState(null, '', '/inbox');
      localStorage.setItem(LAST_ROUTE_KEY, '/inbox');
      lastPathnameRef.current = '/inbox';
    }
  }, [
    selectedProjectId,
    projectsQuery.isSuccess,
    projectsQuery.isFetching,
    projectsQuery.data,
    setSelectedProjectId,
    setActiveBoardId,
    setActiveNav,
  ]);

  // 3b. Auto-resolve first board if activeBoardId is null or invalid
  useEffect(() => {
    if (!selectedProjectId || !boardsQuery.isSuccess || !boardsQuery.data) return;
    const boards = boardsQuery.data;
    if (boards.length === 0) return;

    const boardExists = activeBoardId ? boards.some((b) => b.id === activeBoardId) : false;

    if (!activeBoardId || !boardExists) {
      const firstBoardId = boards[0].id;
      setActiveBoardId(firstBoardId);
    }
  }, [
    selectedProjectId,
    boardsQuery.isSuccess,
    boardsQuery.data,
    activeBoardId,
    setActiveBoardId,
  ]);

  // 4. Sync Store state -> Browser URL (pushState vs replaceState)
  useEffect(() => {
    if (typeof window === 'undefined') return;

    if (isFirstMountRef.current) {
      isFirstMountRef.current = false;
      return;
    }

    if (isPopStateRef.current) {
      isPopStateRef.current = false;
      return;
    }

    const targetUrl = buildUrl({
      activeNav,
      selectedProjectId,
      activeBoardId,
      selectedTaskId,
    });

    const [targetPath, targetSearch] = targetUrl.split('?');
    const currentPath = window.location.pathname;
    const currentSearch = window.location.search ? window.location.search.slice(1) : '';
    const currentFull = window.location.pathname + window.location.search;

    if (targetUrl === currentFull) {
      return;
    }

    const isAutoBoardResolution =
      Boolean(selectedProjectId) &&
      lastPathnameRef.current === `/projects/${encodeURIComponent(selectedProjectId!)}` &&
      targetPath.startsWith(`/projects/${encodeURIComponent(selectedProjectId!)}/boards/`);

    const isSearchOnlyChange = targetPath === currentPath && (targetSearch || '') !== currentSearch;

    const isReplace = isSearchOnlyChange || isAutoBoardResolution;

    if (isReplace) {
      window.history.replaceState(null, '', targetUrl);
    } else {
      window.history.pushState(null, '', targetUrl);
    }

    lastPathnameRef.current = targetPath;
    localStorage.setItem(LAST_ROUTE_KEY, targetUrl);
  }, [activeNav, selectedProjectId, activeBoardId, selectedTaskId]);
};
