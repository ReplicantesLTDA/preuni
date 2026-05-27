import { create } from 'zustand';
import type { Student } from '@/types/student';

export type SessionStatus = 'loading' | 'authed' | 'anon';

export interface SessionState {
  status: SessionStatus;
  student: Student | null;
  setLoading: () => void;
  setAuthed: (student: Student | null) => void;
  setAnon: () => void;
  patchStudent: (patch: Partial<Student>) => void;
}

export const useSessionStore = create<SessionState>((set) => ({
  status: 'loading',
  student: null,
  setLoading: () => set({ status: 'loading' }),
  setAuthed: (student) => set({ status: 'authed', student }),
  setAnon: () => set({ status: 'anon', student: null }),
  patchStudent: (patch) =>
    set((s) => ({ student: s.student ? { ...s.student, ...patch } : s.student })),
}));
