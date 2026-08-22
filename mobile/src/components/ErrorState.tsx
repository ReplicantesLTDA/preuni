import { EmptyState } from './EmptyState';

export interface ErrorStateProps {
  title?: string;
  body?: string;
  onRetry?: () => void;
}

export function ErrorState({ title = 'Algo deu errado', body, onRetry }: ErrorStateProps) {
  return (
    <EmptyState
      title={title}
      body={body ?? 'Tente novamente em instantes.'}
      mascot="thinking"
      cta={onRetry ? { label: 'Tentar novamente', onPress: onRetry } : undefined}
    />
  );
}
