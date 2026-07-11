# Investment Ledger Platform — Agent Guide

## Purpose

This repository is a learning project for implementing a double-entry ledger
database and API with PostgreSQL and Go Fiber. Treat every user request as a
learning opportunity, not merely a request to produce working code.

The user is also learning how to implement an API with correct layered
architecture. Prefer designs that make boundaries, dependencies, and tests
easy to understand.

## How to Collaborate

- Reply in Indonesian unless the user asks for another language.
- Explain the relevant concept, why a design is chosen, and its trade-offs
  before or alongside implementation.
- Connect explanations to the current codebase and schema where possible.
- Make incremental changes that the user can inspect and learn from. Do not
  introduce unnecessary frameworks, abstractions, or production complexity.
- When proposing alternatives, state when each is appropriate.
- Ask only for decisions that materially affect the ledger model or public API;
  otherwise make a clear, conservative assumption.

## Layered Architecture

Keep dependency direction toward the domain. The domain must not depend on
Fiber, PostgreSQL, or other delivery/infrastructure details.

- **Handler / delivery:** parse and validate HTTP input, call a use case,
  translate domain errors to HTTP responses, and serialize output. Keep
  business rules out of handlers.
- **Service / use case:** orchestrate a business operation, define transaction
  boundaries, enforce application-level rules, and depend on repository
  interfaces rather than PostgreSQL implementations.
- **Domain:** entities, value objects, domain errors, and repository
  interfaces. It expresses business/accounting rules and stays framework-free.
- **Repository / infrastructure:** implement persistence and SQL. It maps
  database errors and records to domain-level values, but does not decide HTTP
  responses or own business workflows.

When adding a feature, explain why each piece belongs in its chosen layer and
write tests at the most useful boundary: unit tests for domain/service rules,
and integration tests for PostgreSQL queries and transaction behaviour.

## Ledger Principles

Protect these invariants in schema, application code, and tests:

- A posted journal entry has at least two journal lines.
- For each journal entry, total debits equal total credits.
- Store monetary amounts as integer minor units (`BIGINT`), never floating
  point.
- Write business state and its journal entry in the same PostgreSQL
  transaction.
- Write APIs need idempotency so retries cannot create duplicate journal
  entries.
- Posted journal entries are immutable; correct mistakes with a reversing or
  compensating entry rather than editing historical lines.
- Derive balances from journal lines unless a separately maintained balance
  projection is explicitly designed and kept consistent.

## What to Explain for Changes

For changes involving migrations, the ledger, or write endpoints, cover the
following when relevant:

1. The accounting effect: which accounts are debited and credited.
2. The database invariant and how it is enforced.
3. The transaction boundary, including rollback behavior.
4. Concurrency and idempotency considerations.
5. The layer that owns each responsibility.
6. Tests that demonstrate both the successful and rejected cases.

## Current Technical Scope

- Database: PostgreSQL migrations managed with Goose.
- API: Go with Fiber v3.
- Primary domain: investment cash flows, orders, holdings, accounts, journal
  entries, and journal lines.

Preserve unrelated user changes in the working tree. Do not overwrite or
discard them.
