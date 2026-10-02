import React, { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { apiGetQuestions, apiCreateEvaluation, apiGetSubordinates } from '../api/client';

export const EvaluatePage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const evaluatedId = parseInt(id || '0', 10);
  const navigate = useNavigate();

  const [scores, setScores] = useState<Record<number, number>>({});
  const [submitting, setSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const { data: questions, isLoading: loadingQuestions } = useQuery({
    queryKey: ['questions'],
    queryFn: apiGetQuestions,
  });

  const { data: subordinates } = useQuery({
    queryKey: ['subordinates'],
    queryFn: apiGetSubordinates,
  });

  const subordinate = subordinates?.find((s) => s.id === evaluatedId);

  const handleScoreChange = (questionId: number, value: number) => {
    setScores((prev) => ({ ...prev, [questionId]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!questions) return;

    if (questions.some((q) => !scores[q.id])) {
      setErrorMessage('Por favor, responda todas as 6 questões antes de enviar.');
      return;
    }

    setSubmitting(true);
    setErrorMessage(null);

    try {
      const answers = questions.map((q) => ({
        questionId: q.id,
        value: scores[q.id],
      }));

      await apiCreateEvaluation(evaluatedId, { answers });
      navigate('/');
    } catch (err: any) {
      const code = err.code;
      if (code === 'ALREADY_EVALUATED_THIS_WEEK') {
        setErrorMessage('Você já avaliou este funcionário nesta semana.');
      } else {
        setErrorMessage(err.message || 'Erro ao enviar avaliação.');
      }
    } finally {
      setSubmitting(false);
    }
  };

  if (loadingQuestions) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="text-gray-500 text-lg">Carregando formulário...</div>
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Avaliação Semanal</h1>
          {subordinate && (
            <p className="text-sm text-gray-600 mt-1">
              Avaliando <span className="font-semibold text-gray-900">{subordinate.name}</span> ({subordinate.positionName})
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

      {errorMessage && (
        <div className="mb-6 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
          {errorMessage}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        {questions?.map((q, idx) => (
          <div key={q.id} className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <div className="flex items-start justify-between gap-4">
              <div>
                <span className="text-xs font-semibold text-indigo-600 uppercase tracking-wider">Questão {idx + 1}</span>
                <h3 className="text-base font-semibold text-gray-900 mt-1">{q.label}</h3>
              </div>
              <span className="shrink-0 inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                Peso: {q.weight}
              </span>
            </div>

            <div className="mt-4 grid grid-cols-2 sm:grid-cols-4 gap-3">
              {[
                { val: 1, label: '1 - Insatisfatório' },
                { val: 2, label: '2 - Regular' },
                { val: 3, label: '3 - Bom' },
                { val: 4, label: '4 - Excelente' },
              ].map((opt) => {
                const selected = scores[q.id] === opt.val;
                return (
                  <button
                    key={opt.val}
                    type="button"
                    onClick={() => handleScoreChange(q.id, opt.val)}
                    className={`py-3 px-4 rounded-lg text-sm font-medium border text-center transition-all ${
                      selected
                        ? 'bg-indigo-600 text-white border-indigo-600 shadow-sm'
                        : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'
                    }`}
                  >
                    {opt.label}
                  </button>
                );
              })}
            </div>
          </div>
        ))}

        <div className="flex items-center justify-end gap-4 pt-4">
          <button
            type="button"
            onClick={() => navigate('/')}
            className="px-5 py-2.5 rounded-lg border border-gray-300 text-gray-700 font-medium text-sm hover:bg-gray-50 transition-colors"
          >
            Cancelar
          </button>
          <button
            type="submit"
            disabled={submitting}
            className="px-6 py-2.5 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white font-medium text-sm shadow-sm transition-colors disabled:opacity-50"
          >
            {submitting ? 'Enviando...' : 'Enviar Avaliação'}
          </button>
        </div>
      </form>
    </div>
  );
};
