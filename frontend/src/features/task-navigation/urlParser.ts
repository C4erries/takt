export interface ParsedRoute {
  activeNav: string;
  selectedProjectId: string | null;
  activeBoardId: string | null;
  selectedTaskId: string | null;
}

export const LAST_ROUTE_KEY = 'takt_last_route';

export function parseUrl(pathname: string, search: string): ParsedRoute {
  const searchParams = new URLSearchParams(search);
  const selectedTaskId = searchParams.get('task') || null;

  // 1. Match /projects/:projectId/boards/:boardId
  const projectBoardMatch = pathname.match(/^\/projects\/([^/]+)\/boards\/([^/]+)\/?$/);
  if (projectBoardMatch) {
    const projectId = decodeURIComponent(projectBoardMatch[1]);
    const boardId = decodeURIComponent(projectBoardMatch[2]);
    return {
      activeNav: `project:${projectId}`,
      selectedProjectId: projectId,
      activeBoardId: boardId,
      selectedTaskId,
    };
  }

  // 2. Match /projects/:projectId
  const projectMatch = pathname.match(/^\/projects\/([^/]+)\/?$/);
  if (projectMatch) {
    const projectId = decodeURIComponent(projectMatch[1]);
    return {
      activeNav: `project:${projectId}`,
      selectedProjectId: projectId,
      activeBoardId: null,
      selectedTaskId,
    };
  }

  // 3. Match /today
  if (pathname === '/today' || pathname.startsWith('/today/')) {
    return {
      activeNav: 'today',
      selectedProjectId: null,
      activeBoardId: null,
      selectedTaskId,
    };
  }

  // 4. Match /automations
  if (pathname === '/automations' || pathname.startsWith('/automations/')) {
    return {
      activeNav: 'automations',
      selectedProjectId: null,
      activeBoardId: null,
      selectedTaskId,
    };
  }

  // 5. Match /archive
  if (pathname === '/archive' || pathname.startsWith('/archive/')) {
    return {
      activeNav: 'archive',
      selectedProjectId: null,
      activeBoardId: null,
      selectedTaskId,
    };
  }

  // 4. Default: /inbox or /
  return {
    activeNav: 'inbox',
    selectedProjectId: null,
    activeBoardId: null,
    selectedTaskId,
  };
}

export function buildUrl({
  activeNav,
  selectedProjectId,
  activeBoardId,
  selectedTaskId,
}: ParsedRoute): string {
  let pathname = '/inbox';
  const isProject = activeNav.startsWith('project:') || Boolean(selectedProjectId);
  const projectId =
    selectedProjectId ||
    (activeNav.startsWith('project:') ? activeNav.replace('project:', '') : null);

  if (isProject && projectId) {
    if (activeBoardId) {
      pathname = `/projects/${encodeURIComponent(projectId)}/boards/${encodeURIComponent(activeBoardId)}`;
    } else {
      pathname = `/projects/${encodeURIComponent(projectId)}`;
    }
  } else if (activeNav === 'today') {
    pathname = '/today';
  } else if (activeNav === 'automations') {
    pathname = '/automations';
  } else if (activeNav === 'archive') {
    pathname = '/archive';
  } else {
    pathname = '/inbox';
  }

  const currentSearch = typeof window !== 'undefined' ? window.location.search : '';
  const searchParams = new URLSearchParams(currentSearch);
  if (selectedTaskId) {
    searchParams.set('task', selectedTaskId);
  } else {
    searchParams.delete('task');
  }

  const search = searchParams.toString();
  return search ? `${pathname}?${search}` : pathname;
}

