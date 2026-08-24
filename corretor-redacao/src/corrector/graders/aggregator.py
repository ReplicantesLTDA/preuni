"""Grader-pass aggregator.

MVP: identity passthrough (single grader pass). Multi-grader-ready: future
implementation will compute the ENEM aggregation rule (average of two closest
passes, third pass triggered on > 100 total or > 80 per-comp divergence).
"""

from __future__ import annotations

from src.corrector.graders.single_grader import SingleGraderResult


def aggregate_single(result: SingleGraderResult) -> SingleGraderResult:
    """Identity in the MVP."""
    return result


__all__ = ["aggregate_single"]
