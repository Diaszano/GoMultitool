# GoMultitool TUI and Language Support Design

**Status:** Approved 2026-11-03
**Date:** 2026-11-03

## Intent

Replace GoMultitool's current Bubble Tea and Lip Gloss full-screen interface with `github.com/rivo/tview`, and make its human-facing messages available in Brazilian Portuguese and English. The recovery application should follow the device's configured language when it is Portuguese or English, while remaining usable when locale variables are absent or unsupported.

## Current state

- `cmd/multitool` starts a Bubble Tea TUI for the `tui` command and keeps a separate line-oriented `--console` fallback.
- `internal/presentation/i18n` contains a handful of mutable Brazilian Portuguese string globals. Other UI strings remain embedded in TUI screens, console output, and CLI usage/version output.
- JSON commands are machine-readable interfaces and must remain stable for scripts.
- The Go recovery runtime remains opt-in; this work does not change the image's default legacy startup path.

## Goals

- Use `tview` for the existing full-screen TUI screens and preserve current navigation and confirmation safety behavior.
- Show every planned TUI area in the main menu, but keep only Hardware and Exit actionable in this stage; label the other entries `Em breve` / `Coming soon` and give them no action.
- Centralize human-facing strings for TUI, line console, and human-readable CLI messages in a small Go catalog with `pt-BR` and `en` translations.
- Select the first supported `pt-BR` or `en` value from `LC_ALL`, then `LC_MESSAGES`, then `LANG`; use `pt-BR` when none is supported.
- Keep UI labels and layout behavior separate from application/domain logic.
- Keep the line-oriented console available for terminals where a full-screen TUI cannot run.

## Non-goals

- Add new storage, backup, restore, install, or diagnostics workflows.
- Make future menu entries launch placeholder or fake operational screens.
- Change which frontend the recovery image launches by default.
- Add a runtime language picker, persistent language setting, or `--lang` flag.
- Translate JSON keys, schemas, machine-oriented error values, logs, or raw operating-system/library error text.
- Add an external translation framework for two initial languages.

## Approaches considered

1. **Replace Bubble Tea with `tview` and add a shared language catalog — selected.** This matches the requested library and gives the planned TUI room for lists, tables, forms, and dialogs. It requires rewriting the existing TUI event/rendering code and revalidating keyboard behavior on the RK3229.
2. **Keep Bubble Tea and localize the existing interfaces.** This would be a smaller and lower-risk change, but it would defer the requested `tview` adoption and its widgets.

The user selected the first approach. `tview` is built on `tcell`; the console fallback remains available independently.

## Penpot reference

The GoMultitool TUI mockups are in the connected Penpot file `TUI` (file ID `fdcf7e67-1004-81a9-8008-e33c58e15456`), page `GoMultitool — TUI`. The page contains 16 boards covering the main menu in both languages, hardware, storage, backup, restore, image installation and configuration, diagnostics, system settings, terminal, destructive confirmation in both languages, progress, completion, and power actions. The user approved the visual direction and screen set on 2026-11-03.

For the first implementation stage, the main menu keeps all planned areas visible but marks everything other than Hardware and Exit as unavailable (`Em breve` / `Coming soon`). The detailed workflow boards are future design references and do not imply working actions in this stage.

## Language behavior

The presentation layer resolves a language once at startup using the locale environment variables in this order: `LC_ALL`, `LC_MESSAGES`, `LANG`. Values identifying Portuguese select `pt-BR`; values identifying English select `en`; unsupported values are skipped, and `pt-BR` is the final fallback. No language setting is persisted.

All strings authored by the application for human-facing TUI, console, and text CLI output use the selected catalog. JSON command output remains unchanged so scripts do not receive locale-dependent schemas or messages. Raw errors from Go, the kernel, and external utilities may remain in their original language, with surrounding application context localized where practical.

## Constraints and safety

- Keep destructive confirmations fail-closed: only explicit confirmation approves an operation, regardless of language.
- Preserve existing key behavior: arrows and Enter for navigation, Escape for back/cancel, and `q` or Ctrl+C for exit/cancel where currently supported.
- Do not move storage access or business rules into `tview` widgets.
- Retain the text console fallback and keep the Go recovery runtime opt-in until its separate hardware release gate passes.
- Pin the chosen `tview` module version through `go.mod`/`go.sum`; remove Bubble Tea/Lip Gloss dependencies only if no remaining package uses them.

## Acceptance

- Existing TUI flows render and navigate with `tview` without changing their behavior or safety properties.
- Human-facing TUI, console, version/help, and application-owned error strings render in both `pt-BR` and `en`.
- Locale precedence, unsupported locales, and missing locale variables are covered by tests.
- JSON CLI output remains byte/schema compatible with current behavior.
- `go test ./...`, `go vet ./...`, `make check`, and `make build-armv7` pass.
- The full-screen TUI is manually checked on the RK3229 with its USB keyboard; the line console fallback remains usable.
- The Penpot screen set is reviewed and approved before the implementation plan is created.

## Open implementation details

Use the smallest typed catalog structure that avoids mutable package globals. The builder may choose the exact `tview` primitives and test seam while preserving the current screens and contracts above.
