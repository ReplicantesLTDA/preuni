package com.preuni.shared.presentation.auth

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
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
        Text(
            text = "PreUni",
            style = MaterialTheme.typography.headlineLarge,
            color = MaterialTheme.colorScheme.primary,
        )

        Spacer(Modifier.height(40.dp))

        OutlinedTextField(
            value = state.emailOrUsername,
            onValueChange = { store.accept(LoginStore.Intent.UpdateEmailOrUsername(it)) },
            label = { Text("Email or username") },
            singleLine = true,
            keyboardOptions = KeyboardOptions(
                keyboardType = KeyboardType.Email,
                imeAction = ImeAction.Next,
            ),
            isError = state.error is AppError.Validation &&
                (state.error as AppError.Validation).field == "emailOrUsername",
            supportingText = {
                val err = state.error
                if (err is AppError.Validation && err.field == "emailOrUsername") {
                    Text(err.message, color = MaterialTheme.colorScheme.error)
                }
            },
            modifier = Modifier.fillMaxWidth(),
        )

        Spacer(Modifier.height(8.dp))

        OutlinedTextField(
            value = state.password,
            onValueChange = { store.accept(LoginStore.Intent.UpdatePassword(it)) },
            label = { Text("Password") },
            singleLine = true,
            visualTransformation = if (passwordVisible) VisualTransformation.None
                                   else PasswordVisualTransformation(),
            keyboardOptions = KeyboardOptions(
                keyboardType = KeyboardType.Password,
                imeAction = ImeAction.Done,
            ),
            keyboardActions = KeyboardActions(onDone = { store.accept(LoginStore.Intent.Submit) }),
            trailingIcon = {
                TextButton(onClick = { passwordVisible = !passwordVisible }) {
                    Text(if (passwordVisible) "Hide" else "Show")
                }
            },
            isError = state.error is AppError.Validation &&
                (state.error as AppError.Validation).field == "password",
            supportingText = {
                val err = state.error
                if (err is AppError.Validation && err.field == "password") {
                    Text(err.message, color = MaterialTheme.colorScheme.error)
                }
            },
            modifier = Modifier.fillMaxWidth(),
        )

        // Generic (non-field) error banner
        val err = state.error
        if (err != null && err !is AppError.Validation) {
            Spacer(Modifier.height(8.dp))
            Text(
                text = when (err) {
                    is AppError.Unauthorized -> "Invalid credentials. Please try again."
                    is AppError.NetworkError -> "No internet connection."
                    else -> "Something went wrong. Please try again."
                },
                color = MaterialTheme.colorScheme.error,
                style = MaterialTheme.typography.bodySmall,
            )
        }

        Spacer(Modifier.height(24.dp))

        Button(
            onClick = { store.accept(LoginStore.Intent.Submit) },
            enabled = !state.isLoading,
            modifier = Modifier.fillMaxWidth(),
        ) {
            if (state.isLoading) {
                CircularProgressIndicator(
                    modifier = Modifier.height(20.dp),
                    strokeWidth = 2.dp,
                )
            } else {
                Text("Sign in")
            }
        }

        Spacer(Modifier.height(12.dp))

        TextButton(onClick = onSignInWithCode) {
            Text("Sign in with email code")
        }

        TextButton(onClick = onCreateAccount) {
            Text("Create account")
        }
    }
}
