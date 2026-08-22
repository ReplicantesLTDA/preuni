import { tokenStore } from './tokenStore';
import { useSessionStore } from '@/stores/sessionStore';
import type { ApiClient } from '@/lib/api/client';
import { StudentSchema, type Student } from '@/types/student';
import { isAppError } from '@/lib/api/errors';

export async function bootstrapSession(api: ApiClient): Promise<void> {
  const store = await tokenStore.load();
  const { setAuthed, setAnon } = useSessionStore.getState();
  if (!store) {
    setAnon();
    return;
  }

  try {
    const me = await api.request<Student>({
      method: 'GET',
      path: '/v1/students/me',
      schema: StudentSchema,
    });
    setAuthed(me);
  } catch (err) {
    if (isAppError(err) && err.kind === 'Unauthorized') {
      await tokenStore.clear();
      setAnon();
      return;
    }
    setAnon();
  }
}
