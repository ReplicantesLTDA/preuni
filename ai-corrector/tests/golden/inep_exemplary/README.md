# INEP Exemplary — Secondary Sanity-Check Set

Per Constitution v2.0.0 Article IX and research R19, this directory should hold 10–20
INEP-released exemplary essays (notas 1000 públicas) with provenance metadata. Failing the
per-essay tolerance gate on this set blocks merge even if Essay-BR overall metrics pass.

**Current state (walking-skeleton MVP, T051)**: 3 placeholder essays. They are sourced from
the highest-quality `total = 1000` records inside the Essay-BR corpus itself (not actual INEP
released essays). They are honest placeholders, clearly marked, suitable for plumbing the
secondary-gate pathway. The full curated INEP set (10–20 essays) is task T113.

Each `expected.json` carries:
- `c1..c5`, `total`: human reference scores.
- `prompt_year` (when known): the ENEM year the prompt was used.
- `source_url` (when known): public URL where the essay was released.

For the placeholder set:
- `prompt_year`: null (Essay-BR essays are platform-graded, not real ENEM exams).
- `source_url`: upstream Essay-BR commit URL.
- `note`: "Essay-BR nota-1000 stand-in pending the curated INEP set."
