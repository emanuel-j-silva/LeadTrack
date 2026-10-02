import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { apiGetSubordinates } from '../api/client';

export const SubordinatesPage: React.FC = () => {
  const { data: subordinates, isLoading, error } = useQuery({
    queryKey: ['subordinates'],
    queryFn: apiGetSubordinates,
  });

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="text-gray-500 text-lg">Carregando subordinados...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-6 text-center">
        <h2 className="text-lg font-semibold text-red-800">Erro ao carregar subordinados</h2>
        <p className="text-sm text-red-600 mt-1">{(error as Error).message}</p>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Seus Subordinados</h1>
        <p className="text-sm text-gray-600 mt-1">
          Acompanhe o desempenho e realize as avaliações semanais dos membros da sua hierarquia.
        </p>
      </div>

      {!subordinates || subordinates.length === 0 ? (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-12 text-center">
          <p className="text-gray-500 font-medium">Você não possui subordinados diretos ou indiretos cadastrados.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {subordinates.map((sub) => (
            <div key={sub.id} className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 flex flex-col justify-between">
              <div>
                <div className="flex items-start justify-between">
                  <div>
                    <h3 className="font-semibold text-lg text-gray-900">{sub.name}</h3>
                    <p className="text-xs text-gray-500">{sub.email}</p>
                  </div>
                  <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-indigo-50 text-indigo-700">
                    Nível {sub.depth}
                  </span>
                </div>

                <div className="mt-4 pt-4 border-t border-gray-100">
                  <p className="text-sm font-medium text-gray-700">Cargo: <span className="font-normal text-gray-600">{sub.positionName}</span></p>
                  
                  <div className="mt-2 flex items-center justify-between text-sm">
                    <span className="text-gray-500">Status esta semana:</span>
                    {sub.canEvaluateThisWeek ? (
                      <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-green-50 text-green-700 border border-green-200">
                        Pendente
                      </span>
                    ) : (
                      <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-amber-50 text-amber-700 border border-amber-200">
                        Já avaliado
                      </span>
                    )}
                  </div>

                  {sub.latestEvaluation && (
                    <div className="mt-3 bg-gray-50 p-2.5 rounded-lg border border-gray-100">
                      <div className="flex justify-between text-xs text-gray-500">
                        <span>Última nota:</span>
                        <span className="font-bold text-gray-900">{Number(sub.latestEvaluation.score).toFixed(2)} / 4.00</span>
                      </div>
                      <div className="flex justify-between text-xs text-gray-500 mt-1">
                        <span>Semana:</span>
                        <span>{sub.latestEvaluation.weekStart}</span>
                      </div>
                    </div>
                  )}
                </div>
              </div>

              <div className="mt-6 pt-4 border-t border-gray-100 flex items-center gap-3">
                {sub.canEvaluateThisWeek ? (
                  <Link
                    to={`/evaluate/${sub.id}`}
                    className="flex-1 bg-indigo-600 hover:bg-indigo-700 text-white text-center text-sm font-medium py-2 px-4 rounded-lg transition-colors"
                  >
                    Avaliar
                  </Link>
                ) : (
                  <button
                    disabled
                    className="flex-1 bg-gray-100 text-gray-400 text-center text-sm font-medium py-2 px-4 rounded-lg cursor-not-allowed"
                  >
                    Avaliado
                  </button>
                )}
                <Link
                  to={`/history/${sub.id}`}
                  className="bg-white border border-gray-300 hover:bg-gray-50 text-gray-700 text-sm font-medium py-2 px-3 rounded-lg transition-colors"
                >
                  Histórico
                </Link>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
