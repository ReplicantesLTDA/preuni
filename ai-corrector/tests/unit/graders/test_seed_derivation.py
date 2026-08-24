"""T042: seed = int.from_bytes(sha256(correction_id.bytes)[:8], 'big')."""

from __future__ import annotations

import uuid

from src.corrector.graders.single_grader import derive_seed


def test_seed_deterministic_for_same_id() -> None:
    cid = uuid.uuid4()
    assert derive_seed(cid) == derive_seed(cid)


def test_seed_differs_across_ids() -> None:
    a = derive_seed(uuid.uuid4())
    b = derive_seed(uuid.uuid4())
    assert a != b


def test_seed_fits_postgres_bigint() -> None:
    """grader_passes.seed is a Postgres BIGINT (signed int64). A previous
    unsigned interpretation overflowed that column for ~half of all
    correction ids -- caught by tests/integration/workers/test_process_one.py
    inserting a real grader_passes row end to end."""
    s = derive_seed(uuid.uuid4())
    assert isinstance(s, int)
    assert -(2**63) <= s < 2**63


def test_seed_known_value() -> None:
    # Pin behavior with a known UUID.
    cid = uuid.UUID("00000000-0000-0000-0000-000000000000")
    # sha256(b"\x00" * 16)[:8] interpreted big-endian, signed (BIGINT-safe).
    import hashlib

    expected = int.from_bytes(hashlib.sha256(cid.bytes).digest()[:8], "big", signed=True)
    assert derive_seed(cid) == expected
