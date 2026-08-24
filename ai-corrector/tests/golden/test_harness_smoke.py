"""T044: golden harness smoke against FakeProvider.

Verifies: harness ingests an Example dir, runs the pipeline, computes overall +
per-band metrics + INEP per-essay tolerance gate, emits a Report.
"""

from __future__ import annotations

import json
import pytest
from pathlib import Path
from textwrap import dedent

from src.corrector.llm.fake import FakeProvider
from tests.golden.harness import Example, Report, run_examples

ESSAY = (
    "A inclusão digital de pessoas idosas demanda ação coordenada do Estado.\n"
    "Você pode notar, no cotidiano, que muitos avós têm dificuldade para utilizar\n"
    "aplicativos de banco e plataformas governamentais essenciais para a cidadania.\n"
    "Diante disso, propõe-se uma estratégia coordenada de inclusão, com cursos\n"
    "gratuitos em telecentros e parcerias com escolas técnicas para mediação\n"
    "digital intergeracional efetiva entre famílias brasileiras hoje em dia.\n"
    "Portanto, cabe ao MEC articular um programa nacional de letramento digital."
)


def _write_example(root: Path, slug: str, *, expected_total: int) -> None:
    d = root / slug
    d.mkdir(parents=True)
    (d / "essay.txt").write_text(ESSAY, encoding="utf-8")
    (d / "prompt_theme.txt").write_text(
        dedent("""\
            Inclusão digital de idosos no Brasil
            ---
            A democratização das ferramentas digitais não atingiu uniformemente a
            população idosa, gerando exclusão funcional do exercício de direitos.
        """),
        encoding="utf-8",
    )
    (d / "motivational_texts.txt").write_text("", encoding="utf-8")
    per_comp = expected_total // 5
    (d / "expected.json").write_text(
        json.dumps(
            {
                "c1": per_comp,
                "c2": per_comp,
                "c3": per_comp,
                "c4": per_comp,
                "c5": per_comp,
                "total": expected_total,
            },
            ensure_ascii=False,
        ),
        encoding="utf-8",
    )


@pytest.mark.asyncio
async def test_harness_runs_against_fake_provider(tmp_path: Path) -> None:
    root = tmp_path / "examples"
    _write_example(root, "0001", expected_total=800)
    _write_example(root, "0002", expected_total=400)
    _write_example(root, "0003", expected_total=200)

    # Fake provider returns a perfect-match output for each example.
    def make_output(total: int) -> dict:
        per = total // 5
        comp = {}
        for code in ("c1", "c2", "c3", "c4", "c5"):
            entry: dict[str, object] = {
                "score": per,
                "excerpt": "A inclusão digital de pessoas idosas",
                "justification_pt_br": "ok",
            }
            if per < 200:
                entry["improvement_path_pt_br"] = "revisar."
            comp[code] = entry
        return {
            "eliminatory_flags": [],
            "competencies": comp,
            "final_score": total,
        }

    fake = FakeProvider(
        responses=[make_output(800), make_output(400), make_output(200)],
    )

    examples = Example.iter_dir(root)
    report = await run_examples(examples, provider=fake)

    assert isinstance(report, Report)
    assert report.n_examples == 3
    # Perfect match against the human reference => MAE 0, hit-rate 100%
    assert report.overall.mae_total == 0
    assert report.overall.mae_per_competency == 0
    assert report.overall.hit_rate_within_120 == 1.0
    # Bands: 0–400 (1 ex), 401–700 (0 ex), 701–1000 (1 ex). 200 falls in 0–400.
    # 800 falls in 701–1000. 400 falls in 0–400.
    assert report.bands["0-400"].n == 2
    assert report.bands["401-700"].n == 0
    assert report.bands["701-1000"].n == 1


@pytest.mark.asyncio
async def test_harness_under_powered_bands_marked_as_such(tmp_path: Path) -> None:
    root = tmp_path / "examples"
    _write_example(root, "0001", expected_total=800)
    fake = FakeProvider(
        responses=[
            {
                "eliminatory_flags": [],
                "competencies": {
                    code: {
                        "score": 160,
                        "excerpt": "A inclusão digital de pessoas idosas",
                        "justification_pt_br": "ok",
                        "improvement_path_pt_br": "x",
                    }
                    for code in ("c1", "c2", "c3", "c4", "c5")
                },
                "final_score": 800,
            }
        ]
    )
    report = await run_examples(Example.iter_dir(root), provider=fake)
    # Underpowered = fewer than MIN_BAND_SAMPLES (3) per harness module
    assert report.bands["0-400"].underpowered is True
    assert report.bands["401-700"].underpowered is True
    assert report.bands["701-1000"].underpowered is True  # only 1 sample, below MIN_BAND_SAMPLES
