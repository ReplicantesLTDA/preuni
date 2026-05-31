package com.preuni.shared.presentation.auth

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.text.input.VisualTransformation
import androidx.compose.ui.unit.dp
import com.arkivanov.mvikotlin.extensions.coroutines.stateFlow
import com.preuni.shared.ui.theme.LocalSpacing
import com.preuni.shared.domain.auth.AuthValidator
import com.preuni.shared.domain.auth.ValidationResult
import com.preuni.shared.domain.error.AppError
import com.preuni.shared.ui.components.PreuniButton
import com.preuni.shared.ui.components.PreuniTextField
import androidx.compose.ui.graphics.Color
import com.preuni.shared.ui.theme.SuccessGreen
import com.preuni.shared.ui.theme.WarningAmber

@Composable
fun RegisterScreen(
    store: RegisterStore,
    onBack: () -> Unit,
) {
    val state by store.stateFlow.collectAsState(RegisterStore.State())
    var passwordVisible by remember { mutableStateOf(false) }
    var confirmVisible by remember { mutableStateOf(false) }
    val s = LocalSpacing.current

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(horizontal = s.xl, vertical = s.xxl),
        verticalArrangement = Arrangement.Top,
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text("Criar conta", style = MaterialTheme.typography.headlineMedium)
        Spacer(Modifier.height(s.xl))

        PreuniTextField(
            value = state.displayName,
            onValueChange = { store.accept(RegisterStore.Intent.UpdateDisplayName(it)) },
            label = "Nome completo",
            keyboardOptions = KeyboardOptions(imeAction = ImeAction.Next),
        )
        Spacer(Modifier.height(s.sm))

        PreuniTextField(
            value = state.email,
            onValueChange = { store.accept(RegisterStore.Intent.UpdateEmail(it)) },
            label = "Email",
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Email, imeAction = ImeAction.Next),
            isError = state.emailError != null,
            errorMessage = state.emailError,
        )
        Spacer(Modifier.height(s.sm))

        val usernameResult = AuthValidator.validateUsername(state.username)
        val usernameIsError = state.username.isNotEmpty() && usernameResult is ValidationResult.Invalid
        PreuniTextField(
            value = state.username,
            onValueChange = { store.accept(RegisterStore.Intent.UpdateUsername(it)) },
            label = "Nome de usuário",
            keyboardOptions = KeyboardOptions(imeAction = ImeAction.Next),
            isError = usernameIsError,
            errorMessage = if (usernameIsError) (usernameResult as ValidationResult.Invalid).message else null,
            helperText = if (!usernameIsError) "Letras minúsculas, números, - e _ apenas" else null,
        )
        Spacer(Modifier.height(s.sm))

        val passwordStrength = passwordStrength(state.password)
        PreuniTextField(
            value = state.password,
            onValueChange = { store.accept(RegisterStore.Intent.UpdatePassword(it)) },
            label = "Senha",
            visualTransformation = if (passwordVisible) VisualTransformation.None else PasswordVisualTransformation(),
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Password, imeAction = ImeAction.Next),
            trailingIcon = {
                TextButton(onClick = { passwordVisible = !passwordVisible }) {
                    Text(if (passwordVisible) "Ocultar" else "Mostrar")
                }
            },
            isError = state.passwordError != null,
            errorMessage = state.passwordError,
        )
        if (state.password.isNotEmpty()) {
            LinearProgressIndicator(
                progress = { passwordStrength.fraction },
                color = passwordStrength.color,
                modifier = Modifier.fillMaxWidth().height(4.dp),
            )
        }
        Spacer(Modifier.height(s.sm))

        PreuniTextField(
            value = state.confirmPassword,
            onValueChange = { store.accept(RegisterStore.Intent.UpdateConfirmPassword(it)) },
            label = "Confirmar senha",
            visualTransformation = if (confirmVisible) VisualTransformation.None else PasswordVisualTransformation(),
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Password, imeAction = ImeAction.Done),
            trailingIcon = {
                TextButton(onClick = { confirmVisible = !confirmVisible }) {
                    Text(if (confirmVisible) "Ocultar" else "Mostrar")
                }
            },
            isError = state.confirmPasswordError != null,
            errorMessage = state.confirmPasswordError,
        )

        val globalErr = state.globalError
        if (globalErr != null) {
            Spacer(Modifier.height(s.sm))
            Text(
                text = when (globalErr) {
                    is AppError.Conflict -> globalErr.message
                    is AppError.NetworkError -> "Sem conexão com a internet."
                    else -> "Algo deu errado. Tente novamente."
                },
                color = MaterialTheme.colorScheme.error,
                style = MaterialTheme.typography.bodySmall,
            )
        }

        Spacer(Modifier.height(s.xl))

        PreuniButton(
            text = "Criar conta",
            onClick = { store.accept(RegisterStore.Intent.Submit) },
            isLoading = state.isLoading,
        )

        TextButton(onClick = onBack) {
            Text("Já tem uma conta? Entrar")
        }
    }
}

private data class PasswordStrengthUi(val fraction: Float, val color: Color)

@Composable
private fun passwordStrength(password: String): PasswordStrengthUi {
    val hasUpper = password.any { it.isUpperCase() }
    val hasLower = password.any { it.isLowerCase() }
    val hasDigit = password.any { it.isDigit() }
    val longEnough = password.length >= 6
    val veryLong = password.length >= 12
    val score = listOf(hasUpper, hasLower, hasDigit, longEnough, veryLong).count { it }
    return when {
        score <= 2 -> PasswordStrengthUi(0.25f, MaterialTheme.colorScheme.error)
        score == 3 -> PasswordStrengthUi(0.5f, WarningAmber)
        score == 4 -> PasswordStrengthUi(0.75f, SuccessGreen)
        else -> PasswordStrengthUi(1f, MaterialTheme.colorScheme.primary)
    }
}
