package com.preuni.shared.presentation.common

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import com.preuni.shared.domain.error.AppError

/**
 * Wraps [content] in a scaffold that handles [AppError] display:
 * - [AppError.NetworkError] → snackbar
 * - [AppError.Unauthorized] → full-screen session-expired overlay with "Sign in again" CTA
 * - Other errors → snackbar
 */
@Composable
fun GlobalErrorHandler(
    error: AppError?,
    onSignInAgain: () -> Unit,
    content: @Composable () -> Unit,
) {
    val snackbarHostState = remember { SnackbarHostState() }

    // Snackbar for transient errors
    LaunchedEffect(error) {
        if (error != null && error !is AppError.Unauthorized) {
            val msg = when (error) {
                is AppError.NetworkError -> error.message
                is AppError.Conflict -> error.message
                is AppError.Validation -> "${error.field}: ${error.message}"
                else -> "Something went wrong. Please try again."
            }
            snackbarHostState.showSnackbar(msg)
        }
    }

    Scaffold(snackbarHost = { SnackbarHost(snackbarHostState) }) { padding ->
        Box(modifier = Modifier.padding(padding)) {
            content()

            // Full-screen overlay for session expiry
            if (error is AppError.Unauthorized) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center,
                ) {
                    // Semi-transparent scrim
                    androidx.compose.foundation.layout.Box(
                        modifier = Modifier.fillMaxSize(),
                    ) {}

                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.Center,
                        modifier = Modifier.padding(32.dp),
                    ) {
                        Text(
                            "Session expired",
                            style = MaterialTheme.typography.headlineMedium,
                        )
                        Spacer(Modifier.height(8.dp))
                        Text(
                            "Your session has expired. Please sign in again to continue.",
                            style = MaterialTheme.typography.bodyMedium,
                            textAlign = TextAlign.Center,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                        )
                        Spacer(Modifier.height(24.dp))
                        Button(onClick = onSignInAgain) {
                            Text("Sign in again")
                        }
                    }
                }
            }
        }
    }
}
