import { useRef } from 'react';
import { StyleSheet, Text, TextInput, View } from 'react-native';
import { useTheme } from '@/theme';

export interface OtpInputProps {
  value: string;
  onChange: (v: string) => void;
  length?: number;
  error?: string;
  autoFocus?: boolean;
  testID?: string;
}

export function OtpInput(props: OtpInputProps) {
  const { value, onChange, length = 6, error, autoFocus = false, testID } = props;
  const { color, radius, space, font, size } = useTheme();
  const ref = useRef<TextInput>(null);

  function handleChange(next: string) {
    const digits = next.replace(/\D/g, '').slice(0, length);
    onChange(digits);
  }

  return (
    <View>
      <TextInput
        ref={ref}
        value={value}
        onChangeText={handleChange}
        keyboardType="number-pad"
        autoComplete="one-time-code"
        textContentType="oneTimeCode"
        maxLength={length}
        autoFocus={autoFocus}
        testID={testID}
        accessibilityLabel="Código de verificação"
        style={[
          styles.hidden,
          { borderColor: color.transparent, color: color.transparent },
        ]}
      />
      <View style={styles.row}>
        {Array.from({ length }).map((_, i) => {
          const ch = value[i] ?? '';
          const active = i === value.length;
          return (
            <View
              key={i}
              style={[
                styles.cell,
                {
                  borderColor: error ? color.danger : active ? color.accent : color.ink.muted,
                  borderRadius: radius.sm,
                  padding: space[2],
                  backgroundColor: color.paper[0],
                  minWidth: 36,
                },
              ]}
            >
              <Text
                style={{
                  color: color.ink[0],
                  fontFamily: font.numericBold,
                  fontSize: size.xl,
                  textAlign: 'center',
                }}
              >
                {ch}
              </Text>
            </View>
          );
        })}
      </View>
      {error ? (
        <Text
          style={{
            color: color.danger,
            fontFamily: font.body,
            fontSize: size.sm,
            marginTop: space[2],
            textAlign: 'center',
          }}
        >
          {error}
        </Text>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', justifyContent: 'space-between', gap: 8 },
  cell: { borderWidth: 1, alignItems: 'center', justifyContent: 'center' },
  hidden: { position: 'absolute', width: 1, height: 1, opacity: 0 },
});
