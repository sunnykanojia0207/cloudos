'use client';

import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
  type ReactNode,
} from 'react';
import { useNavigate } from 'react-router-dom';
import {
  LayoutDashboard,
  Folder,
  Cpu,
  Server,
  Boxes,
  ShieldCheck,
  Database,
  Activity,
  Puzzle,
  Settings2,
  PlusCircle,
} from 'lucide-react';
import {
  CommandDialog,
  CommandInput,
  CommandList,
  CommandEmpty,
  CommandGroup,
  CommandItem,
} from '@/components/ui/command';
import { useProjects, useControllers } from '@/hooks/useCloudOS';

/* ── Context ──────────────────────────────────────────────── */
interface CommandPaletteContextValue {
  open: boolean;
  setOpen: (open: boolean) => void;
  toggle: () => void;
}

const CommandPaletteContext = createContext<CommandPaletteContextValue | null>(
  null,
);

export function useCommandPalette(): CommandPaletteContextValue {
  const ctx = useContext(CommandPaletteContext);
  if (!ctx) {
    throw new Error(
      'useCommandPalette must be used within a <CommandPaletteProvider />',
    );
  }
  return ctx;
}

/* ── Provider ──────────────────────────────────────────────── */
interface CommandPaletteProviderProps {
  children: ReactNode;
}

export function CommandPaletteProvider({
  children,
}: CommandPaletteProviderProps) {
  const [open, setOpen] = useState(false);

  const toggle = useCallback(() => setOpen((v) => !v), []);

  // Global keyboard shortcut
  useEffect(() => {
    const handler = (e: globalThis.KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        setOpen((v) => !v);
      }
    };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, []);

  return (
    <CommandPaletteContext.Provider value={{ open, setOpen, toggle }}>
      {children}
      <CommandPalette />
    </CommandPaletteContext.Provider>
  );
}

/* ── Page items ───────────────────────────────────────────── */
interface PageItem {
  label: string;
  to: string;
  icon: React.ElementType;
}

const PAGE_ITEMS: PageItem[] = [
  { label: 'Dashboard', to: '/', icon: LayoutDashboard },
  { label: 'Projects', to: '/projects', icon: Folder },
  { label: 'Controllers', to: '/controllers', icon: Activity },
  { label: 'Capabilities', to: '/capabilities', icon: Boxes },
  { label: 'Providers', to: '/providers', icon: ShieldCheck },
  { label: 'Resources', to: '/resources', icon: Database },
  { label: 'System', to: '/system', icon: Server },
  { label: 'Kernel', to: '/kernel', icon: Cpu },
  { label: 'Plugins', to: '/plugins', icon: Puzzle },
  { label: 'Settings', to: '/settings', icon: Settings2 },
];

/* ── Palette component ───────────────────────────────────── */
function CommandPalette() {
  const { open, setOpen } = useCommandPalette();
  const [search, setSearch] = useState('');
  const navigate = useNavigate();

  const { data: projects } = useProjects();
  const { data: controllers } = useControllers();

  const handleSelect = useCallback(
    (path: string) => {
      navigate(path);
      setOpen(false);
    },
    [navigate, setOpen],
  );

  const handleOpenChange = (next: boolean) => {
    setOpen(next);
    if (!next) setSearch('');
  };

  const projectList = projects?.items ?? [];
  const controllerList = controllers?.controllers ?? [];

  // Filter helpers
  const matches = (label: string) =>
    label.toLowerCase().includes(search.toLowerCase());

  const filteredPages = PAGE_ITEMS.filter((p) => matches(p.label));
  const filteredProjects = projectList.filter((p) =>
    matches(p.spec.displayName ?? p.metadata.id),
  );
  const filteredControllers = controllerList.filter((c) =>
    matches(c.name),
  );

  return (
    <CommandDialog open={open} onOpenChange={handleOpenChange}>
      <CommandInput
        placeholder="Search pages, projects, controllers…"
        value={search}
        onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
          setSearch(e.target.value)
        }
      />
      <CommandList>
        <CommandEmpty>No results found.</CommandEmpty>

        {/* Pages */}
        {filteredPages.length > 0 && (
          <CommandGroup heading="Pages">
            {filteredPages.map((page) => (
              <CommandItem
                key={page.to}
                value={page.label}
                onSelect={() => handleSelect(page.to)}
              >
                <page.icon className="mr-2 h-4 w-4" />
                <span>{page.label}</span>
              </CommandItem>
            ))}
          </CommandGroup>
        )}

        {/* Projects */}
        {filteredProjects.length > 0 && (
          <CommandGroup heading="Projects">
            {filteredProjects.map((project) => (
              <CommandItem
                key={project.metadata.id}
                value={project.spec.displayName ?? project.metadata.id}
                onSelect={() =>
                  handleSelect(`/projects/${project.metadata.id}`)
                }
              >
                <Folder className="mr-2 h-4 w-4 text-muted-foreground" />
                <span>{project.spec.displayName ?? project.metadata.id}</span>
              </CommandItem>
            ))}
          </CommandGroup>
        )}

        {/* Controllers */}
        {filteredControllers.length > 0 && (
          <CommandGroup heading="Controllers">
            {filteredControllers.map((controller) => (
              <CommandItem
                key={controller.name}
                value={controller.name}
                onSelect={() =>
                  handleSelect(`/controllers/${controller.name}`)
                }
              >
                <Cpu className="mr-2 h-4 w-4 text-muted-foreground" />
                <span>{controller.name}</span>
              </CommandItem>
            ))}
          </CommandGroup>
        )}

        {/* Quick actions */}
        {(!search ||
          'create project'.includes(search.toLowerCase())) && (
          <CommandGroup heading="Quick Actions">
            <CommandItem
              value="Create Project"
              onSelect={() => handleSelect('/projects')}
            >
              <PlusCircle className="mr-2 h-4 w-4" />
              <span>Create Project</span>
            </CommandItem>
          </CommandGroup>
        )}
      </CommandList>
    </CommandDialog>
  );
}
