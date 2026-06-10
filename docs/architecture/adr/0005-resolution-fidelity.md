# ADR-0005: Source resolution fidelity (npm, git, url)

- Status: Accepted
- Date: 2026-06-10

## Context

The lockfile is only meaningful if agentguard knows exactly which bytes run for
each declared source. "Resolution fidelity" — turning `npx -y pkg@latest` into a
pinned, integrity-anchored artifact — is the spec's #1 hard part.

## Decision

- **npm:** pin the exact version via `npm view … version`, reuse the npm-reported
  `dist.integrity` (sha512) rather than recomputing it, and fetch the package via
  `npm pack` + tarball extraction for content hashing/analysis. Unpinned specs
  (`latest`, ranges) are flagged.
- **git:** pin the requested ref to a commit SHA via `git ls-remote`. The SHA is
  the integrity anchor; mutable refs are flagged.
- **url:** pin the leaf TLS certificate's SPKI hash. Remote code cannot be
  hashed; this is the honest ceiling, and any cert rotation surfaces as drift.

External commands and TLS dialing sit behind the `run.Runner` and `CertFetcher`
interfaces so every resolver is unit-tested with fakes and no network.

## Consequences

- Drift detection now covers version, npm integrity, and remote cert changes,
  not just local content hashes.
- A resolution failure degrades to a finding; it never aborts a scan.
- Follow-ups: git tree hashing (shallow clone), npm temp-dir cleanup, and an
  offline mode that skips network resolution.
