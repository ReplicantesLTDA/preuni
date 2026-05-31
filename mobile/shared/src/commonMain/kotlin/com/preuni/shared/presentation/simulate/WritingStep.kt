package com.preuni.shared.presentation.simulate

import androidx.compose.runtime.Immutable

/**
 * Minimal model for the Redação guided journey UI. Backend integration
 * arrives in a follow-up feature; for now we represent the stages and
 * lock state so the user can scan their progress.
 */
@Immutable
data class WritingStep(
    val key: Key,
    val title: String,
    val summary: String,
    val status: Status,
) {
    enum class Key { Prompt, Planning, Writing, Review, Result }
    enum class Status { Done, Current, Locked }
}

/** Default stages shown when no backend data is available. */
val DefaultWritingFlow: List<WritingStep> = listOf(
    WritingStep(
        key = WritingStep.Key.Prompt,
        title = "Tema da redação",
        summary = "Leia a proposta e os textos de apoio com calma.",
        status = WritingStep.Status.Current,
    ),
    WritingStep(
        key = WritingStep.Key.Planning,
        title = "Planejamento",
        summary = "Organize tese, argumentos e proposta de intervenção.",
        status = WritingStep.Status.Locked,
    ),
    WritingStep(
        key = WritingStep.Key.Writing,
        title = "Escrita",
        summary = "Escreva sua redação respeitando a estrutura dissertativa.",
        status = WritingStep.Status.Locked,
    ),
    WritingStep(
        key = WritingStep.Key.Review,
        title = "Revisão",
        summary = "Releia em busca de coesão, coerência e norma culta.",
        status = WritingStep.Status.Locked,
    ),
    WritingStep(
        key = WritingStep.Key.Result,
        title = "Nota e feedback",
        summary = "Veja sua nota nas cinco competências e o próximo passo.",
        status = WritingStep.Status.Locked,
    ),
)
