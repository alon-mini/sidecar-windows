# Sync Into Main Safely: Merge Plan

## Objective
Merge the `sunc` sync implementation into `main` so users receive database migrations early, while keeping sync UX and background network behavior disabled by default.

## Constraints
- We want migration coverage in real user environments as soon as practical.
- We do not want end users to see or rely on sync/auth CLI yet.
- We need a fast rollback path if migration issues appear.

## High-Level Strategy
1. Merge sync code to mainline.
2. Keep schema migrations active.
3. Gate all sync-facing behavior with feature flags that default to `false`.
4. Enable flags only for internal testers and canary cohorts.

## Feature Flag Framework
This branch adds a feature framework with:
- Registry of known features in `internal/features/features.go`
- Resolution order:
1. Environment overrides
2. Local project config (`.todos/config.json` -> `feature_flags`)
3. Code default
- Emergency kill switch: `TD_DISABLE_EXPERIMENTAL=1`

Supported env overrides:
- `TD_FEATURE_<FEATURE_NAME>=true|false`
- `TD_ENABLE_FEATURE=flag1,flag2`
- `TD_DISABLE_FEATURE=flag1,flag2`

## Sync Feature Flags
- `sync_cli`: gates user-facing sync/auth/config/doctor commands
- `sync_autosync`: gates startup/post-mutation/monitor autosync behavior
- `sync_monitor_prompt`: gates monitor sync setup prompt UX

All three default to `false`.

## Gate Map
| Feature | Surface | Why It Must Be Gated |
|---|---|---|
| `sync_cli` | `cmd/auth.go` | Exposes login/logout/status |
| `sync_cli` | `cmd/sync.go` | Exposes push/pull/status sync command |
| `sync_cli` | `cmd/sync_conflicts.go` | Exposes conflict inspection |
| `sync_cli` | `cmd/sync_tail.go` | Exposes sync activity tailing |
| `sync_cli` | `cmd/sync_init.go` | Exposes guided setup wizard |
| `sync_cli` | `cmd/project.go` | Exposes sync-project management |
| `sync_cli` | `cmd/config.go` | Exposes sync config toggles |
| `sync_cli` | `cmd/doctor.go` | Exposes sync diagnostics |
| `sync_autosync` | `cmd/root.go` pre-run hook | Prevents startup sync side effects |
| `sync_autosync` | `cmd/root.go` post-run hook | Prevents post-mutation background sync |
| `sync_autosync` | `cmd/monitor.go` | Prevents periodic monitor sync loop |
| `sync_monitor_prompt` | `pkg/monitor/commands.go` | Prevents first-run sync prompt trigger |
| `sync_monitor_prompt` | `pkg/monitor/sync_prompt.go` | Prevents sync prompt modal flow |

The canonical map is also encoded in `internal/features/sync_gate_map.go`.

## Merge Sequencing
### Phase 1: Infrastructure
1. Merge this branch first.
2. Confirm flags resolve correctly from env and local config.
3. Confirm default behavior is unchanged for current users.

### Phase 2: Sync Merge
1. Merge `sunc` into a staging integration branch from `main`.
2. Replace direct sync command registration with `sync_cli`-gated registration.
3. Register autosync hooks through the gated hook interface in `cmd/feature_gate.go`.
4. Gate monitor prompt paths behind `sync_monitor_prompt`.

### Phase 3: Migration Canary
1. Release with all sync flags still off.
2. Monitor migration success/failure rates and support reports.
3. Keep a rollback release ready.

### Phase 4: Internal Enablement
1. Enable flags for internal testers via env or `td feature set`.
2. Run sync regression suite and cross-version upgrade checks.
3. Fix issues before broader exposure.

### Phase 5: Gradual Rollout
1. Enable `sync_cli` for a limited cohort.
2. Later enable `sync_monitor_prompt`.
3. Enable `sync_autosync` last.

## Rollback Plan
If regressions appear:
1. Ship a patch release with `TD_DISABLE_EXPERIMENTAL=1` in launcher/runtime environments.
2. Keep flags default-off in code.
3. Investigate migration failures using affected DB samples.
4. Re-enable per feature only after targeted fixes.

## Validation Checklist
- `td feature list` shows expected defaults and sources.
- Feature env overrides beat config values.
- With all flags off, no sync commands are exposed after sync merge.
- With all flags off, no autosync network behavior runs.
- DB opens and migrations run on upgraded clients.
