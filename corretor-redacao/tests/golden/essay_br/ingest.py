"""Reproducible, idempotent ingest of the Essay-BR extended corpus.

Pins (verified 2026-05-29):
- `UPSTREAM_COMMIT_HASH`: exact commit of github.com/rafaelanchieta/essay to fetch.
- `UPSTREAM_TARBALL_SHA256`: SHA-256 of the GitHub tarball at that commit.

Outputs:
- `<this_dir>/test/<slug>/{essay.txt, prompt_theme.txt, motivational_texts.txt, expected.json}`
- `<this_dir>/valid/<slug>/...`
- `<this_dir>/MANIFEST.json` listing every produced file's SHA-256 plus the two pins above.

The `train/` split is INTENTIONALLY not ingested (Research R17: contamination guard).

The upstream layout is CSV-based: `essay-br/splits/testing.csv`,
`essay-br/splits/development.csv`, `essay-br/splits/training.csv`, plus
`essay-br/prompts.csv` (prompt id → contextualization).

Operating note: the script refuses to overwrite a non-empty `test/` or `valid/` directory unless
`--force` is passed. After ingest, commit the generated files.
"""

from __future__ import annotations

import argparse
import ast
import csv
import hashlib
import io
import json
import logging
import sys
import tarfile
import urllib.request
from dataclasses import dataclass
from pathlib import Path

# ---------------------------------------------------------------------------
# PINS (update both atomically when bumping)
# ---------------------------------------------------------------------------
UPSTREAM_OWNER = "rafaelanchieta"
UPSTREAM_REPO = "essay"
UPSTREAM_COMMIT_HASH = "31e502ff3075e182cdda2560a767be65db677fa5"
UPSTREAM_TARBALL_SHA256 = "270a3b00fed6fb56435635f8dbea583fee84239b1f8f9e5bccdeb0440c9441df"


HERE = Path(__file__).parent

# Upstream split-name → our split-name
SPLIT_MAP = {"testing": "test", "development": "valid"}  # "training" intentionally omitted
TRAIN_SKIPPED_REASON = "Research R17: prevent prompt-engineering contamination"

logger = logging.getLogger("essay_br.ingest")


@dataclass(slots=True)
class CorpusEntry:
    split: str
    essay_text: str
    prompt_theme_title: str
    prompt_theme_context: str
    motivational_texts: str
    expected: dict[str, int]


def fetch_tarball(commit: str) -> bytes:
    url = f"https://codeload.github.com/{UPSTREAM_OWNER}/{UPSTREAM_REPO}/tar.gz/{commit}"
    logger.info("fetching %s", url)
    with urllib.request.urlopen(url, timeout=120) as resp:
        return resp.read()


def verify_tarball(data: bytes, expected_sha256: str) -> None:
    actual = hashlib.sha256(data).hexdigest()
    if actual != expected_sha256:
        raise RuntimeError(
            f"tarball SHA-256 mismatch.\n  expected: {expected_sha256}\n  actual:   {actual}\n"
            "Update UPSTREAM_TARBALL_SHA256 in tests/golden/essay_br/ingest.py if this is intentional."
        )


def parse_corpus(tarball: bytes) -> list[CorpusEntry]:
    out: list[CorpusEntry] = []
    prompts: dict[int, list[str]] = {}
    splits: dict[str, str] = {}  # upstream-split-name → csv text

    with tarfile.open(fileobj=io.BytesIO(tarball), mode="r:gz") as tf:
        for member in tf.getmembers():
            if not member.isfile():
                continue
            name = member.name
            if name.endswith("essay-br/prompts.csv"):
                fh = tf.extractfile(member)
                if fh is not None:
                    prompts = _parse_prompts_csv(fh.read().decode("utf-8"))
                continue
            for upstream_split in SPLIT_MAP:
                if name.endswith(f"essay-br/splits/{upstream_split}.csv"):
                    fh = tf.extractfile(member)
                    if fh is not None:
                        splits[upstream_split] = fh.read().decode("utf-8")
                    break

    if not prompts:
        logger.error("prompts.csv not found in upstream tarball")
        return []

    for upstream_split, our_split in SPLIT_MAP.items():
        csv_text = splits.get(upstream_split)
        if not csv_text:
            logger.warning("split %s missing in tarball; skipping", upstream_split)
            continue
        for entry in _parse_essays_csv(csv_text, prompts=prompts, split=our_split):
            out.append(entry)
    return out


def _parse_prompts_csv(text: str) -> dict[int, list[str]]:
    out: dict[int, list[str]] = {}
    reader = csv.DictReader(io.StringIO(text))
    for row in reader:
        pid_str = row.get("id") or ""
        desc_str = row.get("description") or ""
        if not pid_str or not desc_str:
            continue
        try:
            pid = int(pid_str)
        except ValueError:
            continue
        paragraphs = _safe_literal_list(desc_str)
        if paragraphs is None:
            continue
        out[pid] = paragraphs
    return out


def _parse_essays_csv(
    text: str,
    *,
    prompts: dict[int, list[str]],
    split: str,
) -> list[CorpusEntry]:
    out: list[CorpusEntry] = []
    reader = csv.DictReader(io.StringIO(text))
    for row in reader:
        prompt_id_str = row.get("prompt") or ""
        title = (row.get("title") or "").strip()
        essay_field = row.get("essay") or ""
        competence_field = row.get("competence") or ""
        score_field = row.get("score") or ""
        if not all((prompt_id_str, title, essay_field, competence_field, score_field)):
            continue
        try:
            prompt_id = int(prompt_id_str)
        except ValueError:
            continue
        if prompt_id not in prompts:
            continue
        essay_paragraphs = _safe_literal_list(essay_field) or []
        if not essay_paragraphs:
            continue
        essay_text = "\n\n".join(p.strip() for p in essay_paragraphs).strip()
        comp = _safe_literal_list(competence_field)
        if not comp or len(comp) != 5:
            continue
        try:
            comp_ints = [int(c) for c in comp]
        except (TypeError, ValueError):
            continue
        if any(c not in (0, 40, 80, 120, 160, 200) for c in comp_ints):
            continue
        try:
            total = int(score_field)
        except ValueError:
            total = sum(comp_ints)
        if total != sum(comp_ints):
            continue

        prompt_paragraphs = prompts[prompt_id]
        # First N-1 paragraphs (usually the "discussion + proposta") → context.
        # Last paragraph(s) often the motivational texts (textos motivadores).
        # Heuristic: keep all paragraphs as a single context block; motivational
        # stays empty (we lose the split but never lose information).
        context = "\n\n".join(p.strip() for p in prompt_paragraphs).strip()

        out.append(
            CorpusEntry(
                split=split,
                essay_text=essay_text,
                prompt_theme_title=title,
                prompt_theme_context=context,
                motivational_texts="",
                expected={
                    "c1": comp_ints[0],
                    "c2": comp_ints[1],
                    "c3": comp_ints[2],
                    "c4": comp_ints[3],
                    "c5": comp_ints[4],
                    "total": total,
                },
            )
        )
    return out


def _safe_literal_list(raw: str) -> list[str] | None:
    raw = raw.strip()
    if not raw or not raw.startswith("[") or not raw.endswith("]"):
        return None
    try:
        value = ast.literal_eval(raw)
    except (ValueError, SyntaxError):
        return None
    if not isinstance(value, list):
        return None
    return [str(v) for v in value]


def write_split(entries: list[CorpusEntry], *, root: Path, force: bool) -> dict[str, str]:
    if root.exists() and any(root.iterdir()) and not force:
        raise FileExistsError(f"{root} is non-empty; pass --force to overwrite.")
    root.mkdir(parents=True, exist_ok=True)
    manifest: dict[str, str] = {}
    for index, entry in enumerate(entries, start=1):
        slug = f"{index:04d}"
        d = root / slug
        d.mkdir(parents=True, exist_ok=True)
        _put(d / "essay.txt", entry.essay_text + "\n", manifest, root.parent)
        _put(
            d / "prompt_theme.txt",
            entry.prompt_theme_title + "\n---\n" + entry.prompt_theme_context + "\n",
            manifest,
            root.parent,
        )
        _put(d / "motivational_texts.txt", entry.motivational_texts + "\n", manifest, root.parent)
        _put(
            d / "expected.json",
            json.dumps(entry.expected, ensure_ascii=False, sort_keys=True) + "\n",
            manifest,
            root.parent,
        )
    return manifest


def _put(path: Path, body: str, manifest: dict[str, str], manifest_root: Path) -> None:
    data = body.encode("utf-8")
    path.write_bytes(data)
    rel = path.relative_to(manifest_root).as_posix()
    manifest[rel] = hashlib.sha256(data).hexdigest()


def write_manifest(manifest: dict[str, str], path: Path) -> None:
    payload = {
        "upstream": {
            "owner": UPSTREAM_OWNER,
            "repo": UPSTREAM_REPO,
            "commit": UPSTREAM_COMMIT_HASH,
            "tarball_sha256": UPSTREAM_TARBALL_SHA256,
        },
        "skipped_splits": {"training": TRAIN_SKIPPED_REASON},
        "files": dict(sorted(manifest.items())),
    }
    path.write_text(json.dumps(payload, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")


def main() -> int:
    logging.basicConfig(level=logging.INFO, format="%(message)s")
    parser = argparse.ArgumentParser(prog="essay_br.ingest")
    parser.add_argument("--force", action="store_true", help="Overwrite non-empty split dirs.")
    parser.add_argument(
        "--tarball",
        help="Local path to a pre-downloaded GitHub tarball (skips network).",
    )
    parser.add_argument(
        "--max-per-split",
        type=int,
        default=None,
        help="Cap entries per split (useful for walking-skeleton subset).",
    )
    args = parser.parse_args()

    if args.tarball:
        tarball = Path(args.tarball).read_bytes()
    else:
        tarball = fetch_tarball(UPSTREAM_COMMIT_HASH)
    verify_tarball(tarball, UPSTREAM_TARBALL_SHA256)

    entries = parse_corpus(tarball)
    if not entries:
        logger.error("No usable essays parsed from upstream tarball.")
        return 3

    by_split: dict[str, list[CorpusEntry]] = {"test": [], "valid": []}
    for e in entries:
        if e.split in by_split:
            by_split[e.split].append(e)

    if args.max_per_split is not None:
        for s in by_split:
            by_split[s] = _stratified_subset(by_split[s], k=args.max_per_split)

    full_manifest: dict[str, str] = {}
    for split, items in by_split.items():
        if not items:
            continue
        m = write_split(items, root=HERE / split, force=args.force)
        full_manifest.update(m)
        logger.info("wrote %d entries to %s/", len(items), split)
    write_manifest(full_manifest, HERE / "MANIFEST.json")
    logger.info("manifest written")
    return 0


def _stratified_subset(items: list[CorpusEntry], *, k: int) -> list[CorpusEntry]:
    """Take roughly k entries stratified by the 3 score bands."""
    if k >= len(items):
        return items
    low = [e for e in items if e.expected["total"] <= 400]
    mid = [e for e in items if 401 <= e.expected["total"] <= 700]
    high = [e for e in items if e.expected["total"] >= 701]
    per = max(1, k // 3)
    out: list[CorpusEntry] = []
    for bucket in (low, mid, high):
        out.extend(bucket[:per])
    return out[:k]


if __name__ == "__main__":
    sys.exit(main())
