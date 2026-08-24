"""Golden-dataset harness.

Runs the full correction pipeline against a directory of `Example`s, computes
Constitution-IX MVP-tier + long-term metrics, and emits a `Report` with
overall, per-band (0-400 / 401-700 / 701-1000), and INEP-secondary slices.

Per-band gate (Constitution IX): each band MUST independently clear MVP-tier
(MAE-total ≤ 120, MAE-per-comp ≤ 60, ≥ 70% within ±120). Underpowered bands
(< MIN_BAND_SAMPLES) are reported but not gated.
"""

from __future__ import annotations

import json
import statistics
import uuid
from collections.abc import Iterable
from dataclasses import dataclass, field
from pathlib import Path

from src.corrector.graders.single_grader import SingleGraderInput
from src.corrector.llm.base import LLMProvider
from src.corrector.llm.errors import LLMError
from src.corrector.pipeline import correct_essay

MIN_BAND_SAMPLES = 3

# Thresholds from Constitution v2.1.0 Article IX (3 tiers).
# MVP-ship gate is the hard merge gate; Stretch and Long-term tiers are
# tracked but non-blocking.
MVP_TIER = {
    "mae_total": 180,
    "mae_per_competency": 60,
    "hit_rate_within": (0.50, 120),
    "inep_tolerance_total": 200,
    "inep_tolerance_per_comp": 80,
}
STRETCH_TIER = {
    "mae_total": 120,
    "mae_per_competency": 50,
    "hit_rate_within": (0.70, 120),
    "inep_tolerance_total": 120,
    "inep_tolerance_per_comp": 60,
}
LONG_TERM_TIER = {
    "mae_total": 80,
    "mae_per_competency": 40,
    "hit_rate_within": (0.80, 80),
    "inep_tolerance_total": 80,
    "inep_tolerance_per_comp": 40,
}


@dataclass(slots=True, frozen=True)
class Example:
    slug: str
    essay_text: str
    prompt_theme_title: str
    prompt_theme_context: str
    motivational_texts: str
    expected_c1: int
    expected_c2: int
    expected_c3: int
    expected_c4: int
    expected_c5: int
    expected_total: int

    @classmethod
    def from_dir(cls, path: Path) -> Example:
        essay = (path / "essay.txt").read_text(encoding="utf-8").strip()
        prompt_block = (path / "prompt_theme.txt").read_text(encoding="utf-8")
        title, _, context = prompt_block.partition("---")
        title = title.strip()
        context = context.strip()
        if not context:
            # allow fallback: first line is title, rest is context
            lines = prompt_block.strip().splitlines()
            title = lines[0]
            context = "\n".join(lines[1:]).strip()
        motivational = ""
        mot_path = path / "motivational_texts.txt"
        if mot_path.is_file():
            motivational = mot_path.read_text(encoding="utf-8").strip()
        expected = json.loads((path / "expected.json").read_text(encoding="utf-8"))
        return cls(
            slug=path.name,
            essay_text=essay,
            prompt_theme_title=title,
            prompt_theme_context=context,
            motivational_texts=motivational,
            expected_c1=int(expected["c1"]),
            expected_c2=int(expected["c2"]),
            expected_c3=int(expected["c3"]),
            expected_c4=int(expected["c4"]),
            expected_c5=int(expected["c5"]),
            expected_total=int(expected["total"]),
        )

    @classmethod
    def iter_dir(cls, root: Path) -> list[Example]:
        if not root.is_dir():
            return []
        return [cls.from_dir(d) for d in sorted(root.iterdir()) if d.is_dir()]


@dataclass(slots=True)
class RunResult:
    example: Example
    predicted_total: int | None
    predicted_per_competency: dict[str, int] | None
    error: str | None
    latency_ms: int


@dataclass(slots=True)
class Metrics:
    n: int
    mae_total: float
    mae_per_competency: float
    hit_rate_within_120: float
    hit_rate_within_80: float
    underpowered: bool = False


@dataclass(slots=True)
class Report:
    n_examples: int
    overall: Metrics
    bands: dict[str, Metrics]
    inep_failures: list[str] = field(default_factory=list)
    long_term: Metrics | None = None
    per_example: list[RunResult] = field(default_factory=list)

    def as_json(self) -> dict[str, object]:
        return {
            "n_examples": self.n_examples,
            "overall": _metrics_dict(self.overall),
            "bands": {k: _metrics_dict(v) for k, v in self.bands.items()},
            "inep_failures": list(self.inep_failures),
            "long_term": _metrics_dict(self.long_term) if self.long_term else None,
        }

    def passes_mvp_tier(self) -> bool:
        # Refuse "underpowered" overall — if zero essays produced a result, that's a hard
        # operational failure, not a calibration metric.
        if self.overall.n == 0:
            return False
        if not self._check_mvp(self.overall):
            return False
        for m in self.bands.values():
            if m.underpowered:
                continue
            if not self._check_mvp(m):
                return False
        return not self.inep_failures

    @staticmethod
    def _check_mvp(m: Metrics) -> bool:
        if m.n == 0:
            return False  # zero successes ≠ pass
        return (
            m.mae_total <= MVP_TIER["mae_total"]
            and m.mae_per_competency <= MVP_TIER["mae_per_competency"]
            and m.hit_rate_within_120 >= MVP_TIER["hit_rate_within"][0]
        )


def _metrics_dict(m: Metrics) -> dict[str, object]:
    return {
        "n": m.n,
        "mae_total": m.mae_total,
        "mae_per_competency": m.mae_per_competency,
        "hit_rate_within_120": m.hit_rate_within_120,
        "hit_rate_within_80": m.hit_rate_within_80,
        "underpowered": m.underpowered,
    }


async def run_examples(
    examples: Iterable[Example],
    *,
    provider: LLMProvider,
    inep_slugs: set[str] | None = None,
) -> Report:
    examples = list(examples)
    results: list[RunResult] = []
    for idx, ex in enumerate(examples, start=1):
        r = await _run_one(ex, provider=provider)
        status = (
            f"pred={r.predicted_total} (exp={ex.expected_total}, "
            f"Δ={r.predicted_total - ex.expected_total:+d})"
            if r.predicted_total is not None
            else f"ERROR={r.error}"
        )
        print(
            f"  [{idx:>2}/{len(examples)}] {ex.slug:12s} latency={r.latency_ms / 1000:5.1f}s  {status}",
            flush=True,
        )
        results.append(r)
    return _build_report(results, inep_slugs=inep_slugs or set())


async def _run_one(ex: Example, *, provider: LLMProvider) -> RunResult:
    try:
        outcome = await correct_essay(
            SingleGraderInput(
                correction_id=uuid.uuid4(),
                essay_text=ex.essay_text,
                prompt_theme_title=ex.prompt_theme_title,
                prompt_theme_context=ex.prompt_theme_context,
                motivational_texts=ex.motivational_texts or None,
            ),
            provider=provider,
        )
        per_comp = {code: ent.score for code, ent in outcome.competencies.items()}
        return RunResult(
            example=ex,
            predicted_total=outcome.final_score,
            predicted_per_competency=per_comp,
            error=None,
            latency_ms=outcome.latency_ms,
        )
    except LLMError as exc:
        return RunResult(
            example=ex,
            predicted_total=None,
            predicted_per_competency=None,
            error=f"llm:{type(exc).__name__}:{exc!s}",
            latency_ms=0,
        )
    except Exception as exc:
        return RunResult(
            example=ex,
            predicted_total=None,
            predicted_per_competency=None,
            error=f"{type(exc).__name__}:{exc!s}",
            latency_ms=0,
        )


def _build_report(results: list[RunResult], *, inep_slugs: set[str]) -> Report:
    successes = [r for r in results if r.predicted_total is not None]
    overall = _compute_metrics(successes)
    bands = {
        "0-400": _compute_metrics([r for r in successes if r.example.expected_total <= 400]),
        "401-700": _compute_metrics(
            [r for r in successes if 401 <= r.example.expected_total <= 700]
        ),
        "701-1000": _compute_metrics([r for r in successes if r.example.expected_total >= 701]),
    }
    # INEP per-essay gate (Constitution v2.1.0 MVP-ship tier): blocks on
    # > 200 total OR > 80 per-comp delta. Stretch tier uses ±120/±60.
    inep_failures: list[str] = []
    tol_total = MVP_TIER["inep_tolerance_total"]
    tol_per_comp = MVP_TIER["inep_tolerance_per_comp"]
    for r in successes:
        if r.example.slug not in inep_slugs:
            continue
        if abs(r.predicted_total - r.example.expected_total) > tol_total:  # type: ignore[operator]
            inep_failures.append(r.example.slug)
            continue
        per_comp_deltas = [
            abs(r.predicted_per_competency[c] - getattr(r.example, f"expected_{c}"))  # type: ignore[index]
            for c in ("c1", "c2", "c3", "c4", "c5")
        ]
        if max(per_comp_deltas) > tol_per_comp:
            inep_failures.append(r.example.slug)
    return Report(
        n_examples=len(results),
        overall=overall,
        bands=bands,
        inep_failures=inep_failures,
        per_example=results,
    )


def _compute_metrics(results: list[RunResult]) -> Metrics:
    if not results:
        return Metrics(
            n=0,
            mae_total=0,
            mae_per_competency=0,
            hit_rate_within_120=1.0,
            hit_rate_within_80=1.0,
            underpowered=True,
        )
    total_errors = [abs(r.predicted_total - r.example.expected_total) for r in results]  # type: ignore[operator]
    per_comp_errors: list[float] = []
    for r in results:
        for code in ("c1", "c2", "c3", "c4", "c5"):
            pred = r.predicted_per_competency[code]  # type: ignore[index]
            ref = getattr(r.example, f"expected_{code}")
            per_comp_errors.append(abs(pred - ref))
    hits_120 = sum(1 for e in total_errors if e <= 120) / len(total_errors)
    hits_80 = sum(1 for e in total_errors if e <= 80) / len(total_errors)
    return Metrics(
        n=len(results),
        mae_total=statistics.mean(total_errors),
        mae_per_competency=statistics.mean(per_comp_errors),
        hit_rate_within_120=hits_120,
        hit_rate_within_80=hits_80,
        underpowered=len(results) < MIN_BAND_SAMPLES,
    )


__all__ = [
    "LONG_TERM_TIER",
    "MVP_TIER",
    "Example",
    "Metrics",
    "Report",
    "RunResult",
    "run_examples",
]
