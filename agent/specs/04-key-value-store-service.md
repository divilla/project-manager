# internal.variable.KeyValueStore Service

## Status

- Service: `internal/variable.KeyValueStore`
- Package: `internal/variable`
- Package name: `variable`
- Status: implementation specification

## Base specification

### Purpose

`KeyValueStore` is the run-wide, in-memory store used to share variables between
step runners. It stores string keys and string values, permits each key to be
written exactly once, and supports concurrent readers and writers.

The store preserves values exactly as supplied. Callers serialize step-level
YAML values or extract response values before calling `Set`; consumers receive
the same string from `Get` without parsing or normalization.

One suite run must receive one shared store instance. A later suite run must
receive a new instance so variables never leak between runs.

### Responsibilities

`KeyValueStore` owns:

- Holding run-wide string-to-string variable values in memory.
- Returning the exact stored value for an existing key.
- Distinguishing a missing key from a present key whose value is empty or is
  the JSON literal `null`, `false`, or `0`.
- Enforcing write-once keys.
- Atomically rejecting duplicate writes without changing the original value.
- Synchronizing all map access so `Get` and `Set` are safe for concurrent use.

### Non-responsibilities

`KeyValueStore` does not:

- Serialize YAML values to JSON.
- Parse, validate, compact, quote, or otherwise transform stored values.
- Substitute variables in request bodies or expected responses.
- Validate variable-key syntax. Definition validation and variable-reference
  parsing own the `[A-Za-z_][A-Za-z0-9_:]*` product rule.
- Attach file, step, or YAML source locations to errors. Callers may wrap store
  errors with assignment or lookup context.
- Persist values to disk or share them between suite runs.
- Delete, overwrite, enumerate, or reset keys.
- Perform its own lifetime management or create a global singleton.

### Public contract

```go
var ErrKeyNotFound = errors.New("variable key not found")
var ErrKeyAlreadyExists = errors.New("variable key already exists")

type KeyValueStore struct {
    mu     sync.RWMutex
    values map[string]string
}

func NewKeyValueStore() *KeyValueStore

func (s *KeyValueStore) Get(key string) (string, error)

func (s *KeyValueStore) Set(key, val string) error
```

`Get` and `Set` are the service's only methods. The constructor creates its
state but is not a method on the service.

### State and lifetime

The store retains only:

```text
map[string]string
```

and the mutex protecting it. `NewKeyValueStore` allocates an empty map. The
orchestrator constructs one store for each run and shares its pointer with all
step runners in that run.

The store has no package-global mutable state. Constructing a second store
produces independent state even while the first store remains in use.

`KeyValueStore` contains a mutex and therefore must not be copied after first
use. It is passed and shared as `*KeyValueStore`.

### Concurrency model

Every access to `values`, including existence checks, occurs while holding the
store's mutex:

- `Get` holds `mu.RLock` while checking the key and reading its value.
- `Set` holds `mu.Lock` across both the duplicate-key check and insertion.

The duplicate check and insertion form one atomic critical section. If multiple
goroutines concurrently call `Set` for the same previously absent key, exactly
one call succeeds. Every other call returns an error, and the value written by
the successful call remains stored.

A `Get` concurrent with the first successful `Set` may observe either the
not-yet-written or written state according to which operation acquires the
mutex first. After `Set` returns successfully, subsequent `Get` calls must
observe the stored value.

No method returns an alias to mutable internal state. Since keys and values are
strings, returned values are safe to use after the read lock is released.

### Error contract

Errors must retain stable identity for `errors.Is` checks:

- A missing key returns an error matching `ErrKeyNotFound`.
- A duplicate write returns an error matching `ErrKeyAlreadyExists`.

Each error must identify the relevant key in its text. The store does not know
the assignment or lookup source, so the calling service is responsible for
adding source-aware context and applying the product's fatal runtime
configuration-error policy.

Neither error mutates the store.

## `NewKeyValueStore`

```go
func NewKeyValueStore() *KeyValueStore
```

### Behavior

`NewKeyValueStore` returns a non-nil pointer to an empty, ready-to-use store. It
performs no filesystem access, starts no goroutines, and shares no state with
any previously constructed store.

### Acceptance criteria

#### AC-NewKeyValueStore-1: Create an empty store

When `NewKeyValueStore` is called, then it returns a non-nil store and `Get` for
any key reports `ErrKeyNotFound`.

#### AC-NewKeyValueStore-2: Keep stores independent

Given two separately constructed stores, when a key is set in one, then the key
remains absent from the other.

## `Get`

```go
func (s *KeyValueStore) Get(key string) (string, error)
```

### Input

- `key`: the exact string key to look up.

Key syntax has already been validated by the definition or substitution layer.
`Get` performs an exact, case-sensitive lookup.

### Output

- The exact stored string and `nil` when `key` exists.
- The empty string and an error matching `ErrKeyNotFound` when `key` does not
  exist.

The error return, rather than the returned string, distinguishes a missing key
from a present key storing an empty string.

### Mutation

`Get` does not insert, remove, or modify any key or value.

### Acceptance criteria

#### AC-Get-1: Return an existing value unchanged

Given a key storing any string, when `Get` is called with that key, then it
returns the exact string supplied to `Set` and a nil error.

#### AC-Get-2: Report a missing key

Given an absent key, when `Get` is called, then it returns the empty string and
an error matching `ErrKeyNotFound` whose text identifies the key.

#### AC-Get-3: Distinguish false-like and empty values from absence

Given present keys storing `null`, `false`, `0`, or the empty string, when each
key is read, then `Get` returns its stored string and a nil error.

#### AC-Get-4: Use exact key matching

Given a stored key, when `Get` is called with a differently cased or otherwise
different key, then the lookup reports `ErrKeyNotFound`.

## `Set`

```go
func (s *KeyValueStore) Set(key, val string) error
```

### Input

- `key`: the exact string key to create.
- `val`: the exact string value to store.

`Set` does not interpret `val`. In normal product flow, callers supply compact
JSON for literal step variables or the literal text emitted by response
capture.

### Output

- `nil` when the previously absent key is stored successfully.
- An error matching `ErrKeyAlreadyExists` when the key is already present.

### Mutation

A successful call adds exactly one key-value pair. A failed duplicate call
changes nothing: it does not replace, remove, or temporarily expose the new
value.

### Acceptance criteria

#### AC-Set-1: Store a new key

Given an absent key, when `Set` is called, then it returns nil and a subsequent
`Get` returns the supplied value unchanged.

#### AC-Set-2: Preserve JSON literal text

Given values such as `1`, `"2026-01-01"`, `null`, `[1,2]`, or an object string,
when each is stored, then `Get` returns the byte-for-byte identical string.

#### AC-Set-3: Reject a duplicate key

Given a key with an existing value, when `Set` is called again for that key,
then it returns an error matching `ErrKeyAlreadyExists`, identifies the key in
the error text, and preserves the original value.

#### AC-Set-4: Treat an empty stored value as present

Given a key storing the empty string, when `Set` is called again for that key,
then it reports `ErrKeyAlreadyExists` and does not treat the key as absent.

## Concurrent behavior

### Acceptance criteria

#### AC-Concurrency-1: Support concurrent reads

Given existing keys, when multiple goroutines call `Get` concurrently, then all
calls safely return the values associated with their keys without data races.

#### AC-Concurrency-2: Support concurrent reads and writes

Given multiple goroutines calling `Get` and `Set` for different keys, when the
operations overlap, then no map access races or concurrent-map panics occur and
every successful write remains readable.

#### AC-Concurrency-3: Atomically enforce write-once keys

Given an absent key, when multiple goroutines concurrently call `Set` for that
same key, then exactly one call returns nil, every other call reports
`ErrKeyAlreadyExists`, and the stored value is the one supplied by the single
successful call.

#### AC-Concurrency-4: Remain race-detector clean

Given concurrent coverage of `Get` and `Set`, when the package tests run with
Go's race detector, then the store produces no data-race report.
