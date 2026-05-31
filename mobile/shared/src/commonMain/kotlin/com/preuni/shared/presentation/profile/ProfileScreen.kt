package com.preuni.shared.presentation.profile

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
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
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.arkivanov.mvikotlin.extensions.coroutines.stateFlow
import com.preuni.shared.ui.components.SectionHeader
import com.preuni.shared.ui.theme.LocalSpacing

@Composable
fun ProfileScreen(
    store: ProfileStore,
    onEditUsername: () -> Unit,
    onEditPassword: () -> Unit,
    onChangeEmail: () -> Unit,
    onChangeTrack: () -> Unit,
    onDeleteAccount: () -> Unit,
    onLogout: () -> Unit,
    onSettings: () -> Unit,
) {
    val state by store.stateFlow.collectAsState(ProfileStore.State())
    val student = state.student
    val s = LocalSpacing.current

    LaunchedEffect(Unit) {
        store.accept(ProfileStore.Intent.LoadProfile)
    }

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

    if (state.error != null && student == null) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = s.xl),
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
            Spacer(Modifier.height(s.md))
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
            .verticalScroll(rememberScrollState()),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        // Top bar with settings affordance on the right.
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = s.sm, vertical = s.xs),
            horizontalArrangement = Arrangement.End,
        ) {
            IconButton(onClick = onSettings) {
                Icon(Icons.Filled.Settings, contentDescription = "Ajustes")
            }
        }

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = s.xl),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            AvatarSection(
                avatarUrl = student?.avatarUrl,
                onEditAvatar = { store.accept(ProfileStore.Intent.UploadAvatar) },
            )

            Spacer(Modifier.height(s.lg))

            Text(
                text = student?.displayName ?: "",
                style = MaterialTheme.typography.titleLarge,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis,
            )
            Text(
                text = "@${student?.username ?: ""}",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            Text(
                text = student?.email ?: "",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )

            Spacer(Modifier.height(s.md))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceEvenly,
            ) {
                StatChip(label = "XP", value = (student?.xpTotal ?: 0).toString())
                StatChip(label = "Streak", value = "${student?.streakCount ?: 0} dias")
            }

            Spacer(Modifier.height(s.xl))
            HorizontalDivider()
        }

        SectionHeader(title = "Conta")

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = s.xl),
        ) {
            OutlinedButton(onClick = onEditUsername, modifier = Modifier.fillMaxWidth()) {
                Text("Alterar nome de usuário")
            }
            Spacer(Modifier.height(s.sm))
            OutlinedButton(onClick = onEditPassword, modifier = Modifier.fillMaxWidth()) {
                Text("Alterar senha")
            }
            Spacer(Modifier.height(s.sm))
            OutlinedButton(onClick = onChangeEmail, modifier = Modifier.fillMaxWidth()) {
                Text("Alterar email")
            }
            Spacer(Modifier.height(s.sm))
            OutlinedButton(onClick = onChangeTrack, modifier = Modifier.fillMaxWidth()) {
                Text("Alterar matérias")
            }
        }

        SectionHeader(title = "Sessão")

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = s.xl),
        ) {
            OutlinedButton(onClick = onLogout, modifier = Modifier.fillMaxWidth()) {
                Text("Sair")
            }
            Spacer(Modifier.height(s.sm))
            Button(
                onClick = onDeleteAccount,
                colors = ButtonDefaults.buttonColors(
                    containerColor = MaterialTheme.colorScheme.errorContainer,
                    contentColor = MaterialTheme.colorScheme.onErrorContainer,
                ),
                modifier = Modifier.fillMaxWidth(),
            ) {
                Text("Apagar conta")
            }
            Spacer(Modifier.height(s.xl))
        }
    }
}

@Composable
private fun AvatarSection(avatarUrl: String?, onEditAvatar: () -> Unit) {
    val s = LocalSpacing.current
    Surface(
        modifier = Modifier.size(88.dp).clip(CircleShape),
        color = MaterialTheme.colorScheme.primaryContainer,
        onClick = onEditAvatar,
    ) {
        Box(contentAlignment = Alignment.Center) {
            Text("✎", style = MaterialTheme.typography.titleMedium)
        }
    }
    Spacer(Modifier.height(s.xs))
}

@Composable
private fun StatChip(label: String, value: String) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Text(value, style = MaterialTheme.typography.titleMedium, color = MaterialTheme.colorScheme.primary)
        Text(label, style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
    }
}
