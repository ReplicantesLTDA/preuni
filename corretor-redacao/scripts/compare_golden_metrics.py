"""Compare golden metrics between two report JSON files.

Exit 1 if the new report regresses on overall, any band, or any INEP example
relative to the baseline. Regressions defined as:
  - mae_total increases by > 10
  - mae_per_competency increases by > 5
  - hit_rate_within_120 decreases by > 0.05
  - new inep_failures not present in baseline

Usage:
  python scripts/compare_golden_metrics.py baseline.json current.json
"""
from __future__ import annotations

import json
import sys


def _load(path: str) -> dict:
    with open(path) as f:
        return json.load(f)


def _check_regression(name: str, baseline: dict, current: dict) -> list[str]:
    issues = []
    if current["n"] == 0:
        issues.append(f"{name}: n=0 in current (no successful examples)")
        return issues
    if baseline.get("n", 0) == 0:
        return issues
    if current["mae_total"] - baseline["mae_total"] > 10:
        issues.append(
            f"{name} mae_total regression: {baseline['mae_total']:.1f} → {current['mae_total']:.1f}"
        )
    if current["mae_per_competency"] - baseline["mae_per_competency"] > 5:
        issues.append(
            f"{name} mae_per_comp regression: {baseline['mae_per_competency']:.1f} → {current['mae_per_competency']:.1f}"
        )
    if baseline["hit_rate_within_120"] - current["hit_rate_within_120"] > 0.05:
        issues.append(
            f"{name} hit_rate_120 regression: {baseline['hit_rate_within_120']:.2%} → {current['hit_rate_within_120']:.2%}"
        )
    return issues


def main() -> int:
    if len(sys.argv) != 3:
        print(f"Usage: {sys.argv[0]} baseline.json current.json", file=sys.stderr)
        return 1

    baseline = _load(sys.argv[1])
    current = _load(sys.argv[2])

    issues: list[str] = []

    issues += _check_regression("overall", baseline.get("overall", {}), current.get("overall", {}))

    for band in ("0-400", "401-700", "701-1000"):
        b_band = baseline.get("bands", {}).get(band, {})
        c_band = current.get("bands", {}).get(band, {})
        if b_band and c_band:
            issues += _check_regression(f"band[{band}]", b_band, c_band)

    baseline_inep = set(baseline.get("inep_failures", []))
    current_inep = set(current.get("inep_failures", []))
    new_inep_failures = current_inep - baseline_inep
    if new_inep_failures:
        issues.append(f"New INEP failures: {sorted(new_inep_failures)}")

    if issues:
        print("GOLDEN METRIC REGRESSION DETECTED:")
        for issue in issues:
            print(f"  ✗ {issue}")
        return 1

    print("Golden metrics: no regression vs baseline ✓")
    return 0


if __name__ == "__main__":
    sys.exit(main())
