import { useSessionStore } from '@/stores/sessionStore';

const STUDENT = {
  id: '00000000-0000-4000-a000-000000000011',
  displayName: 'Maria',
  email: 'maria@preuni.com',
  xpTotal: 10,
  streakCount: 2,
  readinessScore: 50,
  onboardingCompleted: true,
};

describe('useSessionStore', () => {
  afterEach(() => useSessionStore.setState({ status: 'loading', student: null }));

  it('starts in the loading state', () => {
    expect(useSessionStore.getState().status).toBe('loading');
  });

  it('setLoading sets status to loading', () => {
    useSessionStore.getState().setAnon();
    useSessionStore.getState().setLoading();
    expect(useSessionStore.getState().status).toBe('loading');
  });

  it('setAuthed sets status and student', () => {
    useSessionStore.getState().setAuthed(STUDENT);
    expect(useSessionStore.getState().status).toBe('authed');
    expect(useSessionStore.getState().student).toEqual(STUDENT);
  });

  it('setAnon clears the student and sets status', () => {
    useSessionStore.getState().setAuthed(STUDENT);
    useSessionStore.getState().setAnon();
    expect(useSessionStore.getState().status).toBe('anon');
    expect(useSessionStore.getState().student).toBeNull();
  });

  it('patchStudent merges into the current student when one exists', () => {
    useSessionStore.getState().setAuthed(STUDENT);
    useSessionStore.getState().patchStudent({ xpTotal: 999 });
    expect(useSessionStore.getState().student?.xpTotal).toBe(999);
    expect(useSessionStore.getState().student?.displayName).toBe('Maria');
  });

  it('patchStudent is a no-op when there is no current student', () => {
    useSessionStore.setState({ status: 'anon', student: null });
    useSessionStore.getState().patchStudent({ xpTotal: 999 });
    expect(useSessionStore.getState().student).toBeNull();
  });
});
