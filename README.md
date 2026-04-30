# kubeswitch

![CI](https://github.com/Breee/kubeswitch/workflows/CI/badge.svg)
![Tag](https://img.shields.io/github/v/tag/Breee/kubeswitch)
![Go Version](https://img.shields.io/github/go-mod/go-version/Breee/kubeswitch)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/Breee/kubeswitch)](https://pkg.go.dev/github.com/Breee/kubeswitch)
[![Go Report Card](https://goreportcard.com/badge/github.com/Breee/kubeswitch)](https://goreportcard.com/report/github.com/Breee/kubeswitch)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)

Switch your current kubernetes context and namespace graphically by selecting from a tree.
kubeswitch talks to the kubernetes API and does not depend on kubectl.

![Demo](https://raw.githubusercontent.com/Breee/kubeswitch/main/demo.gif)

**Note for Non-Admin Users:** If you are a cluster tenant without API-permission to list namespaces, kubeswitch won't work for you (as it can't retrieve available namespaces). Sorry, there's not much we can do about that.

## Install

Available for Linux and MacOS: [Latest Release](https://github.com/Breee/kubeswitch/releases/latest)

```bash
curl -sL https://github.com/Breee/kubeswitch/releases/latest/download/kubeswitch_$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m | sed 's/x86_64/amd64/').tar.gz | sudo tar xz -C /usr/local/bin kubeswitch
```

## Config

Read from the default location `~/.kube/config`. If not present, the location is read from environment variable `KUBECONFIG` (remember to `export`). That env variable can contain multiple locations separated by `:` from where configs are merged together.

## Run

| Run... | to... |
|-|-|
| `kubeswitch` | select context/namespace graphically |
| `kubeswitch <namespace>` | switch to namespace in current context quickly |
| `kubeswitch <context> <namespace>` | switch to context/namespace |
| `kubeswitch <context> .` | switch to `default` namespace of context |
| `kubeswitch version` | print version |
| `kubeswitch version -o json` | print version as JSON|

## TUI Controls

| Key | Action |
|-|-|
| `↑`/`↓` or `k`/`j` | Navigate |
| `←` or `h` | Collapse context / jump to parent context from namespace |
| `→` or `l` | Expand context |
| `Enter` or `Space` | Toggle expand/collapse context or select namespace |
| `/` | Start fuzzy search |
| `Esc` | Clear search filter or quit |
| `q` | Quit |

### Fuzzy Search

Press `/` to activate the search filter. It matches context and namespace names using:

1. **Substring** — `prod` matches `production`
2. **Subsequence** — `prd` matches `production` (letters in order)
3. **Edit distance** — `prodution` matches `production` (tolerates typos)

## Shell Completions

Tab-completion for contexts and namespaces, plus a `ks` alias. Supports bash, zsh, and fish.

```bash
# Bash — add to ~/.bashrc:
source <(kubeswitch completion bash)

# Zsh — add to ~/.zshrc:
source <(kubeswitch completion zsh)

# Fish — add to ~/.config/fish/config.fish:
kubeswitch completion fish | source
```

This gives you:
- `ks` as a short alias for `kubeswitch`
- Tab-completion for context names (first argument)
- Tab-completion for namespaces within a context (second argument)

## Performance

Namespace fetching for all contexts is performed **concurrently**, so startup time scales with the slowest cluster rather than the sum of all clusters.

### Debug / Profiling

Set `KUBESWITCH_DEBUG=1` to print timing information for each operation to stderr:

```bash
KUBESWITCH_DEBUG=1 kubeswitch
```

Example output:

```
[DEBUG] fetched namespaces for "prod-eu": 12 namespaces in 320ms
[DEBUG] fetched namespaces for "prod-us": 8 namespaces in 280ms
[DEBUG] fetched namespaces for "staging": 5 namespaces in 150ms
[DEBUG] total namespace fetch (parallel): 322ms
[DEBUG] total TUI setup: 323ms
```

## E2E Testing

The end-to-end test creates two [kind](https://kind.sigs.k8s.io/) clusters and validates context/namespace switching. Requires `kind`, `kubectl`, and `go`.

```bash
make e2e
```

## Development

All tasks are wrapped in a Makefile:

```bash
make          # vet + test + build
make test     # unit tests only
make e2e      # end-to-end tests with kind clusters
make dist     # cross-compile release tarballs
make clean    # remove build artifacts
make help     # show all targets
```
