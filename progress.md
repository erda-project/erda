# Progress

## 2026-10-26 - Task: Fix streaming panic when dropping future spot metrics
### What was done
- Corrected the dropped-metric counter call to match its single `receiver` label contract, preventing future spot metrics from crashing the streaming process.
- Added a regression test proving future metrics are dropped and counted without reaching the downstream consumer.
### Testing
- Reproduced the original `inconsistent label cardinality: expected 1 label values but got 2` panic before the fix.
- `GOCACHE=/private/tmp/codex-go-build-cache go test ./internal/tools/monitor/oap/collector/plugins/receivers/kafka -count=1` passed.
- Built `streaming:codex-fix-future-metric-panic-20261026-amd64` successfully from commit `3209da3331ebee08e8bd6c23121ba9c0d6a08003`.
- Verified the image is `linux/amd64` with image ID `sha256:2025cd3dad276516552edbebf60c40220d0f83c4d91f17b94e16cd6f21ebcb33`.
### Notes
- `internal/tools/monitor/oap/collector/plugins/receivers/kafka/provider.go` - corrected the Prometheus label values used when dropping future metrics.
- `internal/tools/monitor/oap/collector/plugins/receivers/kafka/provider_test.go` - covers the future-metric drop path and counter update.
- `progress.md` - records this fix and verification.
- Rollback: restore the two-label call in `provider.go`, remove `provider_test.go`, and remove this progress entry.
