import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { apiGetEvaluations, apiGetQuestions, apiGetSubordinates } from '../api/client';

export const EvaluationHistoryPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const evaluatedId = parseInt(id || '0', 10);
  const navigate = useNavigate();

  const { data: evaluations, isLoading: loadingEvals, error } = useQuery({
    queryKey: ['evaluations', evaluatedId],
    queryFn: () => apiGetEvaluations(evaluatedId),
  });

  const { data: questions } = useQuery({
    queryKey: ['questions'],
    queryFn: apiGetQuestions,
  });

  const { data: subordinates } = useQuery({
    queryKey: ['subordinates'],
    queryFn: apiGetSubordinates,
  });

  const subordinate = subordinates?.find((s) => s.id === evaluatedId);
  const qMap = new Map(questions?.map((q) => [q.id, q]) || []);

  if (loadingEvals) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="text-gray-500 text-lg">Carregando histórico...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-6 text-center">
        <h2 className="text-lg font-semibold text-red-800">Acesso negado ou erro</h2>
        <p className="text-sm text-red-600 mt-1">{(error as Error).message}</p>
        <button
          onClick={() => navigate('/')}
          className="mt-4 px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm font-medium"
        >
          Voltar para Início
        </button>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Histórico de Avaliações</h1>
          {subordinate && (
            <p className="text-sm text-gray-600 mt-1">
              Colaborador: <span className="font-semibold text-gray-900">{subordinate.name}</span> ({subordinate.positionName})
            </p>
          )}
        </div>
        <button
          onClick={() => navigate('/')}
          className="text-sm text-indigo-600 hover:text-indigo-800 font-medium"
        >
          &larr; Voltar
        </button>
      </div>

      {!evaluations || evaluations.length === 0 ? (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-12 text-center">
          <p className="text-gray-500 font-medium">Nenhuma avaliação encontrada para este subordinado.</p>
        </div>
      ) : (
        <div className="space-y-6">
          {evaluations.map((evalItem) => (
            <div key={evalItem.id} className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
              <div className="flex items-center justify-between pb-4 border-b border-gray-100">
                <div>
                  <span className="text-xs font-semibold text-indigo-600 uppercase tracking-wider">Semana de {evalItem.weekStart}</span>
                  <p className="text-xs text-gray-500 mt-0.5">Enviado em: {new Date(evalItem.createdAt).toLocaleString('pt-BR')}</p>
                </div>
                <div className="text-right">
                  <span className="text-xs text-gray-500 block">Nota Ponderada</span>
                  <span className="text-xl font-bold text-gray-900">{Number(evalItem.score).toFixed(2)} <span className="text-xs font-normal text-gray-500">/ 4.00</span></span>
                </div>
              </div>

              {evalItem.answers && evalItem.answers.length > 0 && (
                <div className="mt-4 space-y-3">
                  <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider">Respostas detalhadas</h4>
                  <div className="grid grid-cols-1 gap-2">
                    {evalItem.answers.map((ans) => {
                      const q = qMap.get(ans.questionId);
                      return (
                        <div key={ans.questionId} className="bg-gray-50 p-3 rounded-lg flex items-center justify-between text-sm">
                          <span className="text-gray-700 font-medium">{q ? q.label : `Questão #${ans.questionId}`}</span>
                          <span className="shrink-0 ml-4 px-2.5 py-1 rounded bg-white border border-gray-200 font-bold text-indigo-600">
                            Nota: {ans.value}
                          </span>
                        </div>
                      );
                    })}
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
