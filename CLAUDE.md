# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

All project rules and conventions live in [rules/](rules/):

- [rules/conventions.md](rules/conventions.md) — hard rules (architecture, imports, errors, transactions, naming, etc.).
- [rules/architecture.md](rules/architecture.md) — design rationale, heuristics, and tradeoffs behind those rules.

When adding a new rule or convention, route it by intent:

- "You must do X" → `rules/conventions.md`
- "We chose X because Y" → `rules/architecture.md`
- Both → put the rule in `conventions.md` with a `→ see architecture.md#section` link.

@rules/conventions.md
@rules/architecture.md
