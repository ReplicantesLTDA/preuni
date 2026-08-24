"""T104: per-competency golden slice against FakeProvider.

Sanity-only: verifies the harness reports MAE ≤ 60 per competency on fake
provider canned outputs, which by definition always match the expected scores.
This gate runs on every PR (make test-golden-fake) to detect regressions in
parsing, aggregation, or harness arithmetic.
"""

from __future__ import annotations

import pytest

from tests.golden.harness import Example, run_examples

FAKE_ESSAY = (
    "A violência urbana no Brasil tem raízes históricas na desigualdade social. "
    "Comunidades periféricas sofrem com ausência de políticas públicas eficazes. "
    "A educação de qualidade é fundamental para quebrar o ciclo da violência. "
    "Programas sociais integrados reduzem criminalidade e promovem cidadania plena. "
    "O investimento em segurança pública deve vir acompanhado de medidas sociais. "
    "O diálogo entre Estado e sociedade fortalece a democracia e a paz social. "
    "Somente com políticas integradas o Brasil superará o problema da violência. "
) * 4


def _make_fake_example(slug: str, c1: int, c2: int, c3: int, c4: int, c5: int) -> Example:
    return Example(
        slug=slug,
        essay_text=FAKE_ESSAY,
        prompt_theme_title="Violência Urbana",
        prompt_theme_context="Causas e consequências da violência nas cidades.",
        motivational_texts="",
        expected_c1=c1,
        expected_c2=c2,
        expected_c3=c3,
        expected_c4=c4,
        expected_c5=c5,
        expected_total=c1 + c2 + c3 + c4 + c5,
    )


def _fake_response(c1: int, c2: int, c3: int, c4: int, c5: int) -> dict:
    excerpt = FAKE_ESSAY[:30]
    return {
        "eliminatory_flags": [],
        "annulment_reason_pt_br": None,
        "c1": {
            "score": c1,
            "excerpt": excerpt,
            "justification_pt_br": "Ok.",
            "improvement_path_pt_br": "Melhore.",
        },
        "c2": {
            "score": c2,
            "excerpt": excerpt,
            "justification_pt_br": "Ok.",
            "improvement_path_pt_br": "Melhore.",
        },
        "c3": {
            "score": c3,
            "excerpt": excerpt,
            "justification_pt_br": "Ok.",
            "improvement_path_pt_br": "Melhore.",
        },
        "c4": {
            "score": c4,
            "excerpt": excerpt,
            "justification_pt_br": "Ok.",
            "improvement_path_pt_br": "Melhore.",
        },
        "c5": {
            "score": c5,
            "excerpt": excerpt,
            "justification_pt_br": "Ok.",
            "improvement_path_pt_br": "Melhore.",
        },
    }


@pytest.mark.asyncio
async def test_per_competency_mae_fake_provider_passthrough() -> None:
    """FakeProvider returns exact-match scores → MAE = 0 per competency."""
    from src.corrector.llm.fake import FakeProvider

    n = 5
    examples = [
        _make_fake_example(f"ex-{i}", c1=160, c2=120, c3=120, c4=160, c5=80) for i in range(n)
    ]
    responses = [_fake_response(160, 120, 120, 160, 80) for _ in range(n)]
    provider = FakeProvider(responses=responses)

    report = await run_examples(examples, provider=provider)

    assert report.overall.mae_per_competency <= 60, (
        f"Per-competency MAE {report.overall.mae_per_competency} exceeds 60 on fake provider"
    )
    assert report.overall.mae_per_competency == 0.0


@pytest.mark.asyncio
async def test_per_competency_mae_with_known_offset() -> None:
    """FakeProvider returns +40 on c1-c4 → per-comp MAE = 32 ≤ 60."""
    from src.corrector.llm.fake import FakeProvider

    n = 5
    examples = [
        _make_fake_example(f"off-{i}", c1=120, c2=120, c3=120, c4=120, c5=120) for i in range(n)
    ]
    # c1-c4 offset by +40, c5 exact
    responses = [_fake_response(160, 160, 160, 160, 120) for _ in range(n)]
    provider = FakeProvider(responses=responses)

    report = await run_examples(examples, provider=provider)

    assert report.overall.mae_per_competency <= 60, (
        f"Per-competency MAE {report.overall.mae_per_competency} should be ≤ 60 (offset=40)"
    )
