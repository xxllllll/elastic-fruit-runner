# Frontend Development Guidelines

## Guidelines Index

- [API Boundary Guidelines](./api-boundary-guidelines.md): normalize
  Connect/Proto3 JSON payloads before rendering. Status: active.

## Pre-Development Checklist

- Read [API Boundary Guidelines](./api-boundary-guidelines.md) when changing
  RPC response types, fetchers, or components that render API data.

## Quality Check

- Verify raw RPC response types reflect fields that Proto3 JSON may omit.
- Verify normalization happens in `dashboard/src/api/fetchers.ts`, not in
  every consumer.
- Exercise missing/default-valued fields in a browser-level regression.
