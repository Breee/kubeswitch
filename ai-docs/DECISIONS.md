# AI Docs — Architecture Decisions & TODOs

This directory documents decisions made during AI-assisted development sessions and tracks open items.

## Decisions

### 001 — Release workflow with semver auto-bump

**Date:** 2026-04-24
**Status:** Implemented

The release workflow (`release.yaml`) supports both manual `workflow_dispatch` (with patch/minor/major selector) and tag-push triggers. It uses `gh release create --generate-notes` for auto-generated release notes categorized via `.github/release.yml`. Releases are only created when `main.go`, `go.mod`, or `go.sum` have changed since the last tag.

**Rationale:** Copilot cloud agent cannot reliably create Git tags or releases. A deterministic workflow is the right tool for semver bumps.

### 002 — Bash autocomplete with `ks` alias

**Date:** 2026-04-24
**Status:** Implemented

Shell completions live in `completions/kubeswitch.bash`. Users source the file to get tab-completion for contexts (arg 1) and namespaces (arg 2), plus a `ks` alias. Completions use `kubectl` under the hood.

### 003 — E2E testing with kind clusters

**Date:** 2026-04-24
**Status:** Implemented

`e2e_test.sh` spins up two kind clusters, creates test namespaces, and validates all quick-switch modes (namespace-only, context+namespace, context+dot, non-existent namespace). Clusters are cleaned up on exit via trap.

### 004 — Makefile as single entry point

**Date:** 2026-04-24
**Status:** Implemented

All build, test, and release tasks are wrapped in a `Makefile`. This gives contributors a single, discoverable interface (`make help`) instead of having to know individual commands.

## Developer Flow

```
1. Clone & setup
   git clone <repo> && cd kubeswitch

2. Validate changes
   make              # runs: vet → test → build (default target)

3. Run e2e tests (requires kind + kubectl)
   make e2e          # builds binary, spins up 2 kind clusters, runs tests, cleans up

4. Cross-compile release tarballs
   make dist         # produces linux/amd64, darwin/amd64, darwin/arm64 tarballs

5. Install shell completions
   make install-completions   # prints the source line to add to your shell profile

6. Create a release
   Go to Actions → Release → Run workflow → pick patch/minor/major
   (only triggers if main.go, go.mod, or go.sum changed since last tag)

7. Clean up
   make clean        # removes binary and tarballs
```

All available targets: `make help`

## TODOs

- [ ] Add zsh/fish completion scripts
- [ ] Add e2e tests to CI (requires kind-capable runner or a dedicated workflow)
- [ ] Consider supporting `KUBECONFIG` with comma-separated paths (Windows)
- [ ] Explore scheduled release workflow (e.g., weekly patch if changes exist)
