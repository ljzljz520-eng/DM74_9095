# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
?   	theatre39/cmd/theatre39	[no test files]
ok  	theatre39/internal/archive	0.019s
ok  	theatre39/internal/domain	0.002s
ok  	theatre39/internal/intake	0.013s
ok  	theatre39/internal/notify	0.009s
ok  	theatre39/internal/query	0.012s
ok  	theatre39/internal/review	0.023s
ok  	theatre39/internal/store	0.014s
--- FAIL: TestWorkflow39BusinessInvariant (0.01s)
    workflow_test.go:68: mixed outcomes must publish partial notification, got success
FAIL
FAIL	theatre39/internal/workflow39	0.045s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/theatre39): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/theatre39): exit `0`
