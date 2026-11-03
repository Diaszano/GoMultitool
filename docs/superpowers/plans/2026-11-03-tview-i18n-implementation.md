# GoMultitool tview and Portuguese/English UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the current full-screen Bubble Tea UI with tview and localize all application-authored human-facing interface text in pt-BR and English.

**Architecture:** Resolve a small immutable presentation catalog once at startup and pass it to the full-screen TUI, line console, and human-readable CLI output. Keep application/domain behavior outside widgets; JSON output and the recovery image's default startup remain unchanged.

**Tech Stack:** Go 1.27.3, `github.com/rivo/tview` v0.42.0, tcell, Go standard library tests.

**Spec:** `docs/superpowers/specs/2026-11-03-tview-i18n-design.md`

## Global Constraints

- The product label is `GoMultitool`; preserve the `multitool` executable and command syntax.
- Select the first supported `pt-BR` or `en` value from `LC_ALL`, then `LC_MESSAGES`, then `LANG`; use `pt-BR` when none is supported.
- Show all planned TUI menu areas, but only Hardware and Exit are actionable; other entries show `Em breve` / `Coming soon` and do nothing.
- Keep the line-oriented `--console` frontend and the recovery image's opt-in Go runtime behavior.
- Preserve explicit confirmation and cancellation safety; do not add workflows behind future menu entries.
- Keep JSON output stable and do not translate machine-oriented errors, logs, or raw OS/library errors.
- Use tview v0.42.0; remove Bubble Tea/Lip Gloss only after confirming no remaining package uses them.
- Manually validate full-screen keyboard operation on RK3229 hardware before calling the TUI work complete.

## Review Focus

- Unsupported locale ahead of a supported locale: must skip it and select the first supported value; cover in locale precedence tests.
- `pt_BR.UTF-8`, `en_US.UTF-8`, empty variables, and unset variables: normalize/fallback correctly; cover in locale tests.
- Future menu item selected with Enter: must not invoke a workflow; cover in menu behavior tests.
- Empty inventory and inventory errors: remain distinguishable and safely rendered; cover in TUI tests.
- JSON success/error output under either locale: preserve current machine output; cover in CLI tests.

---

### Task 1: Add the two-language catalog and locale detection

**Files:**
- Modify: `internal/presentation/i18n/pt_br.go` (replace mutable globals with catalog and locale detection)
- Create: `internal/presentation/i18n/i18n_test.go`

**Interfaces:**
- Produce `type Locale string`, constants `LocalePtBR` and `LocaleEN`, `type Catalog struct`, `func DetectLocale(getenv func(string) string) Locale`, and `func For(locale Locale) Catalog`.
- Catalog fields cover product/menu labels and hints, navigation, hardware screen/empty/error text, console prompts, and human CLI usage/version labels. Do not add strings for workflows that are not implemented.

- [ ] **Step 1: Write failing table tests** for locale precedence, unsupported and empty values, locale normalization (`pt_BR.UTF-8`, `en_US.UTF-8`), fallback to pt-BR, and the two catalogs' required menu labels.
- [ ] **Step 2: Run** `go test ./internal/presentation/i18n`; confirm the new API/tests fail before implementation.
- [ ] **Step 3: Implement** `DetectLocale` using only `LC_ALL`, `LC_MESSAGES`, and `LANG`, and return catalog values without mutable package globals.
- [ ] **Step 4: Run** `go test ./internal/presentation/i18n`; expect PASS.

### Task 2: Replace the full-screen Bubble Tea UI with tview

**Files:**
- Modify: `go.mod`, `go.sum`
- Modify: `cmd/multitool/main.go` (construct tview application and resolve catalog)
- Modify: `internal/presentation/tui/app.go`
- Modify: `internal/presentation/tui/app_test.go`
- Modify: `internal/presentation/tui/screens/home.go` and `home_test.go`
- Delete if unused after migration: `internal/presentation/tui/screens/confirmation.go`, `confirmation_test.go`, `progress.go`, `progress_test.go`, `internal/presentation/tui/styles/theme.go`

**Interfaces:**
- Consume `i18n.Catalog` from Task 1.
- Use `func New(hardware func() ([]device.Device, error), catalog i18n.Catalog) *tview.Application` (or equivalent minimal constructor) as the TUI entry point.

- [ ] **Step 1: Add failing tests** for all ten menu entries in both locales, Hardware and Exit being the only actionable entries, future entries having no action, and hardware inventory success/empty/error render states.
- [ ] **Step 2: Run** `go test ./internal/presentation/tui/...`; confirm tests fail against the current Bubble Tea implementation.
- [ ] **Step 3: Add** `github.com/rivo/tview@v0.42.0` and implement the existing home and hardware screens using tview widgets; use only the currently supported actions.
- [ ] **Step 4: Update** `cmd/multitool` to start `tview.Application` for `tui`; preserve `q`, Ctrl+C, Escape/back behavior where applicable and surface startup errors to stderr.
- [ ] **Step 5: Remove** obsolete Bubble Tea/Lip Gloss screen code and dependencies only when `rg` confirms no remaining Go references; retain unrelated safety/domain code.
- [ ] **Step 6: Run** `go test ./internal/presentation/tui/... ./cmd/multitool`; expect PASS.

### Task 3: Localize console and human-readable CLI output

**Files:**
- Modify: `internal/presentation/console/app.go`
- Modify: `internal/presentation/console/app_test.go`
- Modify: `cmd/multitool/main.go`
- Modify: `cmd/multitool/main_test.go`

**Interfaces:**
- Pass the same `i18n.Catalog` into `console.Run` and `run` (or their smallest existing equivalents); JSON command paths must not read translated strings.

- [ ] **Step 1: Add failing tests** for `--console`, `version`, usage, and app-authored error context in pt-BR and English; assert existing JSON success and error encodings are unchanged.
- [ ] **Step 2: Run** `go test ./internal/presentation/console ./cmd/multitool`; confirm localization tests fail first.
- [ ] **Step 3: Replace** user-visible hardcoded console/CLI strings with catalog fields, keeping raw underlying errors as-is.
- [ ] **Step 4: Run** `go test ./internal/presentation/console ./cmd/multitool`; expect PASS.

### Task 4: Verify integration and hardware interaction

**Files:**
- Modify only any files exposed by integration failures from Tasks 1–3.

- [ ] **Step 1: Run** `go test ./...`; expect PASS.
- [ ] **Step 2: Run** `go vet ./...`; expect no diagnostics.
- [ ] **Step 3: Run** `make check` and `make build-armv7`; expect both targets to pass.
- [ ] **Step 4: On RK3229 with USB keyboard**, launch `multitool tui` under pt-BR and English locale; verify menu fit, selection, disabled future entries, Hardware view, Escape/back, and exit. Verify `multitool --console` still works.
- [ ] **Step 5: Record** hardware validation result; if unavailable, leave it explicitly pending rather than claim completion.

## Plan relationship

This focused plan supersedes Tasks 14–15 in `docs/superpowers/plans/2026-11-02-go-modernization.md` for the TUI implementation. It does not authorize implementation of the future Penpot workflows.
