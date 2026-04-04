package com.preuni.shared.presentation.profile

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.preuni.shared.domain.auth.AuthValidator
import com.preuni.shared.domain.auth.ValidationResult

@Composable
fun EditUsernameScreen(
    currentUsername: String,
    onSave: (String) -> Unit,
    onBack: () -> Unit,
) {
    var username by remember { mutableStateOf(currentUsername) }
    val validationResult = AuthValidator.validateUsername(username)
    val isValid = validationResult is ValidationResult.Valid

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(horizontal = 24.dp, vertical = 32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text("Edit username", style = MaterialTheme.typography.headlineMedium)
        Spacer(Modifier.height(24.dp))

        OutlinedTextField(
            value = username,
            onValueChange = { if (it.length <= 30) username = it },
            label = { Text("Username") },
            singleLine = true,
            isError = username.isNotEmpty() && !isValid,
            supportingText = {
                val errorMsg = if (username.isNotEmpty() && validationResult is ValidationResult.Invalid)
                    validationResult.message
                else
                    "Only lowercase letters, numbers, - and _"
                val isError = username.isNotEmpty() && !isValid
                Text(
                    errorMsg,
                    color = if (isError) MaterialTheme.colorScheme.error
                            else MaterialTheme.colorScheme.onSurfaceVariant,
                )
                // Character counter
            },
            trailingIcon = {
                Text(
                    "${username.length}/30",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            },
            modifier = Modifier.fillMaxWidth(),
        )

        Spacer(Modifier.height(24.dp))

        Button(
            onClick = { onSave(username) },
            enabled = isValid && username != currentUsername,
            modifier = Modifier.fillMaxWidth(),
        ) {
            Text("Save")
        }
    }
}
