---
description: Check the version and tooling couplings that nothing in CI enforces
allowed-tools: Bash(./scripts/check-drift.sh)
disable-model-invocation: true
---

Run `./scripts/check-drift.sh` and report the result.

If anything has drifted, explain which two files disagree and which one is wrong -
not just that they differ. The version triple is owned by `.github/workflows/version.yml`
and should never be edited by hand; the other pairs are edited by hand and have only a
comment holding them together.
