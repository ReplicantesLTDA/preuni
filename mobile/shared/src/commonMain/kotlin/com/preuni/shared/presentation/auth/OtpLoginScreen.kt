package com.preuni.shared.presentation.auth

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import com.arkivanov.mvikotlin.extensions.coroutines.stateFlow
import com.preuni.shared.ui.components.PreuniButton
import com.preuni.shared.ui.components.PreuniTextField

@Composable
fun OtpLoginScreen(
    store: OtpLoginStore,
    onBack: () -> Unit,
) {
    val state by store.stateFlow.collectAsState(OtpLoginStore.State())

    Column(
        modifier = Modifier
            .fillMaxSize()
            .statusBarsPadding()
            .padding(horizontal = 24.dp),
        verticalArrangement = Arrangement.Top,
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        IconButton(onClick = onBack, modifier = Modifier.align(Alignment.Start)) {
            Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Voltar")
        }

        Spacer(Modifier.height(24.dp))

        if (!state.codeSent) {
            Text("🔑", style = MaterialTheme.typography.displayMedium)
            Spacer(Modifier.height(16.dp))
            Text("Entre sem senha", style = MaterialTheme.typography.headlineMedium)
            Spacer(Modifier.height(12.dp))
            Text(
                text = "Enviaremos um código de 6 dígitos para o seu email.",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                textAlign = TextAlign.Center,
            )
            Spacer(Modifier.height(32.dp))

            PreuniTextField(
                value = state.email,
                onValueChange = { store.accept(OtpLoginStore.Intent.UpdateEmail(it)) },
                label = "Email",
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Email),
            )
            Spacer(Modifier.height(24.dp))

            PreuniButton(
                text = "Enviar código",
                onClick = { store.accept(OtpLoginStore.Intent.RequestCode) },
                enabled = state.email.isNotBlank(),
                isLoading = state.isLoading,
            )
        } else {
            Text("🔑", style = MaterialTheme.typography.displayMedium)
            Spacer(Modifier.height(16.dp))
            Text("Digite seu código", style = MaterialTheme.typography.headlineMedium)
            Spacer(Modifier.height(12.dp))
            Text(
                text = "Enviamos o código para\n${state.email}",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                textAlign = TextAlign.Center,
            )
            Spacer(Modifier.height(32.dp))

            PreuniTextField(
                value = state.otp,
                onValueChange = { v ->
                    if (v.length <= 6 && v.all { it.isDigit() }) {
                        store.accept(OtpLoginStore.Intent.UpdateOtp(v))
                    }
                },
                label = "Código de 6 dígitos",
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.NumberPassword),
                isError = state.error != null,
                errorMessage = state.error,
            )
            Spacer(Modifier.height(24.dp))

            PreuniButton(
                text = "Entrar",
                onClick = { store.accept(OtpLoginStore.Intent.VerifyCode) },
                enabled = state.otp.length == 6,
                isLoading = state.isLoading,
            )
        }
    }
}
