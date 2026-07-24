export type SubjectResult = {
  name: string;
  score: number;
  answers: string;
};

export type ExamResult = {
  finCode: string;
  firstName: string;
  lastName: string;
  fatherName: string;
  examDate: string;
  subjects: SubjectResult[];
  totalScore: number;
  passed: boolean;
};

export type LookupResponse = {
  result: ExamResult;
};

export type ApiError = {
  error?: {
    code?: string;
    message?: string;
  };
};
