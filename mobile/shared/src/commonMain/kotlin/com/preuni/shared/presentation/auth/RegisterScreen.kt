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
import androidx.compose.material3.Button
import androidx.compose.material3.LinearProgressIndicator
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
import com.preuni.shared.domain.auth.AuthValidator
import com.preuni.shared.domain.auth.ValidationResult
import com.preuni.shared.domain.error.AppError

@Composable
fun RegisterScreen(
    store: RegisterStore,
    onBack: () -> Unit,
) {
    val state by store.stateFlow.collectAsState(RegisterStore.State())
    var passwordVisible by remember { mutableStateOf(false) }
    var confirmVisible by remember { mutableStateOf(false) }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(horizontal = 24.dp, vertical = 32.dp),
        verticalArrangement = Arrangement.Top,
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text("Create account", style = MaterialTheme.typography.headlineMedium)
        Spacer(Modifier.height(24.dp))

        OutlinedTextField(
            value = state.displayName,
            onValueChange = { store.accept(RegisterStore.Intent.UpdateDisplayName(it)) },
            label = { Text("Full name") },
            singleLine = true,
            keyboardOptions = KeyboardOptions(imeAction = ImeAction.Next),
            modifier = Modifier.fillMaxWidth(),
        )
        Spacer(Modifier.height(8.dp))

        OutlinedTextField(
            value = state.email,
            onValueChange = { store.accept(RegisterStore.Intent.UpdateEmail(it)) },
            label = { Text("Email") },
            singleLine = true,
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Email, imeAction = ImeAction.Next),
            isError = state.emailError != null,
            supportingText = { state.emailError?.let { Text(it, color = MaterialTheme.colorScheme.error) } },
            modifier = Modifier.fillMaxWidth(),
        )
        Spacer(Modifier.height(8.dp))

        val usernameResult = AuthValidator.validateUsername(state.username)
        OutlinedTextField(
            value = state.username,
            onValueChange = { store.accept(RegisterStore.Intent.UpdateUsername(it)) },
            label = { Text("Username") },
            singleLine = true,
            keyboardOptions = KeyboardOptions(imeAction = ImeAction.Next),
            isError = state.username.isNotEmpty() && usernameResult is ValidationResult.Invalid,
            supportingText = {
                val msg = when {
                    state.username.isNotEmpty() && usernameResult is ValidationResult.Invalid ->
                        usernameResult.message
                    else -> "Lowercase letters, numbers, - and _ only"
                }
                Text(msg, color = if (usernameResult is ValidationResult.Invalid && state.username.isNotEmpty())
                    MaterialTheme.colorScheme.error else MaterialTheme.colorScheme.onSurfaceVariant)
            },
            modifier = Modifier.fillMaxWidth(),
        )
        Spacer(Modifier.height(8.dp))

        // Password strength indicator
        val passwordStrength = passwordStrength(state.password)
        OutlinedTextField(
            value = state.password,
            onValueChange = { store.accept(RegisterStore.Intent.UpdatePassword(it)) },
            label = { Text("Password") },
            singleLine = true,
            visualTransformation = if (passwordVisible) VisualTransformation.None else PasswordVisualTransformation(),
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Password, imeAction = ImeAction.Next),
            trailingIcon = {
                TextButton(onClick = { passwordVisible = !passwordVisible }) {
                    Text(if (passwordVisible) "Hide" else "Show")
                }
            },
            isError = state.passwordError != null,
            supportingText = { state.passwordError?.let { Text(it, color = MaterialTheme.colorScheme.error) } },
            modifier = Modifier.fillMaxWidth(),
        )
        if (state.password.isNotEmpty()) {
            LinearProgressIndicator(
                progress = { passwordStrength.fraction },
                color = passwordStrength.color,
                modifier = Modifier.fillMaxWidth().height(4.dp),
            )
        }
        Spacer(Modifier.height(8.dp))

        OutlinedTextField(
            value = state.confirmPassword,
            onValueChange = { store.accept(RegisterStore.Intent.UpdateConfirmPassword(it)) },
            label = { Text("Confirm password") },
            singleLine = true,
            visualTransformation = if (confirmVisible) VisualTransformation.None else PasswordVisualTransformation(),
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Password, imeAction = ImeAction.Done),
            trailingIcon = {
                TextButton(onClick = { confirmVisible = !confirmVisible }) {
                    Text(if (confirmVisible) "Hide" else "Show")
                }
            },
            isError = state.confirmPasswordError != null,
            supportingText = { state.confirmPasswordError?.let { Text(it, color = MaterialTheme.colorScheme.error) } },
            modifier = Modifier.fillMaxWidth(),
        )

        val globalErr = state.globalError
        if (globalErr != null) {
            Spacer(Modifier.height(8.dp))
            Text(
                text = when (globalErr) {
                    is AppError.Conflict -> globalErr.message
                    is AppError.NetworkError -> "No internet connection."
                    else -> "Something went wrong. Please try again."
                },
                color = MaterialTheme.colorScheme.error,
                style = MaterialTheme.typography.bodySmall,
            )
        }

        Spacer(Modifier.height(24.dp))

        Button(
            onClick = { store.accept(RegisterStore.Intent.Submit) },
            enabled = !state.isLoading,
            modifier = Modifier.fillMaxWidth(),
        ) {
            Text("Create account")
        }

        TextButton(onClick = onBack) {
            Text("Already have an account? Sign in")
        }
    }
}

private data class PasswordStrengthUi(val fraction: Float, val color: androidx.compose.ui.graphics.Color)

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
        score == 3 -> PasswordStrengthUi(0.5f, androidx.compose.ui.graphics.Color(0xFFF59E0B))
        score == 4 -> PasswordStrengthUi(0.75f, androidx.compose.ui.graphics.Color(0xFF10B981))
        else -> PasswordStrengthUi(1f, MaterialTheme.colorScheme.primary)
    }
}
