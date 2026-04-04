# UI Contract: PreuniTextField

**Component**: `PreuniTextField`
**Package**: `com.preuni.shared.ui.components`
**File**: `mobile/shared/src/commonMain/kotlin/com/preuni/shared/ui/components/PreuniTextField.kt`

---

## Purpose

Rounded outlined text field. Wraps Material 3 `OutlinedTextField` overriding the default `ExtraSmall` (4dp) corner shape with `shapes.small` (16dp). Enforces consistent error display pattern and `fillMaxWidth`.

## Signature

```kotlin
@Composable
fun PreuniTextField(
    value: String,
    onValueChange: (String) -> Unit,
    label: String,
    modifier: Modifier = Modifier,
    placeholder: String? = null,
    isError: Boolean = false,
    errorMessage: String? = null,
    singleLine: Boolean = true,
    keyboardOptions: KeyboardOptions = KeyboardOptions.Default,
    keyboardActions: KeyboardActions = KeyboardActions.Default,
    visualTransformation: VisualTransformation = VisualTransformation.None,
    trailingIcon: @Composable (() -> Unit)? = null,
    enabled: Boolean = true,
)
```

## Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `value` | `String` | Yes | — | Current field value |
| `onValueChange` | `(String) -> Unit` | Yes | — | Value change callback |
| `label` | `String` | Yes | — | Floating label text |
| `placeholder` | `String?` | No | `null` | Placeholder shown when empty and unfocused |
| `isError` | `Boolean` | No | `false` | Activates error outline and error text color |
| `errorMessage` | `String?` | No | `null` | Inline error message shown below field when `isError = true` |
| `singleLine` | `Boolean` | No | `true` | Restricts input to one line |
| `keyboardOptions` | `KeyboardOptions` | No | Default | IME type, action key |
| `keyboardActions` | `KeyboardActions` | No | Default | IME action callbacks |
| `visualTransformation` | `VisualTransformation` | No | None | Use `PasswordVisualTransformation()` for password fields |
| `trailingIcon` | `@Composable (() -> Unit)?` | No | `null` | Optional icon/button on the right |
| `enabled` | `Boolean` | No | `true` | Disabled state |

## Visual States

| State | Appearance |
|-------|-----------|
| Default | Outlined border `colorScheme.outline`, rounded 16dp corners |
| Focused | Border changes to `colorScheme.primary`, label floats above |
| Error | Border + label in `colorScheme.error`; `errorMessage` shown in error color below field |
| Disabled | Border and text at 38% opacity; no interaction |

## Constraints

- Always `fillMaxWidth` unless caller overrides.
- Shape: `MaterialTheme.shapes.small` (16dp) — inherited from theme, no hardcoded value in component.
- Error message MUST display inline without causing layout shift to surrounding elements (M3 `supportingText` slot handles this — always reserves space).
- No minimum or maximum character enforcement — validation is the caller's responsibility.

## Usage Example

```kotlin
PreuniTextField(
    value = state.email,
    onValueChange = { store.accept(RegisterStore.Intent.UpdateEmail(it)) },
    label = "Email",
    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Email),
    isError = state.error is AppError.Validation && (state.error as AppError.Validation).field == "email",
    errorMessage = (state.error as? AppError.Validation)?.takeIf { it.field == "email" }?.message,
)
```

## Acceptance Criteria

- [ ] Field corners are visibly rounded (≥ 16dp) in default and focused states.
- [ ] Error message appears below field without shifting other content.
- [ ] Focused state clearly uses brand primary color on the border.
- [ ] Password field hides characters when `visualTransformation = PasswordVisualTransformation()`.
- [ ] Disabled state is visually distinct.
