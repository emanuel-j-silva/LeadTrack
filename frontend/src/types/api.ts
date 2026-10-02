export interface Employee {
  id: number;
  name: string;
  email: string;
  positionName: string;
}

export interface Question {
  id: number;
  label: string;
  weight: number;
}

export interface AnswerInput {
  questionId: number;
  value: number;
}

export interface EvaluationInput {
  answers: AnswerInput[];
}

export interface Answer {
  questionId: number;
  value: number;
}

export interface Evaluation {
  id: number;
  evaluatorId: number;
  evaluatedId: number;
  weekStart: string;
  score: number;
  createdAt: string;
  answers?: Answer[];
}

export interface Subordinate {
  id: number;
  name: string;
  email: string;
  positionName: string;
  depth: number;
  canEvaluateThisWeek: boolean;
  latestEvaluation?: Evaluation;
}

export interface ApiErrorResponse {
  error: {
    code: string;
    message: string;
  };
}
