import React from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { apiGetEmployees, getCurrentEmployeeId, setCurrentEmployeeId } from '../api/client';

export const UserSwitcher: React.FC = () => {
  const queryClient = useQueryClient();
  const [currentId, setCurrentId] = React.useState<number>(() => getCurrentEmployeeId());

  const { data: employees, isLoading } = useQuery({
    queryKey: ['employees'],
    queryFn: apiGetEmployees,
  });

  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const newId = parseInt(e.target.value, 10);
    setCurrentEmployeeId(newId);
    setCurrentId(newId);
    // Invalidate queries so everything refetches with new header
    queryClient.invalidateQueries();
  };

  if (isLoading || !employees) {
    return <div className="text-sm text-gray-500">Carregando usuário...</div>;
  }

  const currentEmp = employees.find((emp) => emp.id === currentId);

  return (
    <div className="flex items-center gap-3 bg-white px-3 py-1.5 rounded-lg shadow-sm border border-gray-200">
      <div className="text-right">
        <span className="block text-xs text-gray-500 font-medium">Líder atual (Simulação)</span>
        <span className="block text-sm font-semibold text-indigo-600">
          {currentEmp ? `${currentEmp.name} (${currentEmp.positionName})` : `ID: ${currentId}`}
        </span>
      </div>
      <select
        value={currentId}
        onChange={handleChange}
        className="text-sm bg-gray-50 border border-gray-300 rounded px-2 py-1 focus:outline-none focus:ring-2 focus:ring-indigo-500"
      >
        {employees.map((emp) => (
          <option key={emp.id} value={emp.id}>
            {emp.id} - {emp.name} ({emp.positionName})
          </option>
        ))}
      </select>
    </div>
  );
};
