package com.preuni.shared.presentation.profile

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.unit.dp
import com.arkivanov.mvikotlin.extensions.coroutines.stateFlow

@Composable
fun ProfileScreen(
    store: ProfileStore,
    onEditUsername: () -> Unit,
    onEditPassword: () -> Unit,
    onChangeEmail: () -> Unit,
    onChangeTrack: () -> Unit,
    onDeleteAccount: () -> Unit,
    onLogout: () -> Unit,
) {
    val state by store.stateFlow.collectAsState(ProfileStore.State())
    val student = state.student

    LaunchedEffect(Unit) {
        store.accept(ProfileStore.Intent.LoadProfile)
    }

    // Loading state
    if (state.isLoading && student == null) {
        Column(
            modifier = Modifier.fillMaxSize(),
            verticalArrangement = Arrangement.Center,
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            CircularProgressIndicator()
        }
        return
    }

    // Error state (no student loaded yet)
    if (state.error != null && student == null) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 24.dp),
            verticalArrangement = Arrangement.Center,
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Text(
                "Não foi possível carregar o perfil.",
                color = MaterialTheme.colorScheme.error,
                style = MaterialTheme.typography.bodyMedium,
            )
            Text(
                "Verifique sua conexão e tente novamente.",
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                style = MaterialTheme.typography.bodySmall,
            )
            Spacer(Modifier.height(12.dp))
            TextButton(onClick = { store.accept(ProfileStore.Intent.Retry) }) {
                Text("Tentar novamente")
            }
        }
        return
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .statusBarsPadding()
            .verticalScroll(rememberScrollState())
            .padding(horizontal = 24.dp, vertical = 32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        // Avatar
        AvatarSection(
            avatarUrl = student?.avatarUrl,
            onEditAvatar = { store.accept(ProfileStore.Intent.UploadAvatar) },
        )

        Spacer(Modifier.height(16.dp))

        Text(student?.displayName ?: "", style = MaterialTheme.typography.titleLarge)
        Text("@${student?.username ?: ""}", style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant)
        Text(student?.email ?: "", style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant)

        Spacer(Modifier.height(12.dp))

        // XP + Streak row (read-only)
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceEvenly,
        ) {
            StatChip(label = "XP", value = (student?.xpTotal ?: 0).toString())
            StatChip(label = "Streak", value = "${student?.streakCount ?: 0} days")
        }

        Spacer(Modifier.height(24.dp))
        HorizontalDivider()
        Spacer(Modifier.height(24.dp))

        // Action buttons
        OutlinedButton(onClick = onEditUsername, modifier = Modifier.fillMaxWidth()) {
            Text("Change username")
        }
        Spacer(Modifier.height(8.dp))
        OutlinedButton(onClick = onEditPassword, modifier = Modifier.fillMaxWidth()) {
            Text("Change password")
        }
        Spacer(Modifier.height(8.dp))
        OutlinedButton(onClick = onChangeEmail, modifier = Modifier.fillMaxWidth()) {
            Text("Change email")
        }
        Spacer(Modifier.height(8.dp))
        OutlinedButton(onClick = onChangeTrack, modifier = Modifier.fillMaxWidth()) {
            Text("Alterar matérias")
        }

        Spacer(Modifier.height(32.dp))
        HorizontalDivider()
        Spacer(Modifier.height(16.dp))

        OutlinedButton(onClick = onLogout, modifier = Modifier.fillMaxWidth()) {
            Text("Sair")
        }

        Spacer(Modifier.height(8.dp))

        Button(
            onClick = onDeleteAccount,
            colors = ButtonDefaults.buttonColors(
                containerColor = MaterialTheme.colorScheme.errorContainer,
                contentColor = MaterialTheme.colorScheme.onErrorContainer,
            ),
            modifier = Modifier.fillMaxWidth(),
        ) {
            Text("Delete account")
        }
    }
}

@Composable
private fun AvatarSection(avatarUrl: String?, onEditAvatar: () -> Unit) {
    // Placeholder — actual image loading wired when Coil/Kamel integration added
    androidx.compose.material3.Surface(
        modifier = Modifier.size(88.dp).clip(CircleShape),
        color = MaterialTheme.colorScheme.primaryContainer,
        onClick = onEditAvatar,
    ) {
        androidx.compose.foundation.layout.Box(contentAlignment = Alignment.Center) {
            Text("✎", style = MaterialTheme.typography.titleMedium)
        }
    }
}

@Composable
private fun StatChip(label: String, value: String) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Text(value, style = MaterialTheme.typography.titleMedium, color = MaterialTheme.colorScheme.primary)
        Text(label, style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
    }
}
