package com.preuni.shared.presentation.navigation

import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.EmojiEvents
import androidx.compose.material.icons.filled.Group
import androidx.compose.material.icons.filled.MenuBook
import androidx.compose.material.icons.filled.Person
import androidx.compose.material.icons.outlined.Edit
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.NavigationBarItemDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.vector.ImageVector

/**
 * Wireframe destinations. Order matches the bottom-tab layout: Trilha first
 * (default after login), Redação center for the writing journey, then the
 * social/identity surfaces.
 */
enum class BottomTab(val label: String, val icon: ImageVector) {
    LEARN("Trilha", Icons.Filled.MenuBook),
    SIMULATE("Redação", Icons.Outlined.Edit),
    FRIENDS("Amigos", Icons.Filled.Group),
    LEAGUE("Liga", Icons.Filled.EmojiEvents),
    PROFILE("Perfil", Icons.Filled.Person),
}

@Composable
fun BottomNavigation(
    selectedTab: BottomTab,
    onTabSelected: (BottomTab) -> Unit,
) {
    NavigationBar {
        BottomTab.entries.forEach { tab ->
            NavigationBarItem(
                selected = selectedTab == tab,
                onClick = { onTabSelected(tab) },
                icon = { Icon(tab.icon, contentDescription = tab.label) },
                label = { Text(tab.label) },
                colors = NavigationBarItemDefaults.colors(
                    indicatorColor = MaterialTheme.colorScheme.secondaryContainer,
                    selectedIconColor = MaterialTheme.colorScheme.onSecondaryContainer,
                    selectedTextColor = MaterialTheme.colorScheme.onSurface,
                    unselectedIconColor = MaterialTheme.colorScheme.onSurfaceVariant,
                    unselectedTextColor = MaterialTheme.colorScheme.onSurfaceVariant,
                ),
            )
        }
    }
}
