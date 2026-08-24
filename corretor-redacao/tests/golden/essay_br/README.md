# Essay-BR — golden dataset for ENEM essay correction

This directory holds the **Essay-BR extended corpus** (Marinho, Anchiêta & Moura, 2022),
converted into our golden-dataset on-disk format.

## Upstream

- Repository: https://github.com/rafaelanchieta/essay
- License: MIT (see top-level `NOTICE` for the full license text and citation).
- Pinned commit hash for reproducible ingest is set in `ingest.py`
  (`UPSTREAM_COMMIT_HASH`).

## Splits

- `test/` — **merge-gating** subset (Constitution v2.0.0 Article IX). The pytest harness at
  `tests/golden/test_golden_dataset.py` iterates this directory.
- `valid/` — for prompt iteration during development. **NOT a CI gate.**
- `train/` is intentionally not ingested to prevent prompt-engineering contamination.

## Layout

```
<split>/<slug>/
  essay.txt
  prompt_theme.txt        # title + "---" separator + contextualization
  motivational_texts.txt  # may be empty
  expected.json           # human reference scores: {c1, c2, c3, c4, c5, total}
```

## Running the ingest

The ingest script is reproducible and idempotent. Run once locally and commit the generated
files (the MIT license permits redistribution; CI never depends on network access).

```bash
uv run python tests/golden/essay_br/ingest.py
# add --force to overwrite an existing tree
```

A `MANIFEST.json` is written next to the splits with SHA-256s of every produced file plus the
upstream pins (commit hash, tarball SHA-256). CI re-verifies the manifest matches.

## Caveats

- Essays come from an online learning platform, not actual ENEM exams; human reference scores
  are expert proxies, not official INEP grades. Acceptable for MVP gating per Constitution IX.
- Distribution is right-skewed (upper-middle bands dominate). The harness reports per-band
  (0–400 / 401–700 / 701–1000) MVP-tier metrics independently to expose low-score regressions.
