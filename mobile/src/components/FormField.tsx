import { StyleSheet, Text, TextInput, View, type TextInputProps } from 'react-native';
import { useTheme } from '@/theme';

export interface FormFieldProps extends Omit<TextInputProps, 'onChange'> {
  label: string;
  value: string;
  onChangeText: (v: string) => void;
  error?: string;
  helper?: string;
  testID?: string;
}

export function FormField(props: FormFieldProps) {
  const { label, value, onChangeText, error, helper, testID, ...rest } = props;
  const { color, radius, space, font, size } = useTheme();

  return (
    <View style={{ marginBottom: space[4] }}>
      <Text
        style={{
          color: color.ink[1],
          fontFamily: font.heading,
          fontSize: size.sm,
          marginBottom: space[1],
        }}
      >
        {label}
      </Text>
      <TextInput
        value={value}
        onChangeText={onChangeText}
        placeholderTextColor={color.ink.muted}
        testID={testID}
        accessibilityLabel={label}
        style={[
          styles.input,
          {
            color: color.ink[0],
            borderColor: error ? color.danger : color.ink.muted,
            borderRadius: radius.sm,
            padding: space[3],
            fontFamily: font.body,
            fontSize: size.md,
            backgroundColor: color.paper[0],
          },
        ]}
        {...rest}
      />
      {error ? (
        <Text
          style={{
            color: color.danger,
            fontFamily: font.body,
            fontSize: size.sm,
            marginTop: space[1],
          }}
        >
          {error}
        </Text>
      ) : helper ? (
        <Text
          style={{
            color: color.ink.muted,
            fontFamily: font.body,
            fontSize: size.sm,
            marginTop: space[1],
          }}
        >
          {helper}
        </Text>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  input: { borderWidth: 1 },
});
