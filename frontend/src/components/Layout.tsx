import React from 'react';
import { Outlet, Link } from 'react-router-dom';
import { UserSwitcher } from './UserSwitcher';

export const Layout: React.FC = () => {
  return (
    <div className="min-h-screen flex flex-col bg-gray-50 text-gray-900">
      <header className="bg-white border-b border-gray-200 sticky top-0 z-10 shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-8">
            <Link to="/" className="flex items-center gap-2">
              <span className="bg-indigo-600 text-white font-bold text-xl px-2.5 py-1 rounded-lg">LT</span>
              <span className="font-bold text-xl text-gray-900 tracking-tight">LeadTrack</span>
            </Link>
            <nav className="hidden md:flex items-center gap-6">
              <Link to="/" className="text-sm font-medium text-gray-600 hover:text-indigo-600 transition-colors">
                Subordinados
              </Link>
            </nav>
          </div>
          <UserSwitcher />
        </div>
      </header>

      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <Outlet />
      </main>

      <footer className="bg-white border-t border-gray-200 py-4 text-center text-xs text-gray-500">
        LeadTrack &copy; {new Date().getFullYear()} - Sistema de Avaliação de Liderados
      </footer>
    </div>
  );
};
