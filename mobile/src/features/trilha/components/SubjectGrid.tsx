import { Pressable, StyleSheet, Text, View } from 'react-native';
import { useTheme } from '@/theme';

export interface Subject {
  id: string;
  label: string;
  emoji: string;
  progress?: number; // 0..1
}

export const DEFAULT_SUBJECTS: Subject[] = [
  { id: 'matematica', label: 'Matemática', emoji: '➗' },
  { id: 'portugues', label: 'Português', emoji: '📚' },
  { id: 'redacao', label: 'Redação', emoji: '✍️' },
  { id: 'natureza', label: 'Natureza', emoji: '🌱' },
  { id: 'humanas', label: 'Humanas', emoji: '🏛️' },
  { id: 'linguagens', label: 'Linguagens', emoji: '🗣️' },
];

export interface SubjectGridProps {
  subjects?: Subject[];
  onSelect?: (subject: Subject) => void;
}

export function SubjectGrid({ subjects = DEFAULT_SUBJECTS, onSelect }: SubjectGridProps) {
  const { color, space, font, size, radius } = useTheme();

  return (
    <View style={styles.grid}>
      {subjects.map((s) => (
        <Pressable
          key={s.id}
          accessibilityRole="button"
          accessibilityLabel={s.label}
          onPress={() => onSelect?.(s)}
          style={({ pressed }) => [
            styles.tile,
            {
              backgroundColor: color.paper[1],
              borderRadius: radius.md,
              padding: space[3],
              borderColor: color.ink.muted,
              opacity: pressed ? 0.85 : 1,
            },
          ]}
        >
          <Text style={{ fontSize: size['2xl'] }}>{s.emoji}</Text>
          <Text
            style={{
              color: color.ink[1],
              fontFamily: font.heading,
              fontSize: size.sm,
              marginTop: space[1],
            }}
          >
            {s.label}
          </Text>
          {typeof s.progress === 'number' ? (
            <Text
              style={{
                color: color.ink.muted,
                fontFamily: font.numeric,
                fontSize: size.xs,
                marginTop: space[1],
              }}
            >
              {Math.round(s.progress * 100)}%
            </Text>
          ) : null}
        </Pressable>
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  grid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'space-between',
    gap: 12,
  },
  tile: {
    width: '47%',
    aspectRatio: 1.6,
    borderWidth: 1,
    alignItems: 'flex-start',
    justifyContent: 'space-between',
  },
});
