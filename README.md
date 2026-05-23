# fluid-pub/probe-gitlab

Fluid probe for GitLab: collects **users**, **groups**, **projects**, cloned **repositories**, and **code_files** for RAG; pushes entities to the control plane over HTTP (`/probes`).

## Repository layout

| Path | Role |
|------|------|
| `core/` | Git submodule → [`fluid-pub/probe-core`](https://github.com/fluid-pub/probe-core) |
| `cmd/` | Entrypoint and `cmd/version.go` (semver for releases) |
| `internal/` | GitLab API client, entities, repository sync |
| `config/probe.example.yml` | Configuration template |
| `config/schema.yml` | Entity schema (shipped in the Docker image; pushed to the control plane on connect) |
| `.github/workflows/` | CI and release via [`fluid-pub/actions`](https://github.com/fluid-pub/actions) |

## Local development

One-time per clone, enable the same **`gofmt`** check as CI:

```bash
./scripts/install-git-hooks.sh
```

```bash
git submodule update --init --recursive
cp config/probe.example.yml config/probe.yml
cp env.secrets.example env.secrets
# Set GITLAB_TOKEN and control plane values in env.secrets (never commit that file).
source env.secrets
make dev
```

Credentials and GitLab URLs come from **your** `env.secrets` and local `config/probe.yml` only. Runtime snapshots and git clones use `state/` and `data/repositories/` (gitignored).

## Changelog

Release notes: [CHANGELOG.md](CHANGELOG.md).

## Releases

Push a semver tag **without** `v` (e.g. `0.1.0`) matching `var Version` in `cmd/version.go`. The release workflow publishes:

- `ghcr.io/fluid-pub/probe-gitlab:<tag>`
- GitHub Release asset `gitlab-probe-linux-amd64` and `SHA256SUMS.txt`

## Control plane

Enroll as a probe with `agent_type: gitlab`. Operational tuning (`data.entities`, `fields.*.rag`, `data.repositories.repos`, intervals) belongs in **runtime_config** on the probe record. The schema contract is **`config/schema.yml`** (image semver); Kubernetes/GitOps should mount only bootstrap YAML (see **`fluid-workload`** `config.schemaInImage`).
