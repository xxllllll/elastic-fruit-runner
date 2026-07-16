# API Boundary Guidelines

## Scenario: Proto3 Default Fields Omitted from Dashboard JSON

### 1. Scope / Trigger

Apply this contract whenever Dashboard code consumes Connect RPC JSON.
Proto3 JSON omits scalar fields whose values equal their defaults, including
empty strings. TypeScript declarations must describe the received JSON rather
than the generated Go struct.

### 2. Signatures

Raw RPC fields that can be omitted must be optional:

```typescript
interface JobRecordsResponse {
  jobRecords: Array<{
    runnerName?: string
    runnerSetName?: string
  }>
}
```

The normalized application model remains stable:

```typescript
interface JobRecord {
  runnerName: string
  runnerSetName: string
}
```

### 3. Contracts

- Backend storage may contain `runner_name = ''` and
  `runner_set_name = ''` for completed-only jobs recorded after a daemon
  restart.
- Connect/Proto3 JSON may omit `runnerName` and `runnerSetName` for those rows.
- `dashboard/src/api/fetchers.ts` owns conversion from optional wire fields
  to required application fields.
- Components consume the normalized model and display `common.unknown` when
  an empty name has no display identity.

### 4. Validation & Error Matrix

- `runnerName` present and non-empty: render `shortName(runnerName)`.
- `runnerName` absent or empty: render localized `common.unknown`; do not call
  `shortName`.
- `runnerSetName` absent or empty: normalize to `''`; preserve the
  completed-only record.
- Static resources or RPC fail: use the existing Dashboard load-error state.

### 5. Good / Base / Bad Cases

- Good: a normal job renders its abbreviated Runner name and original
  status/duration.
- Base: a completed-only job renders `UNKNOWN` or `未知` and remains visible
  in history.
- Bad: a missing wire field reaches `name.split('-')` and aborts React rendering.

### 6. Tests Required

- Run `pnpm run lint` and `pnpm run build` in `dashboard/`.
- Run Playwright against a response containing a job without `runnerName` and
  `runnerSetName`.
- Assert that the page has accessibility content, the unknown label is
  visible, normal Runner names remain visible, and the browser console has no
  errors.

### 7. Wrong vs Correct

#### Wrong

```typescript
interface JobRecordsResponse {
  runnerName: string
}

shortName(job.runnerName)
```

#### Correct

```typescript
interface JobRecordsResponse {
  runnerName?: string
}

runnerName: response.runnerName ?? ''

const label = job.runnerName ? shortName(job.runnerName) : t('common.unknown')
```
