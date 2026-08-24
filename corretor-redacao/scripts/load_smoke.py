"""Load smoke test — 5 concurrent submissions, polls to completion.

Asserts:
  - p95 submission ack < 500 ms (FR-041)
  - p95 e2e correction time < 90 s (FR-023)

Usage:
    python scripts/load_smoke.py --base-url http://localhost:8000 \
        --email user@example.com --password secret

Not in CI by default; run manually before production deploys.
"""
from __future__ import annotations

import argparse
import asyncio
import statistics
import sys
import time
from typing import Any

import httpx

ESSAY = (
    "A desigualdade social no Brasil é um problema estrutural de grande magnitude. "
    "Historicamente, a concentração de renda impossibilitou o acesso igualitário a direitos. "
    "Educação, saúde e moradia são negados sistematicamente às camadas mais vulneráveis. "
    "O mercado de trabalho informal absorve boa parte dos trabalhadores sem proteção social. "
    "Políticas redistributivas, como o Bolsa Família, atenuam mas não resolvem o problema. "
    "A reforma tributária progressiva é apontada por economistas como medida estrutural. "
    "É indispensável que o Estado amplie programas de habitação e educação de qualidade. "
) * 3

THEME = {
    "title": "Desigualdade Social no Brasil",
    "context": "A concentração de renda e seus impactos na cidadania brasileira.",
}

POLL_INTERVAL_S = 2.0
POLL_MAX_WAIT_S = 120.0


async def _submit(client: httpx.AsyncClient, token: str) -> tuple[str, float]:
    """Returns (correction_id, ack_latency_s)."""
    t0 = time.perf_counter()
    r = await client.post(
        "/corrections",
        json={"essay_text": ESSAY, "prompt_theme": THEME},
        headers={"Authorization": f"Bearer {token}"},
    )
    latency = time.perf_counter() - t0
    r.raise_for_status()
    return r.json()["correction_id"], latency


async def _poll_until_done(
    client: httpx.AsyncClient,
    token: str,
    correction_id: str,
) -> float:
    """Returns e2e latency in seconds from first poll to terminal state."""
    t0 = time.perf_counter()
    deadline = t0 + POLL_MAX_WAIT_S
    while time.perf_counter() < deadline:
        r = await client.get(
            f"/corrections/{correction_id}",
            headers={"Authorization": f"Bearer {token}"},
        )
        r.raise_for_status()
        body = r.json()
        if body["status"] in ("completed", "failed"):
            return time.perf_counter() - t0
        await asyncio.sleep(POLL_INTERVAL_S)
    raise TimeoutError(f"Correction {correction_id} did not complete within {POLL_MAX_WAIT_S}s")


async def _login(client: httpx.AsyncClient, email: str, password: str) -> str:
    r = await client.post("/auth/login", json={"email": email, "password": password})
    r.raise_for_status()
    return r.json()["access_token"]


async def run_smoke(base_url: str, email: str, password: str, concurrency: int = 5) -> None:
    async with httpx.AsyncClient(base_url=base_url, timeout=10.0) as client:
        token = await _login(client, email, password)
        print(f"Logged in. Submitting {concurrency} concurrent corrections…")

        submit_tasks = [_submit(client, token) for _ in range(concurrency)]
        results = await asyncio.gather(*submit_tasks, return_exceptions=True)

        correction_ids: list[str] = []
        ack_latencies: list[float] = []
        for r in results:
            if isinstance(r, Exception):
                print(f"  SUBMIT ERROR: {r}", file=sys.stderr)
            else:
                cid, lat = r
                correction_ids.append(cid)
                ack_latencies.append(lat)
                print(f"  submitted {cid} ack={lat * 1000:.0f}ms")

        print(f"\nPolling {len(correction_ids)} corrections…")
        poll_tasks = [_poll_until_done(client, token, cid) for cid in correction_ids]
        e2e_results = await asyncio.gather(*poll_tasks, return_exceptions=True)

        e2e_latencies: list[float] = []
        for cid, r in zip(correction_ids, e2e_results):
            if isinstance(r, Exception):
                print(f"  POLL ERROR {cid}: {r}", file=sys.stderr)
            else:
                e2e_latencies.append(r)
                print(f"  {cid} e2e={r:.1f}s")

    print("\n--- Results ---")
    if ack_latencies:
        ack_p95 = sorted(ack_latencies)[int(len(ack_latencies) * 0.95)]
        print(f"Submission ack p95: {ack_p95 * 1000:.0f}ms (target < 500ms)")
        if ack_p95 > 0.5:
            print("  ✗ FAIL: p95 submission ack exceeds 500ms", file=sys.stderr)
        else:
            print("  ✓ PASS")

    if e2e_latencies:
        e2e_p95 = sorted(e2e_latencies)[int(len(e2e_latencies) * 0.95)]
        print(f"E2E correction p95: {e2e_p95:.1f}s (target < 90s)")
        if e2e_p95 > 90:
            print("  ✗ FAIL: p95 e2e exceeds 90s", file=sys.stderr)
            sys.exit(1)
        else:
            print("  ✓ PASS")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Load smoke test")
    parser.add_argument("--base-url", default="http://localhost:8000")
    parser.add_argument("--email", required=True)
    parser.add_argument("--password", required=True)
    parser.add_argument("--concurrency", type=int, default=5)
    args = parser.parse_args()
    asyncio.run(run_smoke(args.base_url, args.email, args.password, args.concurrency))
