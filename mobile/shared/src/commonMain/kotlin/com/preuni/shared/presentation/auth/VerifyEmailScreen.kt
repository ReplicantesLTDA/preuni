package com.preuni.shared.presentation.auth

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import com.arkivanov.mvikotlin.extensions.coroutines.stateFlow
import com.preuni.shared.ui.components.MascotPlaceholder
import com.preuni.shared.ui.components.PreuniButton
import com.preuni.shared.ui.components.PreuniTextField
import com.preuni.shared.ui.theme.LocalSpacing
import kotlinx.coroutines.delay

@Composable
fun VerifyEmailScreen(
    store: VerifyEmailStore,
    email: String,
    onVerified: () -> Unit,
) {
    val state by store.stateFlow.collectAsState(VerifyEmailStore.State())
    var resendCooldown by remember { mutableIntStateOf(0) }

    LaunchedEffect(resendCooldown) {
        if (resendCooldown > 0) {
            delay(1_000)
            resendCooldown--
        }
    }

    LaunchedEffect(state.verified) {
        if (state.verified) onVerified()
    }

    val s = LocalSpacing.current

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(horizontal = s.xl),
        verticalArrangement = Arrangement.Center,
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        MascotPlaceholder()
        Spacer(Modifier.height(s.lg))
        Text("Verifique seu email", style = MaterialTheme.typography.headlineMedium)
        Spacer(Modifier.height(s.md))
        Text(
            text = "Enviamos um código para $email. Insira-o abaixo.",
            style = MaterialTheme.typography.bodyMedium,
            textAlign = TextAlign.Center,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
        Spacer(Modifier.height(s.xxl))

        PreuniTextField(
            value = state.code,
            onValueChange = { value ->
                if (value.length <= 6 && value.all { it.isDigit() }) {
                    store.accept(VerifyEmailStore.Intent.UpdateCode(value))
                }
            },
            label = "Código de verificação",
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.NumberPassword),
            isError = state.error != null,
            errorMessage = state.error,
        )
        Spacer(Modifier.height(s.xl))

        PreuniButton(
            text = "Verificar",
            onClick = { store.accept(VerifyEmailStore.Intent.Submit) },
            enabled = state.code.length == 6,
            isLoading = state.isLoading,
        )

        Spacer(Modifier.height(s.md))

        if (resendCooldown > 0) {
            Text(
                text = "Reenviar código em ${resendCooldown}s",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        } else {
            TextButton(onClick = {
                store.accept(VerifyEmailStore.Intent.Resend)
                resendCooldown = 60
            }) {
                Text("Reenviar código")
            }
        }
    }
}
