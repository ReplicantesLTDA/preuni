package com.preuni.shared.domain.learn

enum class NodeState { LOCKED, ACTIVE, COMPLETED }

data class ModuleNode(
    val id: String,
    val index: Int,
    val title: String,
    val state: NodeState,
    val xpReward: Int = 10,
)

fun generateStubModules(count: Int = 15): List<ModuleNode> =
    List(count) { i ->
        ModuleNode(
            id = "module-$i",
            index = i,
            title = "Módulo ${i + 1}",
            state = if (i == 0) NodeState.ACTIVE else NodeState.LOCKED,
        )
    }
