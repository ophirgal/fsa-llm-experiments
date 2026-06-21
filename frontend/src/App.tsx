import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar'
import { AppSidebar } from '@/components/AppSidebar'
import DatasetsPage from '@/pages/DatasetsPage'
import ExperimentsPage from '@/pages/ExperimentsPage'
import NotFoundPage from '@/pages/NotFoundPage'

const queryClient = new QueryClient()

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
    <BrowserRouter>
      <SidebarProvider>
        <AppSidebar />
        <main className="flex-1 overflow-auto p-8">
          <SidebarTrigger className="mb-4" />
          <Routes>
            <Route index element={<Navigate to="/datasets" replace />} />
            <Route path="/datasets" element={<DatasetsPage />} />
            <Route path="/experiments" element={<ExperimentsPage />} />
            <Route path="*" element={<NotFoundPage />} />
          </Routes>
        </main>
      </SidebarProvider>
    </BrowserRouter>
    </QueryClientProvider>
  )
}
