# Escape fork of terraform-provider-kubernetes

This repository is Escape's fork of [hashicorp/terraform-provider-kubernetes](https://github.com/hashicorp/terraform-provider-kubernetes), published to the Terraform Registry as `escape-technologies/kubernetes`.

## Why the fork exists

Upstream does not yet support Kubernetes image volumes (`spec.volumes[].image`, GA since Kubernetes 1.33) — tracked in hashicorp/terraform-provider-kubernetes#2850. We need this field in production, so we maintain a patch branch that carries the fix until it lands upstream (PR hashicorp/terraform-provider-kubernetes#2937).

## Patch branch: `escape/release`

`escape/release` is a long-lived maintenance branch, rebased on upstream releases. It is **not** a version tag, so `v*` tags remain free for releases.

Current state:

- Based on upstream `v3.4.0`.
- Carries cherry-picked upstream PR #2937 (commit `930378b1cdbbcffe55e961ca2a088f4647a25228` from `dennismdejong/terraform-provider-kubernetes`, branch `fix/2850-image-volume-source`): adds `volume { image { reference, pull_policy } }` to the common pod volume sources, which flows through pod, deployment, statefulset, daemonset, job, cronjob, and replication controller (v1) resources.
- Adds a schema guard test (`kubernetes/schema_image_volume_test.go`) so a future rebase cannot silently drop the field.

## Version scheme

Releases are tagged `v3.4.0-escape.N` (e.g. `v3.4.0-escape.1`) — the upstream base version plus an `escape.N` prerelease suffix, exact-pinned to the upstream tag the branch is rebased on. Never tag a plain `v3.4.0` here: that name belongs to upstream, and colliding tags make the rebase history ambiguous. Each new upstream base gets a fresh prefix (`v3.5.0-escape.1`, ...) with `N` restarting at 1.

To pin the provider:

```hcl
terraform {
  required_providers {
    kubernetes = {
      source  = "escape-technologies/kubernetes"
      version = "3.4.0-escape.1"
    }
  }
}
```

## Release process

Releases are cut by pushing a `v*` tag; `.github/workflows/release.yml` runs goreleaser (configured in `.goreleaser.yml`, modeled on hashicorp/terraform-provider-scaffolding-framework) and publishes signed archives named `terraform-provider-kubernetes_*` to the GitHub release. Signing requires the `GPG_PRIVATE_KEY` and `PASSPHRASE` repository secrets (owned by Quentin) — do not modify them.

The Terraform Registry manifest (`terraform-registry-manifest.json`) declares protocol version `6.0`: the provider binary is served by `main.go` through `tf6server` with a `tf6muxserver` mux (the SDKv2 provider is upgraded via `tf5to6server`), so it speaks protocol 6 only. Upstream's `5.0` manifest predates the protocol-6 migration and is not correct for this binary.

## Rebasing on upstream

```bash
git remote add upstream https://github.com/hashicorp/terraform-provider-kubernetes
git fetch upstream --tags
git checkout escape/release
git rebase <new-upstream-tag>        # e.g. v3.5.0
go build ./... && go test ./kubernetes/ -run 'TestImageVolume|TestFlatten|TestExpand'
```

Resolve conflicts, rerun the tests (the schema guard test should catch a dropped patch), and retag `v<new-base>-escape.1`.

## Exit plan

When upstream PR #2937 ships in a released upstream version, the fork is retired:

1. Verify the image volume support matches (schema-wise) the upstream release.
2. Migrate every configuration: `terraform state replace-provider hashicorp/kubernetes escape-technologies/kubernetes` inverted to `terraform state replace-provider escape-technologies/kubernetes hashicorp/kubernetes`.
3. Cut one final release announcing the switch, then archive this repository and stop publishing.
