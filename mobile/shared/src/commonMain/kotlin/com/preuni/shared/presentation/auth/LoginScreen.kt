package com.preuni.shared.presentation.auth

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
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
import com.preuni.shared.domain.error.AppError
import com.preuni.shared.ui.components.PreuniButton
import com.preuni.shared.ui.components.PreuniTextField

@Composable
fun LoginScreen(
    store: LoginStore,
    onSignInWithCode: () -> Unit,
    onCreateAccount: () -> Unit,
) {
    val state by store.stateFlow.collectAsState(LoginStore.State())
    var passwordVisible by remember { mutableStateOf(false) }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(horizontal = 24.dp),
        verticalArrangement = Arrangement.Center,
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        // Brand header
        Text(
            text = "PreUni",
            style = MaterialTheme.typography.displaySmall,
            color = MaterialTheme.colorScheme.primary,
        )
        Text(
            text = "Prepare-se para o ENEM.",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )

        Spacer(Modifier.height(48.dp))

        PreuniTextField(
            value = state.emailOrUsername,
            onValueChange = { store.accept(LoginStore.Intent.UpdateEmailOrUsername(it)) },
            label = "Email ou usuário",
            keyboardOptions = KeyboardOptions(
                keyboardType = KeyboardType.Email,
                imeAction = ImeAction.Next,
            ),
            isError = state.error is AppError.Validation &&
                (state.error as AppError.Validation).field == "emailOrUsername",
            errorMessage = (state.error as? AppError.Validation)
                ?.takeIf { it.field == "emailOrUsername" }?.message,
        )

        Spacer(Modifier.height(8.dp))

        PreuniTextField(
            value = state.password,
            onValueChange = { store.accept(LoginStore.Intent.UpdatePassword(it)) },
            label = "Senha",
            visualTransformation = if (passwordVisible) VisualTransformation.None
                                   else PasswordVisualTransformation(),
            keyboardOptions = KeyboardOptions(
                keyboardType = KeyboardType.Password,
                imeAction = ImeAction.Done,
            ),
            keyboardActions = KeyboardActions(onDone = { store.accept(LoginStore.Intent.Submit) }),
            trailingIcon = {
                TextButton(onClick = { passwordVisible = !passwordVisible }) {
                    Text(if (passwordVisible) "Ocultar" else "Mostrar")
                }
            },
            isError = state.error is AppError.Validation &&
                (state.error as AppError.Validation).field == "password",
            errorMessage = (state.error as? AppError.Validation)
                ?.takeIf { it.field == "password" }?.message,
        )

        val err = state.error
        if (err != null && err !is AppError.Validation) {
            Spacer(Modifier.height(4.dp))
            Text(
                text = when (err) {
                    is AppError.Unauthorized -> "Credenciais inválidas. Tente novamente."
                    is AppError.NetworkError -> "Sem conexão com a internet."
                    else -> "Algo deu errado. Tente novamente."
                },
                color = MaterialTheme.colorScheme.error,
                style = MaterialTheme.typography.bodySmall,
            )
        }

        Spacer(Modifier.height(24.dp))

        PreuniButton(
            text = "Entrar",
            onClick = { store.accept(LoginStore.Intent.Submit) },
            isLoading = state.isLoading,
        )

        Spacer(Modifier.height(8.dp))

        TextButton(onClick = onSignInWithCode) {
            Text("Entrar com código")
        }

        TextButton(onClick = onCreateAccount) {
            Text("Criar conta")
        }
    }
}
