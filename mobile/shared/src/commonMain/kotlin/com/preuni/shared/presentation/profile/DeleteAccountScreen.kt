package com.preuni.shared.presentation.profile

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp

@Composable
fun DeleteAccountScreen(
    onConfirm: () -> Unit,
    onBack: () -> Unit,
    isLoading: Boolean = false,
) {
    var confirmText by remember { mutableStateOf("") }
    val canConfirm = confirmText == "DELETE"

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(horizontal = 24.dp, vertical = 32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text("Delete account", style = MaterialTheme.typography.headlineLarge,
            color = MaterialTheme.colorScheme.error)
        Spacer(Modifier.height(24.dp))

        // GDPR disclosure
        Text(
            "What will be deleted:",
            style = MaterialTheme.typography.titleSmall,
            fontWeight = FontWeight.Bold,
        )
        Spacer(Modifier.height(8.dp))
        GdprItem("Your email address and password")
        GdprItem("Your profile photo")
        GdprItem("Your XP and streak data")

        Spacer(Modifier.height(16.dp))

        Text(
            "What is anonymized (not deleted):",
            style = MaterialTheme.typography.titleSmall,
            fontWeight = FontWeight.Bold,
        )
        Spacer(Modifier.height(8.dp))
        GdprItem("Your learning history (aggregated, no personal data)")
        GdprItem("Simulation results (anonymized for research)")

        Spacer(Modifier.height(16.dp))
        Text(
            "⚠ Under Brazilian LGPD and EU GDPR, some data is retained for legal obligations (taxes, compliance) and cannot be erased immediately.",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )

        Spacer(Modifier.height(24.dp))
        HorizontalDivider()
        Spacer(Modifier.height(24.dp))

        Text(
            "Type DELETE to confirm",
            style = MaterialTheme.typography.bodyMedium,
            fontWeight = FontWeight.SemiBold,
        )
        Spacer(Modifier.height(8.dp))

        OutlinedTextField(
            value = confirmText,
            onValueChange = { confirmText = it },
            label = { Text("Type DELETE") },
            singleLine = true,
            modifier = Modifier.fillMaxWidth(),
        )

        Spacer(Modifier.height(24.dp))

        Button(
            onClick = onConfirm,
            enabled = canConfirm && !isLoading,
            colors = ButtonDefaults.buttonColors(
                containerColor = MaterialTheme.colorScheme.error,
                contentColor = MaterialTheme.colorScheme.onError,
            ),
            modifier = Modifier.fillMaxWidth(),
        ) {
            Text("Permanently delete account")
        }

        Spacer(Modifier.height(8.dp))

        TextButton(onClick = onBack) {
            Text("Cancel")
        }
    }
}

@Composable
private fun GdprItem(text: String) {
    Text(
        "• $text",
        style = MaterialTheme.typography.bodySmall,
        color = MaterialTheme.colorScheme.onSurfaceVariant,
        modifier = Modifier.fillMaxWidth().padding(start = 8.dp),
    )
}
