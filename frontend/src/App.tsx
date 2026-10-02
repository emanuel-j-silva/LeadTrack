import React from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Layout } from './components/Layout';
import { SubordinatesPage } from './pages/SubordinatesPage';
import { EvaluatePage } from './pages/EvaluatePage';
import { EvaluationHistoryPage } from './pages/EvaluationHistoryPage';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

export const App: React.FC = () => {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Layout />}>
            <Route index element={<SubordinatesPage />} />
            <Route path="evaluate/:id" element={<EvaluatePage />} />
            <Route path="history/:id" element={<EvaluationHistoryPage />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
};

export default App;
