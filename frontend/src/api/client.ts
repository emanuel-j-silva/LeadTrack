import { EvaluationInput, Employee, Question, Subordinate, Evaluation, ApiErrorResponse } from '../types/api';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export function getCurrentEmployeeId(): number {
  const stored = localStorage.getItem('currentEmployeeId');
  if (!stored) return 1; // Default to Alice (ID 1)
  const num = parseInt(stored, 10);
  return isNaN(num) ? 1 : num;
}

export function setCurrentEmployeeId(id: number) {
  localStorage.setItem('currentEmployeeId', id.toString());
}

async function fetchWithAuth(endpoint: string, options: RequestInit = {}): Promise<Response> {
  const empId = getCurrentEmployeeId();
  const headers = new Headers(options.headers || {});
  headers.set('X-Employee-Id', empId.toString());
  if (options.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  const res = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
  });

  return res;
}

export async function apiGetEmployees(): Promise<Employee[]> {
  const res = await fetch(`${API_BASE_URL}/api/employees`);
  if (!res.ok) {
    throw new Error('Failed to fetch employees');
  }
  return res.json();
}

export async function apiGetMe(): Promise<Employee> {
  const res = await fetchWithAuth('/api/me');
  if (!res.ok) {
    const errData: ApiErrorResponse = await res.json().catch(() => ({ error: { code: 'UNKNOWN', message: 'Failed to fetch current user' } }));
    throw new Error(errData.error?.message || 'Unauthorized');
  }
  return res.json();
}

export async function apiGetQuestions(): Promise<Question[]> {
  const res = await fetchWithAuth('/api/questions');
  if (!res.ok) {
    throw new Error('Failed to fetch questions');
  }
  return res.json();
}

export async function apiGetSubordinates(): Promise<Subordinate[]> {
  const res = await fetchWithAuth('/api/subordinates');
  if (!res.ok) {
    const errData: ApiErrorResponse = await res.json().catch(() => ({ error: { code: 'UNKNOWN', message: 'Failed to fetch subordinates' } }));
    throw new Error(errData.error?.message || 'Failed to fetch subordinates');
  }
  return res.json();
}

export async function apiCreateEvaluation(evaluatedId: number, input: EvaluationInput): Promise<{ id: number }> {
  const res = await fetchWithAuth(`/api/employees/${evaluatedId}/evaluations`, {
    method: 'POST',
    body: JSON.stringify(input),
  });
  if (!res.ok) {
    const errData: ApiErrorResponse = await res.json().catch(() => ({ error: { code: 'UNKNOWN', message: 'Failed to create evaluation' } }));
    const err = new Error(errData.error?.message || 'Failed to create evaluation');
    (err as any).code = errData.error?.code;
    (err as any).status = res.status;
    throw err;
  }
  return res.json();
}

export async function apiGetEvaluations(evaluatedId: number): Promise<Evaluation[]> {
  const res = await fetchWithAuth(`/api/employees/${evaluatedId}/evaluations`);
  if (!res.ok) {
    const errData: ApiErrorResponse = await res.json().catch(() => ({ error: { code: 'UNKNOWN', message: 'Failed to fetch evaluations' } }));
    throw new Error(errData.error?.message || 'Failed to fetch evaluations');
  }
  return res.json();
}

export async function apiGetLatestEvaluation(evaluatedId: number): Promise<Evaluation> {
  const res = await fetchWithAuth(`/api/employees/${evaluatedId}/evaluations/latest`);
  if (!res.ok) {
    const errData: ApiErrorResponse = await res.json().catch(() => ({ error: { code: 'UNKNOWN', message: 'Failed to fetch latest evaluation' } }));
    throw new Error(errData.error?.message || 'Failed to fetch latest evaluation');
  }
  return res.json();
}
