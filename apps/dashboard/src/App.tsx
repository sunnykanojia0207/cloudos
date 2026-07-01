import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ThemeProvider } from '@/components/theme/ThemeProvider';
import { ErrorBoundary } from '@/components/error/ErrorBoundary';
import { AppShell } from '@/components/layout/AppShell';
import { LoadingScreen } from '@/components/error/LoadingScreen';
import { lazy, Suspense } from 'react';

// Lazy-loaded pages for code splitting
const ApplicationsPage = lazy(() => import('@/pages/ApplicationsPage'));
const ApplicationDetailPage = lazy(() => import('@/pages/ApplicationDetailPage'));
const DashboardPage = lazy(() => import('@/pages/DashboardPage'));
const SystemPage = lazy(() => import('@/pages/SystemPage'));
const KernelPage = lazy(() => import('@/pages/KernelPage'));
const CapabilitiesPage = lazy(() => import('@/pages/CapabilitiesPage'));
const CapabilityDetailPage = lazy(() => import('@/pages/CapabilityDetailPage'));
const ProvidersPage = lazy(() => import('@/pages/ProvidersPage'));
const ResourcesPage = lazy(() => import('@/pages/ResourcesPage'));
const ResourceListPage = lazy(() => import('@/pages/ResourceListPage'));
const ResourceDetailPage = lazy(() => import('@/pages/ResourceDetailPage'));
const ControllersPage = lazy(() => import('@/pages/ControllersPage'));
const ControllerDetailPage = lazy(() => import('@/pages/ControllerDetailPage'));
const ProjectsPage = lazy(() => import('@/pages/ProjectsPage'));
const ProjectDetailPage = lazy(() => import('@/pages/ProjectDetailPage'));
const PluginsPage = lazy(() => import('@/pages/PluginsPage'));
const SettingsPage = lazy(() => import('@/pages/SettingsPage'));
const NotFoundPage = lazy(() => import('@/pages/NotFoundPage'));

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 2,
      refetchOnWindowFocus: false,
    },
  },
});

function SuspenseWrapper({ children }: { children: React.ReactNode }) {
  return (
    <Suspense fallback={<LoadingScreen />}>
      {children}
    </Suspense>
  );
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider defaultTheme="dark">
        <ErrorBoundary>
          <BrowserRouter>
            <Routes>
              <Route element={<AppShell />}>
                <Route
                  index
                  element={
                    <SuspenseWrapper>
                      <ApplicationsPage />
                    </SuspenseWrapper>
                  }
                />
                <Route
                  path="applications/:id"
                  element={
                    <SuspenseWrapper>
                      <ApplicationDetailPage />
                    </SuspenseWrapper>
                  }
                />
                <Route
                  path="dashboard"
                  element={
                    <SuspenseWrapper>
                      <DashboardPage />
                    </SuspenseWrapper>
                  }
                />
                <Route
                  path="system"
                  element={
                    <SuspenseWrapper>
                      <SystemPage />
                    </SuspenseWrapper>
                  }
                />
                <Route
                  path="kernel"
                  element={
                    <SuspenseWrapper>
                      <KernelPage />
                    </SuspenseWrapper>
                  }
                />
                <Route
                  path="capabilities"
                  element={
                    <SuspenseWrapper>
                      <CapabilitiesPage />
                    </SuspenseWrapper>
                  }
                />
                <Route
                  path="capabilities/:id"
                  element={
                    <SuspenseWrapper>
                      <CapabilityDetailPage />
                    </SuspenseWrapper>
                  }
                />
                <Route
                  path="providers"
                  element={
                    <SuspenseWrapper>
                      <ProvidersPage />
                    </SuspenseWrapper>
                  }
                />
                <Route
                  path="resources"
                  element={
                    <SuspenseWrapper>
                      <ResourcesPage />
                    </SuspenseWrapper>
                  }
                />
                <Route
                  path="resources/:kind"
                  element={
                    <SuspenseWrapper>
                      <ResourceListPage />
                    </SuspenseWrapper>
                  }
                />
                <Route
                  path="resources/:kind/:id"
                  element={
                    <SuspenseWrapper>
                      <ResourceDetailPage />
                    </SuspenseWrapper>
                  }
                />
                <Route
                  path="controllers"
                  element={
                    <SuspenseWrapper>
                      <ControllersPage />
                    </SuspenseWrapper>
                  }
                />
                <Route
                  path="controllers/:id"
                  element={
                    <SuspenseWrapper>
                      <ControllerDetailPage />
                    </SuspenseWrapper>
                  }
                />
                <Route
                  path="projects"
                  element={
                    <SuspenseWrapper>
                      <ProjectsPage />
                    </SuspenseWrapper>
                  }
                />
                <Route
                  path="projects/:id"
                  element={
                    <SuspenseWrapper>
                      <ProjectDetailPage />
                    </SuspenseWrapper>
                  }
                />
                <Route
                  path="plugins"
                  element={
                    <SuspenseWrapper>
                      <PluginsPage />
                    </SuspenseWrapper>
                  }
                />
                <Route
                  path="settings"
                  element={
                    <SuspenseWrapper>
                      <SettingsPage />
                    </SuspenseWrapper>
                  }
                />
                <Route
                  path="*"
                  element={
                    <SuspenseWrapper>
                      <NotFoundPage />
                    </SuspenseWrapper>
                  }
                />
              </Route>
            </Routes>
          </BrowserRouter>
        </ErrorBoundary>
      </ThemeProvider>
    </QueryClientProvider>
  );
}
