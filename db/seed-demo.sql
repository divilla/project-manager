begin;

truncate table
    public.test_case_history,
    public.test_case,
    public.change_history,
    public.change,
    public.epic_history,
    public.epic,
    public.project
restart identity;

insert into public.project (name) values ('demo1'), ('demo2'), ('demo3');

insert into public.epic (project_id, name)
select p.id, seed.name
from public.project p
cross join (values
    ('Echo Router'),
    ('Echo Middleware'),
    ('Echo Binder'),
    ('Echo Documentation'),
    ('Echo Maintenance')
) as seed(name)
where p.name = 'demo1';

update public.project set last_ref = 200 where name = 'demo1';

do
$$
declare
    _project_id bigint;
    _change_id bigint;
begin
    select id into _project_id from public.project where name = 'demo1';

    -- Echo PR pair 1: #3028 creates the Change, #3024 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Router'),
        'Respect q=0 in gzip content negotiation',
        '## Summary

Fix gzip middleware to respect `q=0` values in the `Accept-Encoding` header.

## Problem

According to RFC 9110, an encoding with `q=0` is explicitly unacceptable. The current implementation does not respect this quality value and may enable gzip even when the client sends:

```http
Accept-Encoding: gzip;q=0
```

## Solution

Update the gzip negotiation logic to honor `q=0` quality values and skip gzip when it is explicitly marked as unacceptable.

## Testing

- added a regression test covering `Accept-Encoding: gzip;q=0`
- `go test ./middleware`'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Split out of #3023 at @aldas''s request so the router and JSON changes land as independent PRs.

## What

`findStaticChild` scanned `[]*node` and dereferenced every child just to read its one-byte label — a pointer chase per static step on the routing hot path. This adds a parallel `scLabels []byte` (kept in sync in `addStaticChild`, `newNode`, the insert split, and `Remove`) and scans that contiguous slice instead, indexing into `staticChildren` only on a hit.

It also folds the is-leaf recomputation — previously duplicated across 5 sites, with the `Remove` copy subtly using `len()==0` while the others used `== nil` — into a single `refreshLeaf()` helper that uses `len()`, so it is correct whether `staticChildren` is `nil` or an emptied-but-non-nil slice left after a removal.

## Numbers

```
benchstat (n=6):
RouterStaticRoutes-14   8.49µs -> 8.01µs   -5.67%  (p=0.002)
RouterGitHubAPI-14      16.93µs -> 16.45µs  -2.86%  (p=0.002)
RouterParseAPI / GooglePlusAPI: neutral (param-heavy)
allocations: 0, unchanged
```

The companion ServeHTTP benchmark harness lands with #3023 (the JSON PR); routing benchmarks there exercise this change once both merge.

🤖 Generated with [Claude Code](https://claude.com/claude-code)',
        pull_request_url = 'https://github.com/labstack/echo/pull/3028'
    where id = _change_id;

    -- Echo PR pair 2: #3023 creates the Change, #3020 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Middleware'),
        'perf(json): pooled-buffer JSON deserialize',
        'Per @aldas''s request, this PR is now JSON-only. The router change moved to #3024.

## What

`DefaultJSONSerializer.Deserialize` used `json.NewDecoder(body).Decode()`, which allocates a decoder and its internal read buffer on every JSON request. It now reads the body into a capped pooled buffer and decodes with `json.Unmarshal`; `Unmarshal` does not retain the input slice, so the buffer is safe to reuse. The pool drops oversized buffers (>64 KiB) so a single large body cannot pin memory.

## Numbers

`BenchmarkBind_JSON`: 1008 → 312 B/op (**-69%**), 9 → 7 allocs, ~12% faster.

## Behavioral note

`json.Unmarshal` is stricter than streaming `Decode` — it rejects trailing data after the first top-level value and reports `"unexpected end of JSON input"` for truncated bodies (both still 400 Bad Request). Two bind tests asserting the old `"unexpected EOF"` message are updated to match. Covered by new tests in `json_test.go` (trailing-data rejection, pooled-buffer reuse/concurrency under `-race`, the >64 KiB cap path, and body-read errors).

## DoS note (from review)

The full body is read into memory before decoding. The previous `json.Decoder` path also fully materialized any *valid* large body, so the only real delta is a malformed-early huge body. The correct guard for untrusted input is `middleware.BodyLimit` / `http.MaxBytesReader`, which makes the read here fail fast — documented on `Deserialize`.

This PR also adds the general ServeHTTP/JSON benchmark harness (`perf_bench_test.go`) used to measure both this and #3024.

🤖 Generated with [Claude Code](https://claude.com/claude-code)'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'changelog + version string bump',
        pull_request_url = 'https://github.com/labstack/echo/pull/3023'
    where id = _change_id;

    -- Echo PR pair 3: #3019 creates the Change, #3018 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Binder'),
        'backport PR 3016 from v5 to v4',
        'backport PR #3016 from v5 for https://github.com/labstack/echo/security/advisories/GHSA-vfp3-v2gw-7wfq

-----------

Make serving static file releated methods  and middleware not unescape path by default - so how the way Router interprets paths and Static methods/middleware is consistent.

Given following situation:
```go
// 0.
// given folder structure:
// private.txt
// public/
// public/index.html
// public/text.txt
// public/admin/private.txt

// 1. share `public/` folder contents from the server root. This folder actually contains subfolder `admin` which
// contents we want to forbid from downloading
e.Static("/", "public")

// 2. naively assume that everything under /admin folder is now forbidden
e.GET("/admin/*", func(c *Context) error {
    return ErrForbidden
})
```

Then requests to `/admin%2fprivate.txt` would not be matched to `GET /admin/*` route (routing does not look unescaped path) and static file serving will use unescaped path to serve the file.

Note: this way of "guarding" subfolders will never work for for paths like `/assets/../admin%2fprivate.txt` which will `path.Clean("/assets/../admin%2fprivate.txt")` to `/admin/private.txt` and are servable if static file serving is configured to unescape paths.

If you want to guard routes - use middlewares on `Static*` methods and before `Static` middleware.

 **Breaking change / migration:** If you serve files whose names contain URL-encoded characters (e.g., `/hello%20world.txt` → `hello world.txt`), you must now opt in:

```go
	e := echo.New()
	e.EnablePathUnescapingStaticFiles = true  // <-- enable old behavior
	e.Static("/", "public")
```
for static middleware
```go
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		EnablePathUnescaping: true, // <-- enable old behavior
	}))
```'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'remove dependency on labstack/echo v5 introduced in go.mod and go.sum in https://github.com/labstack/echo/pull/3017',
        pull_request_url = 'https://github.com/labstack/echo/pull/3019'
    where id = _change_id;

    -- Echo PR pair 4: #3017 creates the Change, #3016 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Documentation'),
        'Update CI action versions for v4 branch',
        'Update CI action versions for v4 branch'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Revert PR #3009 changes to just disabling path escaping by default in static methods/middleware

#3009 is a little bit brute force hack to solve the problem from LLM. Claude proposed checking and fixing path used is not a maintainable solutions and there could be so many clever ways how bad actors try to manipulate the path, and the root cause is that the Router and code serving Static files are conceptionally using path differently - so more maintainable solution is to make these 2 acting consitently.

Note: disabling path escaping in static methods and static middleware is a breaking change.',
        pull_request_url = 'https://github.com/labstack/echo/pull/3017'
    where id = _change_id;

    -- Echo PR pair 5: #3015 creates the Change, #3014 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Maintenance'),
        'fix(static): disable static path unescaping by default to prevent ACL bypass (GHSA-3pmx-cf9f-34xr)',
        '## Summary

Fixes **GHSA-3pmx-cf9f-34xr**, a bypass of the GHSA-vfp3-v2gw-7wfq fix (#3009).

The router matches the raw, still-encoded request path, so encoded separators and dot segments in a static wildcard are not treated as traversal during routing. Unescaping them in the static file resolver afterwards let an unauthenticated attacker read a file across a route-level middleware guard the encoded path never matched:

- `/public/%2E%2E/admin/secret.txt` → `admin/secret.txt` — **high severity, default router**
- `/public%2F..%2Fadmin%2Fsecret.txt` under `UseEscapedPathForMatching=true`, where the router decodes the path itself before the handler sees it — lower severity (non-default config)

Both returned `200` + the protected file on `master` at `5786024` (the exact commit the advisory tested).

## Approach

Rather than keep extending an encoding denylist (the original fix blocked `%2F`/`%5C` but missed `%2E%2E`), this addresses the root cause: **make static path unescaping opt-in.** This follows @aldas''s proposal (`e85ee8f`), rebased onto current `master`.

- **echo:** `Config.EnablePathUnescapingStaticFiles` (default `false`) controls unescaping for `Echo.Static`/`StaticFS` and `Group.Static`/`StaticFS`.
- **middleware:** `StaticConfig.EnablePathUnescaping` replaces the now-deprecated `DisablePathUnescaping`; the default is the safe, no-unescape mode.

With unescaping off, `%2F`/`%5C`/`%2E%2E` stay literal and never become separators or traversal.

As defense in depth — and to also close the `UseEscapedPathForMatching` variant, where the **router** (not the handler) does the decoding — any `..` path segment in the resolved wildcard is rejected via `pathutil.HasDotDotSegment`, mirroring the `fs.ValidPath` "no `..` element" invariant. The existing encoded-separator guard remains as a backstop on the opt-in unescaping path.

## Verification

| Variant | `master` (5786024) | this PR |
|---|---|---|
| `/public/%2E%2E/admin/secret.txt` (default router) | 200 `TOP-SECRET` | **404** |
| `/public%2F..%2Fadmin%2Fsecret.txt` (`UseEscapedPathForMatching`) | 200 `TOP-SECRET` | **404** |
| middleware equivalents (both router modes) | 200 `TOP-SECRET` | **404** |

Regression tests cover both variants in both router modes, the opt-in mode (encoded `%20` filenames serve, but `%2F`/`..` still rejected), and `pathutil.HasDotDotSegment` units. `go test -race ./...` and `go vet ./...` pass.

## ⚠️ Breaking change

Static files whose names contain URL-encoded characters (e.g. `"hello world.txt"` via `/hello%20world.txt`) are **no longer served by default**. Set `EnablePathUnescapingStaticFiles` (echo) / `EnablePathUnescaping` (middleware) to opt back in. Because this flips a default, suggest releasing as **5.3.0** with an upgrade note rather than a patch.

## Notes for reviewers

- Omitted the `RawPath`-preferring directory-redirect tweak from `e85ee8f` to keep this scoped to the vuln; happy to fold it in.
- cc @aldas — this takes your disable-by-default approach plus the `..` rejection needed to close the `UseEscapedPathForMatching` variant your patch didn''t reach.

🤖 Generated with [Claude Code](https://claude.com/claude-code)'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Replaces all occurrences of `interface{}` with `any` for Go 1.18+ compatibility.',
        pull_request_url = 'https://github.com/labstack/echo/pull/3015'
    where id = _change_id;

    -- Echo PR pair 6: #3013 creates the Change, #3012 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Router'),
        'ci: run workflows on the v4 branch, not just master',
        '## Problem
The **v4** LTS maintenance branch receives **no CI**. Both `echo.yml` (Run Tests) and `checks.yml` (Run checks) trigger only on `master` — a stale snapshot from when v4 *was* the default branch. As a result, PRs targeting `v4` (e.g. the recent security backport #3011) get zero automated testing and must be dispatched manually.

## Fix
Add `v4` to the `push` and `pull_request` branch filters in both workflows.

## Note on this PR
Because `pull_request` workflows run from the **base** branch''s config, this PR itself won''t auto-run CI (the base `v4` doesn''t have the trigger yet — chicken-and-egg). I''ve **dispatched both workflows manually on this branch** to validate the YAML and confirm they pass. After merge, all future v4 pushes/PRs will be tested automatically.

🤖 Generated with [Claude Code](https://claude.com/claude-code)'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Release prep for **v4.15.3** (v4 LTS).

- Bumps `Version` in echo.go 4.15.2 → 4.15.3.
- Adds the v4.15.3 `CHANGELOG.md` entry.

Headline: fixes **GHSA-vfp3-v2gw-7wfq** — the encoded path separator static bypass (v4 backport of #3009, merged in #3011). After merge, tag `v4.15.3`, publish the release, and amend the advisory to add the v4 affected product.

🤖 Generated with [Claude Code](https://claude.com/claude-code)',
        pull_request_url = 'https://github.com/labstack/echo/pull/3013'
    where id = _change_id;

    -- Echo PR pair 7: #3011 creates the Change, #3010 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Middleware'),
        'fix(static): reject encoded path separators that bypass route-level middleware (v4 backport)',
        'v4 backport of #3009 (released in v5.2.0) for **GHSA-vfp3-v2gw-7wfq**.

## Summary
An encoded path separator (`%2F` or `%5C`) in a static file URL could bypass route-level access control and disclose files. The router matches against the raw, still-encoded path, so `%2F` is not a separator — `/admin%2Fsecret.txt` skips a protected `/admin/*` group, falls through to static serving, which then unescaped `%2F`→`/` and served `admin/secret.txt`.

v4 is affected on **both** static surfaces:
- `echo_fs.go` `StaticDirectoryHandler` (used by `Static`/`StaticFS`) — vulnerable to `%2F` **and** `%5C` (it used OS-specific `filepath.Clean`).
- `middleware/static.go` — vulnerable to `%2F` (it already used `path.Clean`, so not `%5C`).

## Fix
- Both surfaces reject a wildcard containing an encoded separator (`%2F`/`%2f` or `%5C`/`%5c`) with `404` before unescaping, via a shared `internal/pathutil` helper.
- `StaticDirectoryHandler` switched from `filepath.Clean`+`ToSlash` to `path.Clean` (OS-independent; keeps backslash literal on Windows).

## Tests
- New regression tests for `%2F`, `%5C`, double-encoded `%252F`, group `StaticFS`, and the static middleware on a group; unit test for the detector.
- Updated two existing cases (`open redirect vulnerability`, `Directory redirect#01`) that asserted the old `%2f`-unescaped redirect — they now correctly assert `404` + no `Location`.
- `go test -race ./...` and `go vet ./...` pass.

Targets the **v4** branch for a **v4.15.3** release; the advisory will be amended to add the `github.com/labstack/echo` (v4) affected product once tagged.

🤖 Generated with [Claude Code](https://claude.com/claude-code)'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Release prep for **v5.2.0**.

- Bumps `version.go` 5.1.1 → 5.2.0 (minor: the diff since v5.1.1 includes `feat(middleware): RateLimiterStoreContext` #3007).
- Adds the v5.2.0 `CHANGELOG.md` entry.

Headline: fixes **GHSA-vfp3-v2gw-7wfq** (encoded path separator static bypass, #3009). After merge, tag `v5.2.0` and publish the GitHub release.

🤖 Generated with [Claude Code](https://claude.com/claude-code)',
        pull_request_url = 'https://github.com/labstack/echo/pull/3011'
    where id = _change_id;

    -- Echo PR pair 8: #3009 creates the Change, #3008 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Binder'),
        'fix(static): reject encoded path separator to prevent route-level auth bypass',
        '## Summary

Fixes an access-control bypass / static file disclosure where an **encoded slash (`%2F`) lets a request skip a protected route group yet still resolve the file on disk** (GHSA-vfp3-v2gw-7wfq, CWE-22).

The router matches routes against the raw, still-encoded request path, so `%2F` is **not** treated as a path separator — `/admin%2Fsecret.txt` is a single segment and never matches a protected `/admin/*` group. The request falls through to the static handler, which then `url.PathUnescape`d `%2F` back to `/` and resolved `admin/secret.txt` from disk. Router and file handler disagreed on what counts as a separator.

### Reproduction (before)
```
GET /admin/secret.txt    -> 403  (protected group fires)
GET /admin%2Fsecret.txt  -> 200  body="TOP-SECRET"   ← bypass + disclosure
```

Common affected pattern:
```go
adminGroup := e.Group("/admin", authMiddleware)
e.StaticFS("/", os.DirFS("public"))
```

## Fix

`StaticDirectoryHandler` now rejects a wildcard containing an encoded path separator (`%2F`/`%2f` or `%5C`/`%5c`) with `404` **before** unescaping, keeping the router and the file handler consistent. No real filename contains a path separator, so legitimate static requests are unaffected.

## Tests
- New `static_encoded_separator_test.go`: regression test for the bypass + unit test for the `%2F`/`%5C` detector (incl. lower-case hex and a benign `%25` case).
- Updated the existing `TestEcho_StaticFS` "open redirect" case: `…%2f..` now returns `404` with **no redirect emitted at all** (was a sanitized `301`), which closes that vector harder.
- `go test ./ ./middleware/` and `go vet ./` pass.

## Scope / notes
- This addresses the **static-disclosure** vector. The related report GHSA-xr3h-5475-xgqr (percent-encoded routing bypassing protected middleware on *non-static* routes) is a broader router-level decision and is intentionally **not** included here.
- Reported by @a-tt-om and @oran-gugu.

🤖 Generated with [Claude Code](https://claude.com/claude-code)'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## Summary

Optimizes Echo''s per-request hot paths to remove avoidable allocations and CPU work. **No public API changes; the standard-library JSON serializer remains the default.** All numbers are `benchstat` medians (n=8, Apple M3 Max / arm64, Go 1.26).

> **Note (per @aldas review):** the opt-in sonic serializer was **removed** from this PR — it belongs in `echox/cookbook` as a runnable example, not as a submodule in core. This PR is now purely core hot-path optimizations. See "Using a faster JSON encoder" below.

## What changed

**Core**
- **Middleware chain compiled once** (`echo.go`, `buildRouterChains`) and reused, instead of re-wrapping closures on every request. Routing stores the matched handler on the `Context`.
- **Context** (`context.go`): zero-copy `String`/`HTML`/`JSONP` writes (write-only `unsafe` view), reuse of `delayedStatusWriter` (guarded against re-entrant `c.JSON`) and the store map across requests, inline `Get`/`Set` unlock, and a single-key `QueryParam` fast path proven byte-for-byte equal to `url.ParseQuery().Get` (incl. malformed escapes / `;` / `+`).
- **Binder** (`bind.go`): per-`reflect.Type` field-metadata cache so struct tags are parsed once per type, not per request. Preserves the field-name error wrapping from #3005.
- **Middleware**: precompute the HSTS header once (`secure.go`); pool the request-ID `randomString` scratch buffers (`util.go`).
- New hot-path benchmark suite + pooling/dispatch regression tests.
- `test:` de-flake `TestStartConfig_WithListenerNetwork` (ephemeral port instead of a hard-coded one) — separable commit; fixes a pre-existing CI flake.

## Performance (before → after)

| Path | Before | After | Δ time | Allocs |
|---|--:|--:|--:|:--:|
| 5-middleware request | 101 ns | **34 ns** | **−66%** | 5 → **0** |
| `Set` per request | (1 map alloc) | **0 allocs** | — | 1 → **0** |
| `QueryParam` (single key) | 199 ns | **41 ns** | **−79%** | 4 → **0** |
| `String()` response | 191 ns | 188 ns | flat | 4 → **3** |
| `JSON()` response | 347 ns | 350 ns | flat | 5 → **4** |
| `Bind` query (5 fields) | 961 ns | **688 ns** | **−28%** | 8 |
| `bindData` w/ tags | 4973 ns | **2609 ns** | **−48%** | — |
| request-ID gen | 130 ns | 122 ns | −6% | 2 → **1** (−60% B) |
| Static / Param route | 27 / 42 ns | 27 / 43 ns | flat | 0 |

Headline: the middleware path and the `Set`/`QueryParam` paths are now **allocation-free**; binding is **28–48% faster**.

## Router — profiled, intentionally untouched

`-cpuprofile` shows the router is already **0 allocs/op**, with time dominated by the irreducible LCP byte-loop (58%) and method switch (11%). I implemented the httprouter `indices`/`IndexByte` trick for `findStaticChild` and **measured a 30–37% regression** on hits — Echo''s nodes have small fan-out, where the inlined linear scan beats a non-inlined `IndexByte` call — so it was reverted. No router change.

## Using a faster JSON encoder (e.g. sonic)

This PR does not bundle sonic. The `echo.JSONSerializer` interface already lets any app swap encoders in ~10 lines:

```go
import "github.com/bytedance/sonic"

type sonicJSON struct{}
func (sonicJSON) Serialize(c *echo.Context, v any, _ string) error {
	b, err := sonic.Marshal(v); if err != nil { return err }
	_, err = c.Response().Write(b); return err
}
func (sonicJSON) Deserialize(c *echo.Context, v any) error {
	return sonic.ConfigDefault.NewDecoder(c.Request().Body).Decode(v)
}
// e.JSONSerializer = sonicJSON{}
```

Measured (this machine, arm64): sonic **decode −44%** (a clear win on any arch), **encode +43%** (arm64 is sonic''s weak arch; usually a win on amd64). A full cookbook example with these caveats will be a separate PR to labstack/echox.

## Testing

- `go test ./...` + `-race` pass; `gofmt` + `go vet` clean.
- Added: store no-leak across `Reset`, JSON status across `Reset`, nested `c.JSON`, global/pre middleware on 404/405/OPTIONS, `randomString` concurrency, query fast-path stdlib-equivalence.',
        pull_request_url = 'https://github.com/labstack/echo/pull/3009'
    where id = _change_id;

    -- Echo PR pair 9: #3007 creates the Change, #3006 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Documentation'),
        'feat(middleware): optional RateLimiterStoreContext for response headers (#2961)',
        '## Problem (#2961)

The rate limiter sets no `Retry-After` / `X-RateLimit-*` headers, and the `RateLimiterStore` interface (`Allow(identifier) (bool, error)`) gives a store no way to set them either.

## Fix

Add an **optional** interface:

```go
type RateLimiterStoreContext interface {
	AllowContext(c *echo.Context, identifier string) (bool, error)
}
```

When the configured store implements it, the middleware calls `AllowContext` (with the request context) instead of `Allow`, so the store can set response headers on the allow/deny decision.

**Fully backward compatible** — stores implementing only `Allow` are unaffected; the existing interface and the built-in memory store are unchanged.

This is the optional-interface approach @aldas proposed in the issue thread (mirroring the pattern used in the v4 proxy middleware). It intentionally does **not** retrofit the built-in store with full metadata plumbing (the part flagged as a larger rewrite in the thread) — it just provides the hook so stores can set headers.

## Test

`TestRateLimiter_storeAllowContextIsPreferred` (written first; fails before the change): a store implementing `AllowContext` is preferred over `Allow` and can set a `Retry-After` header. gofmt/vet clean; full rate-limiter suite passes (backward compat confirmed).

Addresses #2961.

🤖 Generated with [Claude Code](https://claude.com/claude-code)'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## Problem (#2599)

A file whose name contains a percent sign cannot be downloaded via the static middleware:

- `100%.txt` → `GET /100%25.txt` → **`invalid URL escape "%.t"`** (500)
- `foo%20bar.txt` → `GET /foo%2520bar.txt` → serves the wrong/missing file

`http.Request.URL.Path` is **already decoded** by net/http (per its docs), but the middleware unescaped it a second time.

## Fix

Default `pathUnescape` to `false`. The non-group path comes from `URL.Path` (already decoded), so it must not be unescaped again. Only the wildcard param from a group route (`c.Param("*")`, set explicitly) may still be escaped, and that case keeps its existing `DisablePathUnescaping` toggle — so group behavior is unchanged.

```
GET /100%25.txt       → 200  "hundred percent"
GET /foo%2520bar.txt  → 200  "literal percent twenty"
```

## Test

`TestStatic_servesFileWithPercentInName` (written first; fails on master with 500) using an in-memory `fstest.MapFS`. gofmt/vet clean; all `TestStatic*` and the full middleware package pass (group/`DisablePathUnescaping` cases unaffected).

Fixes #2599.

🤖 Generated with [Claude Code](https://claude.com/claude-code)',
        pull_request_url = 'https://github.com/labstack/echo/pull/3007'
    where id = _change_id;

    -- Echo PR pair 10: #3005 creates the Change, #3004 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Maintenance'),
        'fix(binder): include field name in bind conversion errors (#2629)',
        '## Problem (#2629)

When `c.Bind()` fails a type conversion on form/struct data, the error gives no indication of *which* field failed:

```
POST /submit  number=10a
→ code=400, message=Bad Request, err=strconv.ParseInt: parsing "10a": invalid syntax
```

This makes it hard to render a useful validation message ("the `number` field must be an integer").

## Fix

`bindData` already has the field name (`inputFieldName`) in scope at each conversion site but returned the bare error. Wrap those returns with the field name using `%w` (so `errors.Is`/`errors.As` still work):

```
→ code=400, message=Bad Request, err=number: strconv.ParseInt: parsing "10a": invalid syntax
```

This is the wrap the reporter proposed in the thread (`fmt.Errorf("%s: %w", inputFieldName, err)`), applied to all four conversion-error sites (unmarshaler, slice element, scalar).

## Test

`TestBind_formConversionErrorIncludesFieldName` (written first; fails on master — error contains no field name). Four existing tests that asserted the bare message are updated to the field-prefixed form. gofmt/vet clean; full root-package suite passes.

Fixes #2629.

🤖 Generated with [Claude Code](https://claude.com/claude-code)'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## Problem (#2771)

A binding error returned from a handler is serialized as `{"message":"Bad Request"}` — the field name **and** the binder message are both lost.

### Root cause
`BindingError` embeds `*HTTPError` but does not implement `json.Marshaler`. In `DefaultHTTPErrorHandler` the type switch runs on the `HTTPStatusCoder` extracted via `errors.As`, whose dynamic type is `*BindingError`. Go''s `case *HTTPError` matches only the exact type, so a `*BindingError` falls through to the `default` branch → `{"message": http.StatusText(code)}`. Regression from fbfe216 (#2456).

Verified on current `master`:
```
GET /doc?docNum=abc  →  400  {"message":"Bad Request"}
```

## Fix

Implement `MarshalJSON` on `*BindingError` so it takes the handler''s existing `case json.Marshaler` branch (which is checked *before* `*HTTPError`). Restores the v4.10.2 structured response:
```
GET /doc?docNum=abc  →  400  {"field":"docNum","message":"failed to bind field value to int"}
```

This is the approach the maintainer outlined in the issue thread ("we could make `echo.BindingError` … implement `json.Marshaler`").

## Test

`TestBindingError_serializesToStructuredJSON` (written first; fails on master with `field=<nil>`, `message="Bad Request"`). gofmt/vet clean; full root-package suite passes.

Fixes #2771.

🤖 Generated with [Claude Code](https://claude.com/claude-code)',
        pull_request_url = 'https://github.com/labstack/echo/pull/3005'
    where id = _change_id;

    -- Echo PR pair 11: #3003 creates the Change, #3002 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Router'),
        'test: lock in v5 group route method-handling (405 + OPTIONS)',
        '## What

Adds `group_method_handling_test.go`: tests that lock in v5''s method-handling semantics for routes registered through a `Group`, and act as a regression gate for changes that would reintroduce an implicit per-group catch-all (e.g. #2996, which restores it to fix CORS-on-group preflight).

## Behavior asserted (verified on `master`)

| Request on a `GET`-only group route | v5 today |
|---|---|
| `POST /api/users` | **405** + `Allow: OPTIONS, GET` |
| `OPTIONS /api/users` | **204** + `Allow: OPTIONS, GET` (preflight-relevant) |
| `GET /api/users` | 200 |
| `GET /health` (root) | 200, unaffected by the group |

## Why these are real gates

A group-level catch-all — whether registered manually via `g.RouteNotFound("/*", …)` or automatically (as in #2996) — matches **every** method, which masks both the 405 and the automatic OPTIONS response as **404**. Verified empirically: with such a catch-all, `POST` and `OPTIONS` on `/api/users` both return 404. The first two tests would catch that regression.

## Note

This replaces an earlier version of this PR whose comments described a middleware-triggered "auto catch-all" that does not exist on `master` (v5) — the tests passed for the wrong reason. Reworked after review to assert the actual, verified behavior and drop the inert no-op middleware. All four tests pass; gofmt- and vet-clean.

🤖 Generated with [Claude Code](https://claude.com/claude-code)'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## Why

Echo is actively maintained and shipping (v5.1.1 + v4.15.2 on 2026-05-01, `master` commits within days), but to a casual visitor the repo can read as inactive. This PR adds low-cost, **self-updating** signals that Echo is alive and clarifies its positioning.

## Changes

- **Dynamic badges** for latest release and last commit. These pull live from GitHub, so they can never go stale the way a hand-written "last updated" line does.
- **Positioning line** under the tagline explaining what Echo adds on top of Go''s `net/http` — a gentle answer to the "do I still need a framework after Go 1.22 routing?" question.
- **"Actively maintained" note** pointing at the badges.
- **ROADMAP.md (DRAFT)** documenting the version-support policy (v5 current, v4 LTS until 2026-12-31) and a Now/Next/Later view seeded from current issues. Marked DRAFT — maintainers own the final content. Linked from the README.

No code changes. Docs only.

🤖 Generated with [Claude Code](https://claude.com/claude-code)',
        pull_request_url = 'https://github.com/labstack/echo/pull/3003'
    where id = _change_id;

    -- Echo PR pair 12: #3000 creates the Change, #2994 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Middleware'),
        'fix(middleware): reset ContentLength after gzip decompression',
        '## Summary

Fix `Decompress` so middleware after it does not keep using the compressed request size after the body has been replaced with the decoded gzip stream.

For example:

```go
e.Use(middleware.Decompress())
e.Use(middleware.BodyLimit(4))
```

A gzipped request whose decoded body is `ok` should pass:

```bash
printf "ok" | gzip | curl -X POST \
  -H "Content-Type: text/plain" \
  -H "Content-Encoding: gzip" \
  --data-binary @- \
  http://localhost:8080/
```

Before this change, `Decompress` replaced `req.Body` but left `req.ContentLength` set to the compressed size, so `BodyLimit` could reject the request before reading it:

```text
{"message":"Request Entity Too Large"}
```

With this change, `Decompress` sets `req.ContentLength = -1` after gzip decompression is set up, so downstream middleware enforces limits while reading the decoded body:

```text
ok
```

This intentionally does not change `Content-Encoding`, `GetBody`, or multiple content-coding behavior.

## Test plan

- [x] `go test -count=1 ./...`
- [x] `go test -race -count=1 ./...`
- [x] `go vet ./...`
- [x] `staticcheck ./...`
- [x] Manual curl check with the same handler returned `ok` and HTTP 200'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## Description

This PR fixes https://github.com/labstack/echo/issues/2993

The proxy middleware''s WebSocket path currently sets `X-Forwarded-For` only when the header is empty, dropping the proxy''s own peer IP from the chain whenever upstream proxies had already added entries. This breaks downstream services that rely on the "rightmost untrusted" rule to extract the real client IP, including echo''s own `ExtractIPFromXFFHeader`.

The HTTP path delegates to `net/http/httputil.ReverseProxy`, which appends `RemoteAddr` to the existing `X-Forwarded-For` chain — either inline in `ServeHTTP`''s default Director path ([reverseproxy.go#L519-L531](https://github.com/golang/go/blob/go1.26.3/src/net/http/httputil/reverseproxy.go#L519-L531)) or via the explicit [`(*ProxyRequest).SetXForwarded`](https://github.com/golang/go/blob/go1.26.3/src/net/http/httputil/reverseproxy.go#L82-L96)
when a `Rewrite` callback is configured. The WebSocket path uses `proxyRaw`,
which writes the request verbatim via `r.Write(out)`, so this middleware is the only place where the appending happens for WebSocket Upgrade requests.

## Change

Replace the "set if empty" guard with always-append. Read values via map access to preserve multi-line `X-Forwarded-For` headers (RFC 9110 §5.3 allows combining them by joining values with commas).

## Test

Added TestProxyWebSocketXForwardedFor exercising 4 cases:

- no incoming X-Forwarded-For → only c.RealIP()
- single-line single-entry → preserved + c.RealIP() at the tail
- single-line comma-separated → preserved + c.RealIP() at the tail
- multi-line headers (multiple X-Forwarded-For occurrences) → joined with , + c.RealIP() at the tail

Each case captures the request header at WebSocket Upgrade time inside the upstream handler and asserts both the appended tail and the preserved prefix.

## Backwards compatibility

The change is additive: existing entries are preserved and the proxy''s own peer IP is added at the tail. Downstream readers using the standard "rightmost untrusted" rule (e.g. echo.ExtractIPFromXFFHeader) see no behavioral difference for chains where they already worked, and start returning the correct IP for chains where the proxy IP was previously
dropped.

## Background

We hit this in production with an Echo-based WebSocket reverse proxy fronting an internal service that uses echo.ExtractIPFromXFFHeader for IP-based authorization. Multi-hop deployments (customer proxy → our reverse proxy → backend) broke because the reverse proxy''s egress IP was missing from the chain reaching the backend.',
        pull_request_url = 'https://github.com/labstack/echo/pull/3000'
    where id = _change_id;

    -- Echo PR pair 13: #2992 creates the Change, #2990 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Binder'),
        'fix(middleware): correct documented KeyAuth KeyLookup default',
        'The `KeyLookup` field comment in `KeyAuthConfig` documents the default as `"header:Authorization"`, but the default applied by `DefaultKeyAuthConfig` is:

```go
KeyLookup: "header:" + echo.HeaderAuthorization + ":Bearer ",
```

which evaluates to `"header:Authorization:Bearer "` (also the value `ToMiddleware` falls back to when `KeyLookup` is empty). The `:Bearer ` cut-prefix trims the scheme from the header value, exactly as the same comment block describes.

The comment was correct in v4, where `AuthScheme: "Bearer"` was a separate field. In v5 `AuthScheme` was folded into `KeyLookup`, but this comment was left showing the old default.

Doc comment only — no behavior change.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Fixes #2853.

When Echo CORS middleware is run in a chained proxy setup (or in front of any handler copying upstream headers using Add), headers like Access-Control-Allow-Origin and Vary get duplicated in the response.

Changes:
- Run simple request CORS header writing inside a Before hook on the response. This allows the proxy''s CORS config to cleanly Set the headers, overriding duplicated upstream headers from the proxy or downstream response copy.
- Implement addVaryHeader helper to merge and deduplicate Vary values case-insensitively.
- Add unit test simulating ReverseProxy behavior to verify headers are not duplicated.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2992'
    where id = _change_id;

    -- Echo PR pair 14: #2989 creates the Change, #2988 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Documentation'),
        'chore: improve echo maintenance path',
        'Summary:
- Add or tighten focused edge-case tests or type assertions in ip_test.go, json_test.go related to CI, Docker, Kubernetes, build tooling, release automation; avoid docs-only changes and broad refactors.
- Keep the change narrow so it is straightforward to review.

Notes:
- I kept this scoped to the relevant implementation and tests.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Spotted while reading binder.go. The doc comments on `MustUnixTime`, `MustUnixTimeMilli`, `MustUnixTimeNano` all say "bind to time.Duration variable" but the function signature is `dest *time.Time` and the non-Must variants directly above each one correctly say "binds parameter to time.Time variable". Looks like a copy-paste from the actual `MustDuration` doc that never got updated.

While there, dropped a stray double space and changed "nano second" to "nanosecond" on the Nano variant. No code change.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2989'
    where id = _change_id;

    -- Echo PR pair 15: #2986 creates the Change, #2985 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Maintenance'),
        'fix: inherit parent RouteNotFound handler for groups with middleware',
        '## Summary
- When `Group.Use()` registers catch-all RouteNotFound routes for middleware execution, inherit the most specific parent RouteNotFound handler from the router tree
- Fall back to the default `NotFoundHandler` only when no parent handler exists

Fixes #2485

## Problem
Groups created with middleware always registered the default JSON `NotFoundHandler` for 404 cases, shadowing a custom root `RouteNotFound` handler.

## Test plan
- [x] `go test ./...`
- [x] Added `TestGroupRouteNotFoundFallsBackToRootHandler` from the issue reproducer
- [x] Updated `TestGroup_RouteNotFoundWithMiddleware` expectations for inherited handlers'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## Summary

Fixes #2961

Rate limit middleware now sets standard rate limit response headers when using the built-in `RateLimiterMemoryStore`:

- `X-RateLimit-Limit` — configured burst
- `X-RateLimit-Remaining` — tokens remaining after the request
- `Retry-After` — seconds until the next token is available (on 429 only)

Implementation follows the maintainer suggestion from the issue: an optional unexported `rateLimiterStoreContext` interface with `AllowContext(c echo.Context, identifier string) (bool, error)`. When the configured store implements it, the middleware calls it instead of `Allow`. Custom stores can opt in without breaking the existing `RateLimiterStore` API.

`Retry-After` is computed via `rate.Limiter.ReserveN(...).Delay()` as suggested in the issue discussion.

## Test plan

- [x] `go test ./middleware -run TestRateLimiter`
- [x] `go test -race ./middleware -run TestRateLimiterMemoryStore`
- [x] New test `TestRateLimiterMemoryStore_AllowContext_SetsHeaders` verifies headers on allowed and denied requests',
        pull_request_url = 'https://github.com/labstack/echo/pull/2986'
    where id = _change_id;

    -- Echo PR pair 16: #2984 creates the Change, #2983 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Router'),
        'fix(middleware): keep handler 404 responses with Static HTML5 mode',
        '## Summary

- Restrict Static middleware HTML5 SPA fallback to router-level 404 responses
- Skip index.html fallback when a matched handler returns its own `404` (e.g. API resource not found)
- Add regression test covering `/api/test/3` JSON 404 vs `/client-route` SPA index fallback

## Test plan

- [x] `go test ./middleware/...`
- [x] `TestStaticHTML5DoesNotOverrideHandler404`

Fixes #2775'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## Summary

- For simple (non-preflight) CORS requests, apply CORS response headers after the handler runs
- Skip setting `Access-Control-Allow-Origin` and related headers when an upstream handler (e.g. reverse proxy) already set them
- Add regression test covering chained CORS + reverse proxy setup from #2853

## Test plan

- [x] `go test ./middleware/...`
- [x] `TestCORSNoDuplicateHeadersFromUpstream` — proxy layer + upstream both use CORS middleware, response has single `Access-Control-Allow-Origin` and `Vary`

Fixes #2853',
        pull_request_url = 'https://github.com/labstack/echo/pull/2984'
    where id = _change_id;

    -- Echo PR pair 17: #2982 creates the Change, #2979 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Middleware'),
        'fix(bind): restore []string values for map[string]interface{} duplicates',
        '## Summary
- Fix regression when binding multipart form data with duplicate field names to `map[string]interface{}`
- Store a single value as `string` and multiple values as `[]string`

## Problem
Since v4.13.0, binding duplicate multipart fields like two `ima_slice` values to `map[string]interface{}` only kept the first value. Applications expecting a slice silently broke.

This regressed after #2656, which intended to preserve pre-v4.12.0 single-string behavior but always bound `v[0]`.

## Fix
For `map[string]interface{}` / `map[string]any`:
- one value → `string`
- multiple values → `[]string`

## Test plan
- [x] Updated `TestDefaultBinder_bindDataToMap`
- [x] Added `TestBindMultipartFormToMapInterface`
- [x] `go test ./... -run ''TestDefaultBinder_bindDataToMap|TestBindMultipartFormToMapInterface''`

Fixes #2731'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Fixes typos in CSRFConfig comments so they reference the actual exported field names:

- TrustedOrigin -> TrustedOrigins
- AllowSecFetchSameSite -> AllowSecFetchSiteFunc
- CRSF -> CSRF

Also clarifies the trusted origin wording.

No behavior changes.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2982'
    where id = _change_id;

    -- Echo PR pair 18: #2977 creates the Change, #2973 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Binder'),
        'Fix proxy panic when balancer has no targets',
        'Fixes #2976

This fixes a panic when a proxy balancer returns no target.

Built-in balancers can return `nil` when the target list is empty, for example after `RemoveTarget` removes the last target. `ProxyWithConfig` now returns `502 Bad Gateway` through the configured error handler instead of passing the nil target to `proxyHTTP` or `proxyRaw`.

I also added a test for a custom balancer returning a target with a nil URL.

Tests pass with `go test ./... -count=1`.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Enables zero-copy (sendfile) serving. Disabled when ''After'' hooks are present to maintain backward compatibility.

fix #2725
-----

```
Environment
   - OS: darwin
   - Arch: amd64
   - CPU: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
   - Go: goos: darwin, goarch: amd64

  Benchmark Results (100MB File)

   1 BenchmarkContext_File_RealServer/Zero-Copy-Optimized-12       32718676 ns/op    3204.82 MB/s
   2 BenchmarkContext_File_RealServer/User-Space-Standard-12       40801873 ns/op    2569.92 MB/s
```',
        pull_request_url = 'https://github.com/labstack/echo/pull/2977'
    where id = _change_id;

    -- Echo PR pair 19: #2971 creates the Change, #2970 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Documentation'),
        'Update GitHub actions deps versions',
        'Closed Echo pull request #2971 did not include a body.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Modernizes the codebase using the Go 1.26 gofix tool to adopt newer idioms and library features while maintaining compatibility with the current toolchain.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2971'
    where id = _change_id;

    -- Echo PR pair 20: #2969 creates the Change, #2966 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Maintenance'),
        'refactor: replace Split in loops with more efficient SplitSeq',
        'strings.SplitSeq (introduced in Go 1.23)  returns a lazy sequence (strings.Seq), allowing gopher to iterate over tokens one by one without creating an intermediate slice.

It significantly reduces memory allocations and can improve performance for long strings.

More info: https://github.com/golang/go/issues/61901'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'In Go 1.21, the standard library includes built-in [max/min](https://pkg.go.dev/builtin@go1.21.0#max) function, which can greatly simplify the code.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2969'
    where id = _change_id;

    -- Echo PR pair 21: #2965 creates the Change, #2964 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Router'),
        'Changelog for v5.1.1',
        'Closed Echo pull request #2965 did not include a body.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Context.Json should not unwrap response and just wrap Response so other middlewares can use their own "wrapping" Responses and see the status code.

I found this during #2895 when to tried to create middleware that wraps existing response to own and status code setting did not work anymore with `Context.JSON` (always sends 200 to client).',
        pull_request_url = 'https://github.com/labstack/echo/pull/2965'
    where id = _change_id;

    -- Echo PR pair 22: #2963 creates the Change, #2962 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Middleware'),
        'Changelog for v4.15.2',
        'changelog'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Context.Scheme should validate values taken from header

Backport PR #2953 (d1d8ad3f99dd9b80542cd0c357d56a8916c515df) to `v4`',
        pull_request_url = 'https://github.com/labstack/echo/pull/2963'
    where id = _change_id;

    -- Echo PR pair 23: #2958 creates the Change, #2953 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Binder'),
        'chore: fix typos in httperror.go',
        'Closed Echo pull request #2958 did not include a body.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Relates to: https://github.com/labstack/echo/issues/2952',
        pull_request_url = 'https://github.com/labstack/echo/pull/2958'
    where id = _change_id;

    -- Echo PR pair 24: #2951 creates the Change, #2946 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Documentation'),
        'docs: fix README command and slog reference',
        '**Repo:** labstack/echo (⭐ 30000)
**Type:** docs
**Files changed:** 1
**Lines:** +3/-3

## What
Updated `README.md` to fix three user-facing documentation issues: the installation snippet now uses a valid shell comment marker inside the `sh` code block, the `slog-echo` entry now links to the current standard-library `log/slog` package documentation instead of the retired `x/exp/slog` path, and the third-party middleware guidance sentence has a small grammar correction.

## Why
These changes remove small but real sources of friction in the first-touch documentation. The `//` line inside a shell block is misleading if copied verbatim, and the outdated `slog` link points readers at a pre-standard-library package path. The grammar fix keeps contributor-facing guidance polished without changing behavior.

## Testing
No code tests were run because this is a README-only change. Verification was done by inspecting the rendered diff and ensuring the updated commands and links are syntactically correct.

## Risk
Low / Documentation-only change with no runtime impact.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Adds a new middleware that mounts a JSON-RPC 2.0 endpoint at a configurable path and auto-exposes registered Echo routes as MCP tools, so AI clients (Claude Desktop, Cursor, etc.) can discover and call them.

Implements initialize, tools/list and tools/call. Tool input schemas are derived from RouteInfo.Parameters; tool calls substitute path parameters via RouteInfo.Reverse and dispatch the synthesized request through e.ServeHTTP, preserving the full middleware chain.

No core Echo files are modified and no new dependencies are introduced.

Fixes #2935',
        pull_request_url = 'https://github.com/labstack/echo/pull/2951'
    where id = _change_id;

    -- Echo PR pair 25: #2945 creates the Change, #2944 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Maintenance'),
        'feat(middleware): add MCP (Model Context Protocol) middleware',
        'Adds a new middleware that mounts a JSON-RPC 2.0 endpoint at a configurable path and auto-exposes registered Echo routes as MCP tools, so AI clients (Claude Desktop, Cursor, etc.) can discover and call them.

Implements initialize, tools/list and tools/call. Tool input schemas are derived from RouteInfo.Parameters; tool calls substitute path parameters via RouteInfo.Reverse and dispatch the synthesized request through e.ServeHTTP, preserving the full middleware chain.

No core Echo files are modified and no new dependencies are introduced.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## What this does

This adds an **optional `AutoHead` flag** that automatically enables HEAD requests for any GET route. No need to define separate HEAD handlers anymore.

## Why this is useful

Right now, if you want proper HEAD support, you have to manually add a HEAD route for every GET route. That leads to:

* Repeated boilerplate
* Easy-to-miss HEAD routes (resulting in 405 errors)
* Inconsistent headers like missing `Content-Length`

According to HTTP spec (RFC 7231), HEAD should behave like GET without a body — so this just makes Echo handle that for you.

## How it works

When `AutoHead` is turned on:

* Every GET route automatically gets a HEAD equivalent
* The same handler runs, but the response body is suppressed
* Headers and status code are preserved
* `Content-Length` is still set correctly
* If you’ve defined a HEAD route manually, it won’t be overridden

## Implementation notes

* Added `AutoHead` to `Echo` and `Config`
* Wrapped handlers using a custom response writer
* Hooked into route registration (`add()`) to register HEAD routes

## Usage

```go
e := echo.New()
e.AutoHead = true

e.GET("/api/users", handler)
// HEAD /api/users now works automatically
```

## Performance

Very small overhead:

* ~532 ns per request
* 3 allocations

No impact if the feature is disabled (default).

## Testing

Covers:

* Enabled vs disabled behavior
* Explicit HEAD route priority
* Middleware compatibility

All tests pass.

---

## Compatibility

* No breaking changes
* Fully opt-in
* Existing code works as-is
#2895 @markbates @aldas',
        pull_request_url = 'https://github.com/labstack/echo/pull/2945'
    where id = _change_id;

    -- Echo PR pair 26: #2941 creates the Change, #2937 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Router'),
        'fix(lint): resolve staticcheck issues and improve code quality',
        'This PR resolves several `staticcheck` lint issues, got the report from  `golangci-lint`.

### Changes

* Applied De Morgan’s law to simplify boolean expressions
* Replaced `fmt.Sprintf` with `fmt.Fprintf` to avoid unnecessary allocations
* Removed redundant embedded field access
* Improved overall code readability and idiomatic Go usage

### Impact

* No functional changes
* Minor performance improvements
* Cleaner and more maintainable code

### Notes

All changes were verified with:

```bash
golangci-lint run
```

there is no remaining lint issues.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'the behavior is opt-out.

I created one flag in both Echo and Group struct (since they are related to the register of new routes) that is private and a function to explicit cancel this behavior.

Why: Mentioned in the issue #2895 I searched and saw that the default behavior in many frameworks is to automatically register a head request with GET, so I agree with the author of the issue that it should be included to guarantee an expected behavior from the programmer.

I added tests and only modified the high level functions, if the author think it is good and relevant enough to be merged, it will be good.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2941'
    where id = _change_id;

    -- Echo PR pair 27: #2936 creates the Change, #2934 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Middleware'),
        'Make StartConfig listener creation context-aware',
        '## Summary
- Create listeners with net.ListenConfig so StartConfig respects the provided context during listener setup
- Keep the existing serving behavior by defaulting ListenerNetwork to `tcp` and wrapping TLS listeners with
`tls.NewListener`
- simplify the listener creation path by using the same flow for TLS and non-TLS listeners

## Benefit
`StartConfig.Start` and `StartConfig.StartTLS` already accept a context, but listener creation previously used `net.Listen`
and `tls.Listen`, which do not use that context. This change makes listener setup context-aware without changing how the
server behaves after the listener is created.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Closed Echo pull request #2934 did not include a body.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2936'
    where id = _change_id;

    -- Echo PR pair 28: #2933 creates the Change, #2932 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Binder'),
        'Remove legacy IP extraction logic from context.RealIP method',
        'This change does not break the API contract, but it does introduce breaking changes in logic/behavior.  But as promised - 31.03.2026 will be last day for potentially breaking stuff.  I want to get this security related thing done.

-------------

Remove legacy IP extraction logic from context.RealIP method and move it to LegacyIPExtractor IP extraction function.

`v4` behavior can be restored with:
```go
e := echo.New()
e.IPExtractor = echo.LegacyIPExtractor()
```

but you should instead with proper trust options
- https://pkg.go.dev/github.com/labstack/echo/v5#ExtractIPFromRealIPHeader
- https://pkg.go.dev/github.com/labstack/echo/v5#ExtractIPFromXFFHeader

For example:
 ```go
_, lbIPRange, _ := net.ParseCIDR("203.0.113.199/24")
e.IPExtractor = echo.ExtractIPFromXFFHeader(
	echo.TrustLinkLocal(false),
	echo.TrustIPRange(lbIPRange),
)
 ```

Read https://echo.labstack.com/docs/ip-address'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Relates to https://github.com/labstack/echox/issues/397 and https://github.com/labstack/echo/issues/2918

we did not set in `v4` and it causing problems for users',
        pull_request_url = 'https://github.com/labstack/echo/pull/2933'
    where id = _change_id;

    -- Echo PR pair 29: #2931 creates the Change, #2930 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Documentation'),
        'Add NewDefaultFS function to help create filesystem that allows absolute paths',
        'Add `NewDefaultFS` function to help create filesystem that allows absolute paths with `Open` method.

fs.FS does not like paths like `/tmp/temp.file` even if fs is created like `os.DirFS("/"). You need to remove leading slash.

Also when file is absolute (`/tmp/temp.file`) and has same prefix as filesystem  `os.DirFS("/tmp/")` this would not work. Echo `defaultFs` was working similarly to `os.Open` and therefore allowed absolute paths.

This PR makes `echo.defaultFs` to accept absolute path filenames in Open when it matches the filesystem prefix.

Relates to https://github.com/labstack/echo/issues/2929'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Relates to: https://github.com/labstack/echo/issues/2924

NB: it does not fix couple of staticcheck problems that are being reported
```bash
[x@x echo]$ golangci-lint run
bind.go:158:6: QF1001: could apply De Morgan''s law (staticcheck)
                if !(isElemSliceOfStrings || isElemString || isElemInterface) {
                   ^
middleware/basic_auth.go:124:8: QF1001: could apply De Morgan''s law (staticcheck)
                                if !(len(auth) > l+1 && strings.EqualFold(auth[:l], basic)) {
                                   ^
middleware/static.go:253:8: QF1001: could apply De Morgan''s law (staticcheck)
                                if !(errors.As(err, &he) && config.HTML5 && he.StatusCode() == http.StatusNotFound) {
                                   ^
route.go:88:4: QF1012: Use fmt.Fprintf(...) instead of WriteString(fmt.Sprintf(...)) (staticcheck)
                        uri.WriteString(fmt.Sprintf("%v", pathValues[n]))
                        ^
router.go:998:30: QF1008: could remove embedded field "RouteInfo" from selector (staticcheck)
                rPath = matchedRouteMethod.RouteInfo.Path
```',
        pull_request_url = 'https://github.com/labstack/echo/pull/2931'
    where id = _change_id;

    -- Echo PR pair 30: #2928 creates the Change, #2925 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Maintenance'),
        'Add doc comments to clarify usage of File related methods and leading slash handling',
        'Add doc comments to clarify usage of File related methods and leading slash handling

Relates to https://github.com/labstack/echo/issues/2922#issuecomment-4150975101

Example:
```go
package main

import (
	"embed"

	"github.com/labstack/echo/v5"
)

//go:embed dist/*
var efs embed.FS

func main() {
	e := echo.New()
	e.Filesystem = efs

	e.File("/test", "dist/private.txt") // <--- file path must not have a leading slash

	if err := e.Start(":8080"); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}

```

and
```go
package main

import (
	"os"

	"github.com/labstack/echo/v5"
)

func main() {
	e := echo.New()
	e.Filesystem = os.DirFS("/")

	e.File("/test.jpg", "opt/app/assets/test.jpg") // <--- file path must not have a leading slash

	if err := e.Start(":8080"); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
```'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'The documentation for `NewRateLimiterMemoryStore` and
`NewRateLimiterMemoryStoreWithConfig` states that the default Burst
value is the "rounded down" value of the rate. This was accurate when
the documentation was added in #2366, where the code used `int(config.Rate)`.

However, #2852 changed the default burst calculation to use `math.Ceil`,
making it the rounded up value. The documentation was not updated to
reflect this change.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2928'
    where id = _change_id;

    -- Echo PR pair 31: #2921 creates the Change, #2920 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Router'),
        'fix: prefer exact path match over parameterized match for method checking',
        '## Summary
- Fix route matching to prefer exact path matches over parameterized matches when checking HTTP methods
- Previously, a POST to `/VerifiedCallerId/Verification/` would match `GET VerifiedCallerId/:phone_number` and return 405

## Root Cause
The router matches parameterized routes before checking for exact matches at the same path level. This causes incorrect 405 responses when an exact match exists for a different method.

## Testing
- Added test case for the reported scenario

Fixes #2547

Made with [Cursor](https://cursor.com)'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Add StartConfig.Listener so server with custom Listener is easier to create

relates to https://github.com/labstack/echo/issues/2918#issuecomment-4089341521
https://github.com/labstack/echo/issues/1942

Example:
```go
package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
)

func main() {
	e := echo.New()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	lc := net.ListenConfig{
		KeepAlive: 15 * time.Second,
		//Control:   nil,
		//KeepAliveConfig: net.KeepAliveConfig{
		//	Enable:   false,
		//	Idle:     0,
		//	Interval: 0,
		//	Count:    0,
		//},
	}
	l, err := lc.Listen(ctx, "tcp", ":8080")
	if err != nil {
		e.Logger.Error("failed to create listener", "error", err)
		return
	}
	defer l.Close()

	sc := echo.StartConfig{
		Listener: l,
	}
	if err := sc.Start(ctx, e); err != nil {
		e.Logger.Error("failed to run server", "error", err)
	}
}

```',
        pull_request_url = 'https://github.com/labstack/echo/pull/2921'
    where id = _change_id;

    -- Echo PR pair 32: #2919 creates the Change, #2917 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Middleware'),
        'Add https://github.com/labstack/echo-prometheus to the middleware list in README.md',
        'Add https://github.com/labstack/echo-prometheus to the middleware list in README.md'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## Summary

Fixes #2485

When a route exists in the router but the HTTP method is not allowed, the router previously always fell back to `methodNotAllowedHandler`. However, this bypassed any `RouteNotFound` handler registered at a parent or root group level.

## Root Cause

In `DefaultRouter.Route()`, when `currentNode.isHandler` is true (path matched but method not), the code immediately set `rInfo = methodNotAllowedRouteInfo` without checking if any parent node had a `notFoundHandler` registered.

## Fix

Traverse the parent node chain to look for a `notFoundHandler`. If found, use it; otherwise fall back to the existing `methodNotAllowedHandler` behavior.

## Test

Existing route tests pass. The fix ensures `RouteNotFound` handlers registered at a group level are properly invoked for sub-paths.

Signed-off-by: lyydsheep <2230561977@qq.com>',
        pull_request_url = 'https://github.com/labstack/echo/pull/2919'
    where id = _change_id;

    -- Echo PR pair 33: #2916 creates the Change, #2910 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Binder'),
        'fix: correct spelling mistakes in comments and field name',
        'Fix multiple typos found across the codebase:

- `response.go`: rename unexported field `commited` to `committed` and fix related comments
- `server.go`: fix `addres` to `address` in comment
- `middleware/csrf.go`: fix `formated` to `formatted` in comment
- `middleware/request_logger.go`: fix `commited` to `committed` in comment
- `echotest/context.go`: fix `wil` to `will` in comments (2 occurrences)

All changes are either in comments or in an unexported (internal) struct field name. No behavioral changes.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Fix for issue: #2761

This PR addresses #2761 by introducing SkipMiddlewareOnNotFound. This allows developers to avoid executing heavy global middleware (Auth, DB logging) for requests that result in a 404 or 405, improving performance and reducing log noise

**Testing with below cmd saves 3 allocations.**

```bash
go test -bench=BenchmarkMiddleware404 -benchmem
```
Sample output:
```
BenchmarkMiddleware404/Normal_404-8             {"time":"2026-03-04T22:00:36.5728429Z","level":"ERROR","msg":"failed to shut down server within given timeout","error":"context deadline exceeded"}
 1416462               890.7 ns/op           894 B/op         10 allocs/op
BenchmarkMiddleware404/Optimized_404-8           1000000              1472 ns/op             515 B/op          7 allocs/op
PASS
ok      github.com/labstack/echo/v5     6.217s
```

10-7 = 3 allocations saved for faster runtime',
        pull_request_url = 'https://github.com/labstack/echo/pull/2916'
    where id = _change_id;

    -- Echo PR pair 34: #2908 creates the Change, #2905 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Documentation'),
        'Add echo-opentelemetry to the README.md',
        'Add https://github.com/labstack/echo-opentelemetry to the middleware list in README.md'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'CSRF: support older token-based CSRF protection handler that want to render token into template

(cherry picked from commit 9183f1e80901fe3e55a61fce607e2c925e1e987b)
same thing in `v5` https://github.com/labstack/echo/pull/2894


relates to:
- https://github.com/labstack/echo/issues/2874
- https://github.com/labstack/echo/pull/2903',
        pull_request_url = 'https://github.com/labstack/echo/pull/2908'
    where id = _change_id;

    -- Echo PR pair 35: #2903 creates the Change, #2902 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Maintenance'),
        'Fix CSRF middleware to set token in context when Sec-Fetch-Site validation passes',
        '## Summary

Fixes #2874 - CSRF middleware now correctly sets the token in context and cookie even when Sec-Fetch-Site validation passes, allowing handlers to render forms with CSRF tokens.

## Problem

In v4.15.0, when `checkSecFetchSiteRequest()` returns `(true, nil)` (e.g., for direct URL navigation with `Sec-Fetch-Site: none`), the middleware calls `return next(c)` without setting the CSRF token in context. This breaks handlers that need the token to render forms.

**Reproduction:**
```bash
curl -H "Sec-Fetch-Site: none" https://example.com/users/register
# Returns 500 "CSRF token not found" in v4.15.0
```

## Solution

Move token generation/retrieval before the Sec-Fetch-Site check and set the token in context and cookie even when the request is deemed "safe". This ensures handlers can always access the CSRF token for form rendering while still skipping token validation for safe requests.

## Changes

- Token generation/retrieval now happens before Sec-Fetch-Site validation
- When a request passes Sec-Fetch-Site validation, the middleware now:
  - Sets the token in context via `c.Set(config.ContextKey, token)`
  - Sets the CSRF cookie with proper expiration and security flags
  - Adds the `Vary: Cookie` header to prevent caching issues
- Token validation is still skipped for safe requests (no behavior change)

## Test Plan

- [x] All existing CSRF tests pass
- [x] Added `TestCSRF_SecFetchSite_SetsTokenInContext` with 4 test cases covering:
  - GET requests with `Sec-Fetch-Site: none`
  - GET requests with `Sec-Fetch-Site: same-origin`
  - POST requests with `Sec-Fetch-Site: none`
  - POST requests with `Sec-Fetch-Site: same-origin`
- [x] All tests verify that:
  - Token is set in context
  - CSRF cookie is set
  - Token in context matches cookie value

## Impact

This fix restores the ability to render server-side forms with CSRF tokens when users access pages via direct browser navigation (typing URL, bookmarks, external links). The `Sec-Fetch-Site: none` header is automatically sent by all modern browsers in these scenarios.

## Backward Compatibility

This change is backward compatible - it only adds functionality (setting token in context) that was missing, without changing any existing validation logic.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## Summary

Fixes issue #2853 - CORS middleware was duplicating headers when multiple Echo services with CORS middleware were chained (e.g., Service A proxies to Service B, both with `middleware.CORS` enabled).

## Changes

- Added check to detect existing `Access-Control-Allow-Origin` headers in responses (indicating the request was proxied from an upstream service that already applied CORS)
- When CORS headers are already present, the middleware now skips re-applying them to prevent duplication
- Updated `Vary` header handling to check if values already exist before adding them, preventing duplicate Vary entries
- Added comprehensive test cases for proxy chain scenarios (both regular and preflight requests)

## Test Plan

- [x] All existing CORS tests pass
- [x] Added `TestCORSProxyChain` to verify headers are not duplicated in proxy scenarios
- [x] Added `TestCORSProxyChainPreflight` to verify preflight requests in proxy chains
- [x] Verified that the fix prevents duplicate `Access-Control-Allow-Origin` and `Vary` headers

## Reproduces Issue

This fix addresses the exact scenario described in #2853 where multiple proxy layers each independently apply CORS headers, causing accumulation.

**Before**: Multiple CORS middlewares in a chain would each add headers, resulting in duplicates
**After**: Middleware detects existing CORS headers and skips processing, preventing duplication',
        pull_request_url = 'https://github.com/labstack/echo/pull/2903'
    where id = _change_id;

    -- Echo PR pair 36: #2901 creates the Change, #2900 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Router'),
        'Add changelog for v5.0.4 release',
        'Closed Echo pull request #2901 did not include a body.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Add `ResolveResponseStatus` function to help middleware/handlers determine HTTP status code and echo.Response.

Loggers and tracing middlewares need to determine status code from either from error or `echo.Response`.  Also - response is needed often for knowing response size from `echo.Response.Size`.  so this function tries to shorten these 2 requirements.

Relates to https://github.com/labstack/echo-contrib/pull/141
and https://github.com/labstack/echo-contrib/blob/master/internal/helpers/statuscode.go


also https://github.com/labstack/echo/pull/2892',
        pull_request_url = 'https://github.com/labstack/echo/pull/2901'
    where id = _change_id;

    -- Echo PR pair 37: #2899 creates the Change, #2898 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Middleware'),
        'add Go 1.26 to CI flow',
        'https://go.dev/doc/go1.26

+ update security.md'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'After `http.Server.Serve` returns we need to wait for graceful shutdown goroutine to finish because when application exits immediately there are no graceful shutdown.

Fixes: https://github.com/labstack/echo/issues/2897',
        pull_request_url = 'https://github.com/labstack/echo/pull/2899'
    where id = _change_id;

    -- Echo PR pair 38: #2896 creates the Change, #2894 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Binder'),
        'Update location of oapi-codegen in README',
        'Closed Echo pull request #2896 did not include a body.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'I though I already merged this.  I think  https://github.com/labstack/echo/pull/2876 got closed when I purged all old branches at my fork. I should not have deleted that branch as it was not merged yet

-------------

In case CSRF flow is using `Sec-Fetch-Site` header but there is form still wanting to render CSRF token fields into form we  can help them by setting dummy value to context that atleast something can be rendered into form. As we already know that this browser is able to send `Sec-Fetch-Site` header, we do not need to generate token value or deal with cookies.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2896'
    where id = _change_id;

    -- Echo PR pair 39: #2893 creates the Change, #2892 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Documentation'),
        'Fix CSRF middleware not setting token when Sec-Fetch-Site passes',
        '## Summary

Fixes #2874

When `checkSecFetchSiteRequest()` returns `(true, nil)` — e.g. for direct URL navigation where `Sec-Fetch-Site: none`, or same-origin requests — the CSRF middleware immediately calls `return next(c)` **without**:

1. Generating or retrieving the CSRF token
2. Setting the CSRF cookie
3. Storing the token in context via `c.Set(config.ContextKey, token)`
4. Adding the `Vary: Cookie` response header

This breaks all server-rendered forms that use `c.Get("csrf")` to embed a CSRF token in HTML for subsequent POST requests. The bug is triggered by every modern browser during direct navigation (typing a URL, clicking a bookmark, or following an external link), since browsers automatically send `Sec-Fetch-Site: none` in these scenarios.

## Changes

- **`middleware/csrf.go`**: Extract token generation and cookie/context setting into a `setTokenInContext` helper closure, and call it in the `Sec-Fetch-Site` allow path before calling `next(c)`. The existing legacy token path remains unchanged.
- **`middleware/csrf_test.go`**:
  - Add `expectCookieContains` assertion to the existing `SecFetchSite=same-origin` test case
  - Add two new table-driven test cases verifying cookie is set for `Sec-Fetch-Site: none` and `same-origin` GET requests
  - Add a dedicated `TestCSRF_SecFetchSiteSetsTokenInContext` regression test that verifies the token is accessible in context, the cookie is set, the `Vary` header is correct, and existing cookie tokens are reused

## Test plan

- [x] All existing CSRF tests pass (`go test ./middleware/ -run TestCSRF -v`)
- [x] New regression tests pass for the exact scenario from the issue
- [x] Full test suite passes with race detector (`go test -race ./...`)
- [x] Verified the fix handles both new token generation and reuse of existing cookie tokens

🤖 Generated with [Claude Code](https://claude.com/claude-code)'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## Overview
Implemented the `Is` method on the `HTTPError` struct, enabling error checking using `errors.Is` (particularly for comparing with sentinel errors).

## Background
Starting with Go 1.13, `errors.Is` is recommended for error checking, but Echo''s `HTTPError` did not previously implement the `Is` method. Echo''s predefined errors (such as `echo.ErrNotFound`) are of the internal `httpError` type, while errors created with `NewHTTPError` are of type `*HTTPError`. Because these are distinct types in Go''s type system, `errors.Is(err, echo.ErrNotFound)` would return false if `err` was of type `*HTTPError`, making intended error handling difficult. This change resolves the issue by adding logic that considers errors with matching status codes to be the same error.

## Changes
- `httperror.go`: Added an `Is` method to the `HTTPError` struct. When comparing against `*HTTPError` or `*httpError`, it compares the status code.
- `httperror_test.go`: Added a test case for the `Is` method (`TestHTTPError_Is`).',
        pull_request_url = 'https://github.com/labstack/echo/pull/2893'
    where id = _change_id;

    -- Echo PR pair 40: #2891 creates the Change, #2889 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Maintenance'),
        'Fix staticmw',
        'Fix directory traversal vulnerability under Windows in Static middleware when default Echo filesystem is used. Reported by @shblue21.

This applies to cases when:
- Windows is used as OS
- `middleware.StaticConfig.Filesystem` is `nil` (default)
- `echo.Filesystem` is has not been set explicitly (default)'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Closed Echo pull request #2889 did not include a body.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2891'
    where id = _change_id;

    -- Echo PR pair 41: #2888 creates the Change, #2887 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Router'),
        'Changelog for version 5.0.2',
        'Closed Echo pull request #2888 did not include a body.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'fix Static middleware listing all files from given filesystem root when browser=true

fixes: https://github.com/labstack/echo/issues/2886',
        pull_request_url = 'https://github.com/labstack/echo/pull/2888'
    where id = _change_id;

    -- Echo PR pair 42: #2885 creates the Change, #2881 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Middleware'),
        'Fill c.Request().Pattern field with route path to help standard library based middlewares',
        'Fill c.Request().Pattern field with route path to help standard library based middlewares.  For example Otel standard library  Request/Response field extration uses `Request.Pattern` as route.

Relates to https://github.com/labstack/echo-contrib/pull/141'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'The repository lacked proper documentation for security vulnerability reporting. The existing SECURITY.md only stated "look for maintainers email(s) in commits and email them."

## Changes

- **Primary reporting method**: Added instructions for GitHub Private Vulnerability Reporting with direct link to Security tab
- **Fallback contact**: Listed all current maintainers with GitHub profile links for email-based reporting
- **Reporter guidance**: Added sections for what to include in reports and response time expectations (48h acknowledgment, 7d detailed response)
- **Security process**: Documented vulnerability handling workflow
- **Version support**: Fixed table formatting and clarified supported versions (`>= 4.15.x` vs previous ambiguous `> 4.15.x`)

The updated policy provides a professional disclosure pathway while GitHub Private Vulnerability Reporting awaits admin enablement in repository settings.

<!-- START COPILOT ORIGINAL PROMPT -->



<details>

<summary>Original prompt</summary>

>
> ----
>
> *This section details on the original issue you should resolve*
>
> <issue_title>Enable GitHub Private Vulnerability Reporting</issue_title>
> <issue_description>Could you please enable GitHub Private Vulnerability Reporting for this repository?
> https://docs.github.com/en/code-security/how-tos/report-and-fix-vulnerabilities/configure-vulnerability-reporting/configuring-private-vulnerability-reporting-for-a-repository
>
> This will allow users and contributors to open private issues or advisories regarding security fixes. These issues are private to maintainers and reporters until published. There is currently not a SECURITY.md or security contact present in the documentation.
>
> </issue_description>
>
> ## Comments on the Issue (you are @copilot in this section)
>
> <comments>
> <comment_new><author>@aldas</author><body>
> @vishr , I do not see these [options](https://docs.github.com/en/code-security/how-tos/report-and-fix-vulnerabilities/configure-vulnerability-reporting/configuring-private-vulnerability-reporting-for-a-repository) under settings, so probably only you can enable this.</body></comment_new>
> <comment_new><author>@aldas</author><body>
> @wodzen  I created the most basic security.md that can be. At the moment if you have something, look for my or @vishr email. For example in verified commits and email us.</body></comment_new>
> </comments>
>


</details>



<!-- START COPILOT CODING AGENT SUFFIX -->

- Fixes labstack/echo#2879

<!-- START COPILOT CODING AGENT TIPS -->
---

✨ Let Copilot coding agent [set things up for you](https://github.com/labstack/echo/issues/new?title=✨+Set+up+Copilot+instructions&body=Configure%20instructions%20for%20this%20repository%20as%20documented%20in%20%5BBest%20practices%20for%20Copilot%20coding%20agent%20in%20your%20repository%5D%28https://gh.io/copilot-coding-agent-tips%29%2E%0A%0A%3COnboard%20this%20repo%3E&assignees=copilot) — coding agent works faster and does higher quality work when set up for your repo.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2885'
    where id = _change_id;

    -- Echo PR pair 43: #2880 creates the Change, #2878 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Binder'),
        'Changelog for v5.0.1 release',
        'Closed Echo pull request #2880 did not include a body.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Hi maintainers,
Just a quick doc fix about the DenyHandler provided example.

Re-opened from #2864',
        pull_request_url = 'https://github.com/labstack/echo/pull/2880'
    where id = _change_id;

    -- Echo PR pair 44: #2877 creates the Change, #2876 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Documentation'),
        'Context: json should not send status code before serialization is complete',
        'Context: json should not send status code before serialization is complete

Relates to https://github.com/labstack/echo/pull/2866#issuecomment-3778036694'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'CSRF: support older token-based CSRF protection handler that want torender token into template',
        pull_request_url = 'https://github.com/labstack/echo/pull/2877'
    where id = _change_id;

    -- Echo PR pair 45: #2875 creates the Change, #2871 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Maintenance'),
        'fix: validate Connection header in IsWebSocket() [per RFC 6455]',
        '## Description
This PR fixes the `IsWebSocket()` function to properly validate WebSocket upgrade requests according to RFC 6455 specification.

## Problem
The current implementation only checks the `Upgrade` header but ignores the `Connection` header requirement specified in RFC 6455 Section 1.3. A valid WebSocket upgrade request must have both headers present with specific values.

## Solution
Updated `IsWebSocket()` to validate both required headers:
- `Upgrade: websocket` (case-insensitive)
- `Connection: upgrade` (case-insensitive, may contain other values)

## Changes
- Modified `IsWebSocket()` in `context.go` to check both headers
- Updated `TestContext_IsWebSocket()` in `context_test.go` with additional test cases
- Added test case for missing/invalid `Connection` header

## Testing
All existing tests pass, including new test cases that verify:
- Valid WebSocket requests with both headers present
- Case-insensitive header matching
- Invalid requests missing the `Connection` header
- Invalid requests with `Connection: close`

## RFC Reference
[RFC 6455 - The WebSocket Protocol](https://tools.ietf.org/html/rfc6455#section-1.3)'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Panic middleware: will now return a custom PanicStackError with stack trace when config.DisablePrintStack is set to false.

relates to https://github.com/labstack/echo/issues/2869#issuecomment-3771782789',
        pull_request_url = 'https://github.com/labstack/echo/pull/2875'
    where id = _change_id;

    -- Echo PR pair 46: #2868 creates the Change, #2866 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Router'),
        'Merge V5 to master branch',
        'See https://github.com/labstack/echo/discussions/2861

If you are using Linux you can migrate easier parts like that:
```bash
find . -type f -name "*.go" -exec sed -i ''s/ echo.Context/ *echo.Context/g'' {} +
find . -type f -name "*.go" -exec sed -i ''s/echo\/v4/echo\/v5/g'' {} +
```
or in your favorite IDE

Replace all:
1. ` echo.Context` -> ` *echo.Context`
2. `echo/v4` -> `echo/v5`'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '##  Bug Fix

### Problem
The `json()` method in `context.go` (line 504) was inconsistent with other response methods in how it sets the HTTP status code.

**Current behavior:**
```go
func (c *context) json(code int, i any, indent string) error {
    c.writeContentType(MIMEApplicationJSON)
    c.response.Status = code  // ❌ Directly setting Status field
    return c.echo.JSONSerializer.Serialize(c, i, indent)
}
```

This approach directly sets the `Status` field instead of properly calling `WriteHeader()`, which bypasses header commitment and prevents warnings about header modifications after the status is set.

### Solution
Updated the method to use `c.response.WriteHeader(code)` for consistency with other response methods:

```go
func (c *context) json(code int, i any, indent string) error {
    c.writeContentType(MIMEApplicationJSON)
    c.response.WriteHeader(code)  // ✅ Properly calls WriteHeader
    return c. echo.JSONSerializer.Serialize(c, i, indent)
}
```

All other similar response methods already use `WriteHeader()`:
- `jsonPBlob()` - line 489:  `c.response.WriteHeader(code)`
- `xml()` - line 543: `c.response.WriteHeader(code)`
- `Blob()` - line 578: `c.response.WriteHeader(code)`
- `JSONPBlob()` - line 530: `c.response.WriteHeader(code)`',
        pull_request_url = 'https://github.com/labstack/echo/pull/2868'
    where id = _change_id;

    -- Echo PR pair 47: #2864 creates the Change, #2860 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Middleware'),
        'docs: add missing err parameter to DenyHandler example',
        'Hi maintainers,
Just a quick doc fix about the DenyHandler provided example.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '# Changelog

## v4.15.0 - 2026-01-01


**Security**

NB: **If your application relies on cross-origin or same-site (same subdomain) requests do not blindly push this version to production**


The CSRF middleware now supports the [**Sec-Fetch-Site**](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Sec-Fetch-Site) header as a modern, defense-in-depth approach to [CSRF
protection](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html#fetch-metadata-headers), implementing the OWASP-recommended Fetch Metadata API alongside the traditional token-based mechanism.

**How it works:**

Modern browsers automatically send the `Sec-Fetch-Site` header with all requests, indicating the relationship
between the request origin and the target. The middleware uses this to make security decisions:

- **`same-origin`** or **`none`**: Requests are allowed (exact origin match or direct user navigation)
- **`same-site`**: Falls back to token validation (e.g., subdomain to main domain)
- **`cross-site`**: Blocked by default with 403 error for unsafe methods (POST, PUT, DELETE, PATCH)

For browsers that don''t send this header (older browsers), the middleware seamlessly falls back to
traditional token-based CSRF protection.

**New Configuration Options:**
- `TrustedOrigins []string`: Allowlist specific origins for cross-site requests (useful for OAuth callbacks, webhooks)
- `AllowSecFetchSiteFunc func(echo.Context) (bool, error)`: Custom logic for same-site/cross-site request validation

**Example:**
  ```go
  e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
      // Allow OAuth callbacks from trusted provider
      TrustedOrigins: []string{"https://oauth-provider.com"},

      // Custom validation for same-site requests
      AllowSecFetchSiteFunc: func(c echo.Context) (bool, error) {
          // Your custom authorization logic here
          return validateCustomAuth(c), nil
          // return true, err  // blocks request with error
          // return true, nil  // allows CSRF request through
          // return false, nil // falls back to legacy token logic
      },
  }))
  ```
PR: https://github.com/labstack/echo/pull/2858

**Type-Safe Generic Parameter Binding**

* Added generic functions for type-safe parameter extraction and context access by @aldas in https://github.com/labstack/echo/pull/2856

  Echo now provides generic functions for extracting path, query, and form parameters with automatic type conversion,
  eliminating manual string parsing and type assertions.

  **New Functions:**
  - Path parameters: `PathParam[T]`, `PathParamOr[T]`
  - Query parameters: `QueryParam[T]`, `QueryParamOr[T]`, `QueryParams[T]`, `QueryParamsOr[T]`
  - Form values: `FormParam[T]`, `FormParamOr[T]`, `FormParams[T]`, `FormParamsOr[T]`
  - Context store: `ContextGet[T]`, `ContextGetOr[T]`

  **Supported Types:**
  Primitives (`bool`, `string`, `int`/`uint` variants, `float32`/`float64`), `time.Duration`, `time.Time`
  (with custom layouts and Unix timestamp support), and custom types implementing `BindUnmarshaler`,
  `TextUnmarshaler`, or `JSONUnmarshaler`.

  **Example:**
  ```go
  // Before: Manual parsing
  idStr := c.Param("id")
  id, err := strconv.Atoi(idStr)

  // After: Type-safe with automatic parsing
  id, err := echo.PathParam[int](c, "id")

  // With default values
  page, err := echo.QueryParamOr[int](c, "page", 1)
  limit, err := echo.QueryParamOr[int](c, "limit", 20)

  // Type-safe context access (no more panics from type assertions)
  user, err := echo.ContextGet[*User](c, "user")
  ```

PR: https://github.com/labstack/echo/pull/2856



**DEPRECATION NOTICE** Timeout Middleware Deprecated - Use ContextTimeout Instead

The `middleware.Timeout` middleware has been **deprecated** due to fundamental architectural issues that cause
data races. Use `middleware.ContextTimeout` or `middleware.ContextTimeoutWithConfig` instead.

**Why is this being deprecated?**

The Timeout middleware manipulates response writers across goroutine boundaries, which causes data races that
cannot be reliably fixed without a complete architectural redesign. The middleware:

- Swaps the response writer using `http.TimeoutHandler`
- Must be the first middleware in the chain (fragile constraint)
- Can cause races with other middleware (Logger, metrics, custom middleware)
- Has been the source of multiple race condition fixes over the years

**What should you use instead?**

The `ContextTimeout` middleware (available since v4.12.0) provides timeout functionality using Go''s standard
context mechanism. It is:

- Race-free by design
- Can be placed anywhere in the middleware chain
- Simpler and more maintainable
- Compatible with all other middleware

**Migration Guide:**

```go
// Before (deprecated):
e.Use(middleware.Timeout())

// After (recommended):
e.Use(middleware.ContextTimeout(30 * time.Second))
```

**Important Behavioral Differences:**

1. **Handler cooperation required**: With ContextTimeout, your handlers must check `context.Done()` for cooperative
   cancellation. The old Timeout middleware would send a 503 response regardless of handler cooperation, but had
   data race issues.

2. **Error handling**: ContextTimeout returns errors through the standard error handling flow. Handlers that receive
   `context.DeadlineExceeded` should handle it appropriately:

```go
e.GET("/long-task", func(c echo.Context) error {
    ctx := c.Request().Context()

    // Example: database query with context
    result, err := db.QueryContext(ctx, "SELECT * FROM large_table")
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            // Handle timeout
            return echo.NewHTTPError(http.StatusServiceUnavailable, "Request timeout")
        }
        return err
    }

    return c.JSON(http.StatusOK, result)
})
```

3. **Background tasks**: For long-running background tasks, use goroutines with context:

```go
e.GET("/async-task", func(c echo.Context) error {
    ctx := c.Request().Context()

    resultCh := make(chan Result, 1)
    errCh := make(chan error, 1)

    go func() {
        result, err := performLongTask(ctx)
        if err != nil {
            errCh <- err
            return
        }
        resultCh <- result
    }()

    select {
    case result := <-resultCh:
        return c.JSON(http.StatusOK, result)
    case err := <-errCh:
        return err
    case <-ctx.Done():
        return echo.NewHTTPError(http.StatusServiceUnavailable, "Request timeout")
    }
})
```

**Enhancements**

* Fixes by @aldas in https://github.com/labstack/echo/pull/2852
* Generic functions by @aldas in https://github.com/labstack/echo/pull/2856
* CRSF with Sec-Fetch-Site checks by @aldas in https://github.com/labstack/echo/pull/2858',
        pull_request_url = 'https://github.com/labstack/echo/pull/2864'
    where id = _change_id;

    -- Echo PR pair 48: #2859 creates the Change, #2858 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Binder'),
        'fix: RouteNotFound handler does not falls back to root one',
        'Fixes #2485

## Changes
- Store RouteNotFound handler for "/*" path on Echo instance for fallback use
- Modify group middleware to check for root RouteNotFound handler before using default
- Add tests verifying root 404 handler fallback behavior for groups with middleware'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'From: https://github.com/labstack/echo/issues/2855

Note to self: [Hackernews](https://news.ycombinator.com/item?id=46351666) had this blog post posted.

* https://blog.miguelgrinberg.com/post/csrf-protection-without-tokens-or-hidden-form-fields
* https://words.filippo.io/csrf/
* https://github.com/rails/rails/pull/56350

see https://github.com/golang/go/blob/master/src/net/http/csrf.go  which was introduced in GO 1.25',
        pull_request_url = 'https://github.com/labstack/echo/pull/2859'
    where id = _change_id;

    -- Echo PR pair 49: #2856 creates the Change, #2852 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Documentation'),
        'Generic functions',
        'Add generic functions to get typed PathParam/QueryParam/FormParam values. Also *Or variants for default values.
Note: those who have forgotten - structs can not have generic methods. only generic functions are allowed.
Types that are supported:
  - bool
  - float32
  - float64
  - int
  - int8
  - int16
  - int32
  - int64
  - uint
  - uint8/byte
  - uint16
  - uint32
  - uint64
  - string
  - echo.BindUnmarshaler interface
  - encoding.TextUnmarshaler interface
  - json.Unmarshaler interface
  - time.Duration
  - time.Time use echo.TimeOpts or echo.TimeLayout to set time parsing configuration

Scalar values:
* `id, err := echo.PathParam[int](c, "id")`
* `id, err := echo.QueryParam[int](c, "id")`
* `id, err := echo.FormParam[int](c, "id")`
* `id, err := echo.PathParamOr[int](c, "id", -1)`
* `id, err := echo.QueryParamOr[int](c, "id", -1)`
* `id, err := echo.FormParamOr[int](c, "id", -1)`

For getting slices:
* `ids, err := echo.QueryParams[int](c, "id")`   returns `[]int`
* `ids, err := echo.FormParams[int](c, "id")`

Generic parse functions:
* `id, err := echo.ParseValue[int]("123")`
* `id, err := echo.ParseValueOr[int]("123",-1)`
* `ids, err := echo.ParseValues[int]([]string{"123", "124"})`
* `ids, err := echo.ParseValuesOr[int]([]string{"123", "124"}, []int{666})`

Example
```go
	e.GET(`/user/:id`, func(c echo.Context) error {
		id, err := echo.PathParam[int](c, "id")
		if err == nil {
			return err
		}
		return c.String(http.StatusOK, fmt.Sprintf("Hello %v\n", id))
	})
```

-----------------------

For `Context.Get()` generic versions

* `v, err := ContextGetOr[int64](c, "key", 999)`
* `v, err := ContextGet[int64](c, "key")`

Example with JWT middleware:
```go
	e.Use(echojwt.JWT([]byte("secret")))

	e.GET(`/user/:id`, func(c echo.Context) error {
		token, err := echo.ContextGet[*jwt.Token](c, "user")
		if err == nil {
			return echo.ErrUnauthorized.WithInternal(err)
		}
		return c.String(http.StatusOK, fmt.Sprintf("Hello %v\n", token.Claims))
	})
```'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '* Mark timeout middleware as deprecated
* fixes/improvements for things that Claude and Codex analysis found',
        pull_request_url = 'https://github.com/labstack/echo/pull/2856'
    where id = _change_id;

    -- Echo PR pair 50: #2851 creates the Change, #2850 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Maintenance'),
        'Changelog for 4.14.0',
        'changelog and version number bump'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Closed Echo pull request #2850 did not include a body.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2851'
    where id = _change_id;

    -- Echo PR pair 51: #2849 creates the Change, #2843 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Router'),
        'Logger middleware json string escaping and deprecation',
        '**Security**

* Logger middleware: escape string values when logger format looks like JSON. See  #2846



**Enhancements**

* Add `middleware.RequestLogger` function to replace `middleware.Logger`. `middleware.RequestLogger` uses default slog logger.
  Default slog logger output can be configured to JSON format like that:
  ```go
  slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
  e.Use(middleware.RequestLogger())
  ```
* Deprecate `middleware.Logger` function and point users to `middleware.RequestLogger` and `middleware.RequestLoggerWithConfig`'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'dependabot complains that [golang.org/x/crypto](http://golang.org/x/crypto) need upgrading

Altough we do not used SSH package we still are "marked" as affected:
* https://github.com/advisories/GHSA-f6x5-jh6r-wrfv
* https://github.com/advisories/GHSA-j5w8-q4qc-rx2x',
        pull_request_url = 'https://github.com/labstack/echo/pull/2849'
    where id = _change_id;

    -- Echo PR pair 52: #2842 creates the Change, #2838 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Middleware'),
        'Add Echo MCP tool',
        'Hi!

I added a reference to the echo-mcp tool.

I created this lib to automatically convert any Echo API to an MCP Tool that can be called by any agent.

The setup is really simple:

```go
    e := echo.New()

    // Existing API routes
    e.GET("/ping", func(c echo.Context) error {
        return c.JSON(http.StatusOK, map[string]string{"message": "pong"})
    })


    // Add MCP support
    mcp := server.New(e)
    mcp.Mount("/mcp")

    e.Start(":8080")
```'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '(#2837)
Ensure proxy connection is closed in proxy middleware `proxyRaw` function',
        pull_request_url = 'https://github.com/labstack/echo/pull/2842'
    where id = _change_id;

    -- Echo PR pair 53: #2835 creates the Change, #2833 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Binder'),
        'Replace custom private IP range check with built-in net.IP.IsPrivate',
        'Replace `isPrivateIPRange` with built-in method `net.IP.IsPrivate`'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Correct the fixture path used in `group_test.go`.

### Change

- Remove the redundant slash in the file path used in `group_test.go`',
        pull_request_url = 'https://github.com/labstack/echo/pull/2835'
    where id = _change_id;

    -- Echo PR pair 54: #2832 creates the Change, #2829 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Documentation'),
        'Fix duplicate plus operator in router backtracking logic',
        '# Summary

This PR fixes a typo where searchIndex += +len(search) was incorrectly written with
duplicate plus operators. The correct expression should be searchIndex += len(search).

# Description

## Problem

In router.go at line 695, there is a syntax error with duplicate plus operators:

searchIndex += +len(search)

While Go interprets this as searchIndex = searchIndex + (+len(search)) (treating the
second + as a unary plus operator), it''s clearly a typo and not the intended expression.

## Solution

Changed the expression to the correct form:

searchIndex += len(search)

## Impact

• Functionality: The behavior remains the same since Go''s unary plus operator doesn''t
change the value, but the code is now clearer and matches the intended logic.
• Readability: Improves code clarity by removing the confusing duplicate operator.

## Testing

• Existing tests should continue to pass as the functional behavior is unchanged
• The fix is in the router''s backtracking logic for "any" node matching

## Location

• File: router.go
• Line: 695
• Function: Router.Find()'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## Summary

Complete redesign of the README with a modern, professional, and visually appealing layout that positions Echo as the premium choice for Go web development.

## ✨ Key Improvements

### **Visual & Design**
- 🎨 Modern layout with emojis, icons, and professional styling
- 🏢 Centered hero section with logo and clear value proposition
- 📊 Feature grid layout for better organization
- 🏗️ ASCII architecture diagram
- 📈 Performance benchmarks table
- 🔥 GitHub badges and project statistics

### **Content Enhancement**
- 🎯 **"Why Echo?"** section highlighting key differentiators
- 🚀 **Quick Start** with practical, copy-paste examples
- 🌟 **Feature showcase** organized in logical categories
- 📦 **Ecosystem** section with official and community packages
- 🎓 **Learning Resources** for developers at all levels
- 🏢 **Trust signals** showing companies using Echo

### **Developer Experience**
- ⚡ **60-second quick start** with complete working example
- 🛠️ **Enhanced contribution guidelines** with clear steps
- 🎯 **Roadmap** showing future features and vision
- 🤝 **Community** focus with discussion links
- 🆚 **Framework comparison** table

### **Professional Positioning**
- 📊 **Performance benchmarks** demonstrating superiority
- 🔒 **Production-ready** messaging with security focus
- 🌍 **Enterprise adoption** highlighting major users
- 📈 **Project metrics** showing healthy ecosystem

## 🎯 Impact

This README transforms Echo''s first impression from a simple framework description to a **premium, enterprise-ready solution** that developers and organizations can trust for critical applications.

### Before vs After:
- ❌ **Before**: Basic feature list, minimal examples
- ✅ **After**: Compelling value proposition, comprehensive showcase, professional presentation

## 🧪 Testing

- [x] All markdown renders correctly
- [x] All links are valid and functional
- [x] Images and badges display properly
- [x] No breaking changes to existing functionality
- [x] Maintains all original information while enhancing presentation

This positions Echo competitively against other frameworks and provides a compelling case for adoption.

🤖 Generated with [Claude Code](https://claude.ai/code)',
        pull_request_url = 'https://github.com/labstack/echo/pull/2832'
    where id = _change_id;

    -- Echo PR pair 55: #2828 creates the Change, #2827 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Maintenance'),
        'Fix typo in SetParamValues comment',
        '## Summary

Fixes a simple typo in the SetParamValues method documentation.

**Change:**
- Line 362: `brake` → `break` in Router#Find code comment

**Benefits:**
- Correct English spelling
- Better code documentation clarity

## Test plan

- [x] No functional changes
- [x] Tests pass

🤖 Generated with [Claude Code](https://claude.ai/code)'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## Summary

Fixes a simple typo in the ContextTimeout middleware documentation.

**Change:**
- Line 19: `aries` → `arises` in ErrorHandler comment

**Benefits:**
- Correct English grammar
- Better code documentation

## Test plan

- [x] No functional changes
- [x] Tests pass

🤖 Generated with [Claude Code](https://claude.ai/code)',
        pull_request_url = 'https://github.com/labstack/echo/pull/2828'
    where id = _change_id;

    -- Echo PR pair 56: #2826 creates the Change, #2825 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Router'),
        'Improve secure middleware readability and add deprecation notice',
        '## Summary

Improves code readability and maintainability of the secure middleware with better user guidance.

**Changes:**
1. **Refactor HSTS header construction** - Replace nested `fmt.Sprintf` with slice building and `strings.Join` for clearer logic
2. **Add X-XSS-Protection deprecation notice** - Document that CSP is recommended over the deprecated header
3. **Clean up imports** - Remove unused `fmt` import

**Benefits:**
- Cleaner, more maintainable HSTS directive building
- Better user guidance about modern security practices
- Improved code readability

## Test plan

- [x] All existing tests pass
- [x] Linting passes
- [x] No behavioral changes to security headers

Fixes #2799

🤖 Generated with [Claude Code](https://claude.ai/code)'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## Summary

Modernizes the BasicAuth middleware with improved code readability and RFC compliance.

**Changes:**
1. **Use `strings.Cut` for credential parsing** - Replaces manual for loop with Go 1.18+ `strings.Cut`
2. **Fix RFC 7617 compliance** - Always quote realm parameter as required by RFC

**Benefits:**
- Cleaner, more readable code using modern Go idioms
- Proper RFC 7617 compliance for WWW-Authenticate header
- Reduced code complexity (fewer lines, simpler logic)

## Test plan

- [x] All existing tests pass
- [x] Linting passes
- [x] No behavioral changes to authentication logic

Fixes #2794

🤖 Generated with [Claude Code](https://claude.ai/code)',
        pull_request_url = 'https://github.com/labstack/echo/pull/2826'
    where id = _change_id;

    -- Echo PR pair 57: #2824 creates the Change, #2823 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Middleware'),
        'Modernize remaining interface{} to any in context.go',
        '## Summary

Completes the modernization of `context.go` by replacing all remaining `interface{}` types with `any`.

**Changes:**
- Updated Context interface method signatures (23 occurrences)
- Updated implementation method signatures
- Methods affected: Get, Set, Bind, Validate, Render, JSON*, XML*, JSONP

**Benefits:**
- Improved code readability
- Follows Go 1.18+ best practices
- Consistent with modern Go codebases

## Test plan

- [x] All existing tests pass
- [x] Linting passes
- [x] No behavioral changes

🤖 Generated with [Claude Code](https://claude.ai/code)'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## Summary

Modernizes a for loop in `context.go` to use Go 1.22''s new range over int syntax for cleaner iteration.

**Changes:**
- Replace `for i := 0; i < limit; i++` with `for i := range limit` in `SetParamValues` method

**Benefits:**
- Cleaner, more idiomatic Go 1.22+ code
- Slight performance improvement
- Reduced cognitive load

## Test plan

- [x] All existing tests pass
- [x] Linting passes
- [x] No behavioral changes

🤖 Generated with [Claude Code](https://claude.ai/code)',
        pull_request_url = 'https://github.com/labstack/echo/pull/2824'
    where id = _change_id;

    -- Echo PR pair 58: #2822 creates the Change, #2821 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Binder'),
        'Modernize context.go by replacing interface{} with any',
        '## Summary

Modernizes the Echo Context interface by replacing all instances of `interface{}` with the more readable `any` type alias introduced in Go 1.18.

## Changes

**23 Modernizations across Context interface methods:**
- `Get(key string) interface{}` → `Get(key string) any`
- `Set(key string, val interface{})` → `Set(key string, val any)`
- `Bind(i interface{})` → `Bind(i any)`
- `Validate(i interface{})` → `Validate(i any)`
- `Render(code int, name string, data interface{})` → `Render(code int, name string, data any)`
- `JSON(code int, i interface{})` → `JSON(code int, i any)`
- `JSONP(code int, callback string, i interface{})` → `JSONP(code int, callback string, i any)`
- `XML(code int, i interface{})` → `XML(code int, i any)`
- `Blob(code int, contentType string, b []byte)` → (internal any usage)
- `Stream(code int, contentType string, r io.Reader)` → (internal any usage)
- Plus all other Context interface methods with `interface{}` parameters

## Benefits

### 🚀 **Modernization**
- Aligns with Go 1.18+ best practices and conventions
- Makes the API more approachable for developers familiar with modern Go
- Improves code readability and reduces cognitive load

### 📖 **Developer Experience**
- `any` is more intuitive and self-documenting than `interface{}`
- Easier to read in IDE tooltips and documentation
- Follows patterns used by modern Go libraries

### 🔒 **Safety & Compatibility**
- **Zero breaking changes** - `any` is just an alias for `interface{}`
- **100% backward compatible** - all existing code continues to work
- **Identical runtime behavior** - no performance or functional differences

## Testing

- ✅ **Compilation**: Code builds successfully
- ✅ **Tests**: Context tests pass without issues
- ✅ **Compatibility**: No API changes, only type alias substitution
- ✅ **Linting**: Addresses modernization suggestions from static analysis

## Type of Change

- 🎨 **Code modernization** - Updates to current Go standards
- 📚 **API clarity** - Improves readability without functional changes
- 🔧 **Developer experience** - Makes interfaces more approachable

## Impact

This change modernizes Echo''s public API to follow current Go conventions while maintaining perfect backward compatibility. Developers using Echo will benefit from:

- More readable method signatures
- Consistency with modern Go codebases
- Better IDE experience with cleaner type information

---

*This is a pure modernization change with zero risk - `any` and `interface{}` are functionally identical.*'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## Summary

Addresses issue #2382 by correcting the misleading comment on `Context.Bind` that did not accurately describe the actual binding behavior.

## Problem

The comment on `Context.Bind` in `context.go` was incomplete and confusing:

**Previous comment:**
```go
// Bind binds path params, query params and the request body into provided type `i`. The default binder
// binds body based on Content-Type header.
```

**Issues with the old comment:**
1. ❌ Didn''t explain the binding **order** (path → query → body)
2. ❌ Didn''t mention that later steps can **override** earlier values
3. ❌ Didn''t specify that query params are only bound for **GET/DELETE/HEAD**
4. ❌ Didn''t reference **single-source binding methods** for more control

## Solution

**New accurate comment:**
```go
// Bind binds data from multiple sources to the provided type `i` in this order:
// 1) path parameters, 2) query parameters (for GET/DELETE/HEAD only), 3) request body.
// Each step can override values from the previous step. For single source binding use
// BindPathParams, BindQueryParams, BindHeaders, or BindBody directly.
// The default binder handles body based on Content-Type header.
```

**Improvements:**
- ✅ **Clear binding order**: Explicitly states the 1-2-3 sequence
- ✅ **Override behavior**: Warns that later steps can override earlier values
- ✅ **HTTP method specifics**: Notes query param binding only for GET/DELETE/HEAD
- ✅ **Alternative methods**: References single-source binding methods
- ✅ **Content-Type info**: Preserves useful body binding information

## Impact

This fix prevents confusion for developers who might expect different behavior from `Context.Bind()` based on the previous misleading documentation. Now the comment accurately reflects the actual implementation in `bind.go`.

## Type of Change

- 📚 **Documentation fix** - No code logic changes
- 🔧 **Comment improvement** - Better developer experience
- 🎯 **Issue resolution** - Directly addresses reported confusion

## Testing

- ✅ Code compiles without issues
- ✅ No functional changes - documentation only
- ✅ Comment aligns with actual `DefaultBinder.Bind` implementation

Fixes #2382

---

*This is a simple documentation improvement that enhances clarity for Echo developers using the binding functionality.*',
        pull_request_url = 'https://github.com/labstack/echo/pull/2822'
    where id = _change_id;

    -- Echo PR pair 59: #2819 creates the Change, #2818 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Documentation'),
        'Document ContextTimeout middleware with comprehensive examples',
        '## Summary

Addresses issue #2745 by providing comprehensive documentation for the ContextTimeout middleware, which was completely undocumented despite being the recommended approach for handling request timeouts in Echo.

## Problem Solved

Users were confused because:
- ContextTimeout middleware exists but has **zero documentation**
- Website only documents the **deprecated Timeout middleware** (with warnings against it)
- No examples showing how to properly implement handlers that work with ContextTimeout
- No explanation of why ContextTimeout is preferred over Timeout

## Changes

### 📚 **Comprehensive Configuration Documentation**

**Overview & Key Differences:**
- Clear explanation of ContextTimeout vs deprecated Timeout middleware
- Safety benefits: no response writer interference, no data races
- Cooperative cancellation model explanation

**Configuration Examples (3 scenarios):**
- Basic usage: `middleware.ContextTimeout(30 * time.Second)`
- Custom error handling with JSON responses
- Route-specific timeout skipping for health checks

### 🛠️ **Practical Handler Examples**

**3 Detailed Real-World Scenarios:**
1. **Database Operations**: Context-aware SQL queries with proper error handling
2. **Long-Running Processing**: Goroutine-based operations with select statements
3. **HTTP Proxy/Client**: Outbound requests with context propagation

### 📖 **Best Practices & Patterns**

**Common Integration Patterns:**
- Database: `db.QueryContext(ctx, query, args...)`
- HTTP Client: `http.NewRequestWithContext(ctx, method, url, body)`
- Redis: `redisClient.Get(ctx, key)`
- CPU-intensive loops with `ctx.Done()` checking

**Practical Guidelines:**
- Recommended timeout values for different use cases
- How to handle context cancellation gracefully
- When and where to place the middleware

### 🔧 **Enhanced Field Documentation**

- **Skipper**: Examples for excluding health check endpoints
- **ErrorHandler**: Custom timeout response patterns with JSON
- **Timeout**: Recommended durations for APIs, uploads, background tasks

### 🚀 **Function Documentation**

- **ContextTimeout()**: Basic usage with handler requirements
- **ContextTimeoutWithConfig()**: Advanced configuration examples
- **ToMiddleware()**: Validation and error handling scenarios

## Impact

This documentation addresses the exact concerns raised in issue #2745:

1. ✅ **"Cannot find any mention of ContextTimeout middleware"** → Now has 280+ lines of comprehensive docs
2. ✅ **"Documentation only lists the recommended-against Timeout middleware"** → Clear explanation of why ContextTimeout is preferred
3. ✅ **"Usage example showing what the handler should look like"** → 3 detailed handler examples + common patterns

## Before/After

**Before:** Zero documentation, users confused about which timeout middleware to use
**After:** Enterprise-grade documentation with practical examples and best practices

## Testing

- ✅ All existing ContextTimeout tests pass
- ✅ Code compiles without issues
- ✅ Documentation follows Go conventions
- ✅ Examples are syntactically correct and functional

Fixes #2745

---

This PR complements PR #2818 (Logger middleware documentation) as part of ongoing efforts to improve Echo''s middleware documentation quality.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## Summary

Addresses issue #2665 by providing comprehensive documentation for the Logger middleware that was previously lacking detailed explanations and examples.

## Changes

**📚 Configuration Examples (8 different scenarios):**
- Basic usage with default settings
- Custom simple format
- JSON format with custom fields
- Custom time formatting
- Logging headers, query params, form data, and cookies
- File output configuration
- Custom tag functions with user logic
- Conditional logging with Skipper
- External logging service integration

**🏷️ Complete Tag Reference organized by category:**
- **Time Tags**: 7 different timestamp formats
- **Request Information**: 10 request-related tags
- **Response Information**: 6 response-related tags
- **Dynamic Tags**: 4 parameterized tag types with examples

**📖 Enhanced Field Documentation:**
- Clear purpose explanation for each LoggerConfig field
- Usage examples and best practices
- Default values and behavior
- Proper Go reference time format examples

**🔧 Troubleshooting Section:**
- 4 common issues with solutions
- Performance optimization tips
- Best practices for high-traffic applications

**🚀 Function Documentation:**
- Detailed explanation of default Logger() behavior with example JSON output
- Comprehensive LoggerWithConfig() documentation with usage examples

## Impact

This enhancement transforms the Logger middleware from having minimal documentation to having enterprise-grade documentation that:

- **Helps new users** quickly understand and configure the middleware
- **Provides advanced patterns** for experienced developers
- **Reduces support burden** by answering common questions upfront
- **Improves developer experience** with clear examples and troubleshooting

## Testing

- ✅ All existing tests pass
- ✅ Code compiles without issues
- ✅ Documentation follows Go documentation conventions
- ✅ Examples are syntactically correct and functional

Fixes #2665

## Before/After

**Before:** Basic tag list with minimal explanations
**After:** 200+ lines of comprehensive documentation with 8+ complete configuration examples

The issue specifically requested "detailed explanations of configuration options and comprehensive examples for various use cases" - this PR delivers exactly that.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2819'
    where id = _change_id;

    -- Echo PR pair 60: #2815 creates the Change, #2812 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Maintenance'),
        'router: return error when registering nil handler',
        'Currently, `router.insert` only logs a warning when a nil handler is registered,
and continues processing. This can lead to runtime panics due to unnoticed misconfiguration.

This change makes router registration fail fast by returning an error
when a nil handler is provided, so that misconfigurations can be detected earlier.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Closed Echo pull request #2812 did not include a body.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2815'
    where id = _change_id;

    -- Echo PR pair 61: #2810 creates the Change, #2807 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Router'),
        'Use Go 1.25 in CI',
        'https://tip.golang.org/doc/go1.25'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'update `golang.org/x/` libs to current versions',
        pull_request_url = 'https://github.com/labstack/echo/pull/2810'
    where id = _change_id;

    -- Echo PR pair 62: #2798 creates the Change, #2797 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Middleware'),
        'add optional IP anonymization',
        'This PR provides an option to anonymize IP:
* in context via `Context.AnonymizedIP()`
* in (old) logger middleware via `remote_ip_anon` token
* in (new) request logger middleware via `AnonymizeRemoteIP` option'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## Proposal
### 1. Error handling when reading request body:
Currently, errors in io.ReadAll(c.Request().Body) are ignored. If the read fails, subsequent handlers may receive an incomplete body, causing unexpected behavior. I believe that the robustness of the middleware can be improved by interrupting processing when an error occurs and returning an error.

```diff
import (
	"bufio"
	"bytes"
+++	"fmt"
	"errors"
	"io"
	"net"
// omission of sentence
			// Request
			reqBody := []byte{}
			if c.Request().Body != nil { // Read
---				reqBody, _ = io.ReadAll(c.Request().Body)
+++				var errRead error
+++				reqBody, errRead = io.ReadAll(c.Request().Body)
+++				if errRead != nil {
+++					return errRead
+++				}
			}
			c.Request().Body = io.NopCloser(bytes.NewBuffer(reqBody)) // Reset
```

### 2. Improvement of Flush panic messages:
Improve the messages when a panic occurs in the Flush method to be more detailed and consistent with the rest of the Echo framework (e.g., response.go). We believe this will make it clearer which ResponseWriters do not support the Flusher interface and make debugging easier.

```diff
import (
	"bufio"
	"bytes"
+++	"fmt"
	"errors"
	"io"
	"net"
// omission of sentence
func (w *bodyDumpResponseWriter) Flush() {
	err := http.NewResponseController(w.ResponseWriter).Flush()
	if err != nil && errors.Is(err, http.ErrNotSupported) {
		panic(errors.New("response writer flushing is not supported"))
		panic(fmt.Errorf("echo: response writer %T does not support flushing (http.Flusher interface)", w.ResponseWriter))
	}
}
```

### 3. Modification of tests to accommodate the above changes',
        pull_request_url = 'https://github.com/labstack/echo/pull/2798'
    where id = _change_id;

    -- Echo PR pair 63: #2795 creates the Change, #2793 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Binder'),
        'Improved code readability and RFC compliance',
        '## Proposal
### 1. Improved parsing of authentication information:
I currently use a for loop to split username and password, but we believe that using strings.Cut, available in Go 1.18 or later, would make the code simpler and more readable.

```diff
// basic_auth.go


				cred := string(b)
---				for i := 0; i < len(cred); i++ {
---					if cred[i] == '':'' {
---						// Verify credentials
---						valid, err := config.Validator(cred[:i], cred[i+1:], c)
---						if err != nil {
---							return err
---						} else if valid {
---							return next(c)
---						}
---						break
---					}
+++				user, pass, ok := strings.Cut(cred, ":")
+++				if ok {
+++					// Verify credentials
+++					valid, err := config.Validator(user, pass, c)
+++					if err != nil {
+++						return err
+++					} else if valid {
+++						return next(c)
+++					}
				}
```

### 2. Added Realm quoting in WWW-Authenticate header:
RFC 7617 requires that the value of the realm parameter be a quoted-string. In the current implementation, the default realm is unquoted. You can comply with this specification by always using strconv.Quote

```diff
// basic_auth.go
---	realm := defaultRealm
---			if config.Realm != defaultRealm {
---				realm = strconv.Quote(config.Realm)
---			}

			// Need to return `401` for browsers to pop-up login box.
---			c.Response().Header().Set(echo.HeaderWWWAuthenticate, basic+" realm="+realm)
+++			// Realm is case-insensitive, so we can use "basic" directly. See RFC 7617.
+++			c.Response().Header().Set(echo.HeaderWWWAuthenticate, basic+" realm="+strconv.Quote(config.Realm))
			return echo.ErrUnauthorized

```

These changes will further improve the robustness and maintainability of the middleware.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Sorry, I accidentally deleted the local directory I forked from, so I created a new Pull Request!

## Changes
Changed the PANIC message to be more specific.
`Tests related to it''s changes`

## Improved debugging efficiency.
When Flush is not supported, it is clear which ResponseWriter is the culprit, speeding problem identification and resolution.

## Improved Developer Experience:
More specific error information makes it easier for developers to understand the cause of problems.
More specific error information helps developers understand the cause of problems.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2795'
    where id = _change_id;

    -- Echo PR pair 64: #2792 creates the Change, #2790 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Documentation'),
        'fixed issue #2791 Improved robustness of Response process and added debugging information',
        '## Changes
Improvements have been made to the WriteHeader and Flush methods of response.go.

## 1. **Check logger existence with ``WriteHeader``**:
`r.echo` and `r.echo.Logger.Warn` are not `nil` before calling `r.echo.Logger.Warn`. This reduces the risk of panic due to unexpected nil pointer references

## 2. **Additional error logging with `Flush`**:.
 If `http.ResponseController.Flush()` returns an error other than `http.ErrNotSupported`, it will now log the error if in debug mode (`r.echo.Debug == true`). This additional logging is useful for debugging during development, since the `Flush` method of the `http.Flusher` interface is by convention not to return errors, but the `ResponseController` may.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## Changes
Changed the PANIC message to be more specific.

## Improved debugging efficiency.
When Flush is not supported, it is clear which ResponseWriter is the culprit, speeding problem identification and resolution.

## Improved Developer Experience:
More specific error information makes it easier for developers to understand the cause of problems.
More specific error information helps developers understand the cause of problems.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2792'
    where id = _change_id;

    -- Echo PR pair 65: #2787 creates the Change, #2783 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Maintenance'),
        'Add authorization header handling in proxy middleware',
        '## Problem

When a proxy target URL contains User credentials, proxy middleware does not send request with the authorization header. This can lead to UnAuthorized Error (401) when connecting to upstream server.

## Solution

- Implemented logic to pass the Authorization header to the target if the proxy URL includes user credentials.
- Added unit tests to verify behavior for both scenarios: with and without user credentials in the proxy URL.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Closed Echo pull request #2783 did not include a body.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2787'
    where id = _change_id;

    -- Echo PR pair 66: #2782 creates the Change, #2781 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Router'),
        'Revert "CORS: reject requests with 401 for non-preflight request with not matching origin header"',
        'Reverts labstack/echo#2732

See: https://github.com/labstack/echo/issues/2767'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Closed Echo pull request #2781 did not include a body.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2782'
    where id = _change_id;

    -- Echo PR pair 67: #2780 creates the Change, #2764 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Middleware'),
        'Upgrade dependencies',
        'related to  https://github.com/labstack/echo/discussions/2779

* https://pkg.go.dev/vuln/GO-2025-3487  (affects: `golang.org/x/crypto/ssh`)
* https://pkg.go.dev/vuln/GO-2025-3503 (affects: `golang.org/x/net/http/httpproxy` and `golang.org/x/net/proxy` )
* https://pkg.go.dev/vuln/GO-2025-3595 (affects: `golang.org/x/net/html` )'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Added test explicit verification for [reuse CSRF token logic](https://github.com/labstack/echo/blob/3598f295f95f316bbeb252b7b332fe34e120815c/middleware/csrf.go#L136):
- Strictly validates token is reused when cookie exists
- Confirms new token is generated when no cookie provided

This reinforces detailed test for CSRF token behavior.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2780'
    where id = _change_id;

    -- Echo PR pair 68: #2762 creates the Change, #2760 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Binder'),
        'Add support for TLS WebSocket proxy',
        'This PR fixes the issue below.
https://github.com/labstack/echo/issues/2200


## Detail

- middleware/proxy.go
  - I change proxyRaw method to be able to proxy tls connection
- middleware/proxy_test.go
  - I added two tests. In this test, I use [golang.org/x/net/websocket](https://pkg.go.dev/golang.org/x/net/websocket)

## About websocket package in testing
I used an external package to easily implement web socket testing.

We considered the following packages, but we thought it would be better not to add too many third-party package dependencies for echo, so we adopted the official golang.org/x/net/websocket.

https://pkg.go.dev/github.com/gorilla/websocket
https://pkg.go.dev/github.com/coder/websocket'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Closed Echo pull request #2760 did not include a body.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2762'
    where id = _change_id;

    -- Echo PR pair 69: #2755 creates the Change, #2753 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Documentation'),
        'Add Transfer-Encoding as header',
        'Closed Echo pull request #2755 did not include a body.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Closed Echo pull request #2753 did not include a body.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2755'
    where id = _change_id;

    -- Echo PR pair 70: #2752 creates the Change, #2750 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Maintenance'),
        'feat: Add Mountpoint interface and related tests for App routing flexibility',
        '# Add Mountpoint Interface for Flexible Routing

## Description
This PR introduces a new `Mountpoint` interface that abstracts common routing methods shared between `Echo` and `Group` types. This enhancement improves code flexibility by allowing applications to register routes against either an `Echo` instance or a `Group` without needing to know which type they''re working with.

## Changes
- Added new `mountpoint.go` file defining the `Mountpoint` interface with all common routing methods
- Added comprehensive tests in `mountpoint_test.go` demonstrating the interface usage with both `Echo` instances and `Group` instances
- Verified that both `Echo` and `Group` types properly implement the interface

## Benefits
- Enables more modular application design by allowing route registration against any mountpoint
- Simplifies code that needs to work with both `Echo` instances and route groups
- Facilitates better code organization in larger applications with complex routing needs
- Maintains backward compatibility with existing code

## Testing
The PR includes extensive tests that verify:
- Basic functionality with both `Echo` and `Group` instances
- Proper handling of route prefixes with groups
- Nested group functionality
- Multiple mountpoint registration
- Middleware application

All tests pass successfully, confirming the interface works as expected.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'relates to #23

- Replaces large switch statement to a map to lower CCN from 15 to 5 (67% reduction).',
        pull_request_url = 'https://github.com/labstack/echo/pull/2752'
    where id = _change_id;

    -- Echo PR pair 71: #2749 creates the Change, #2748 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Router'),
        'test(insertNode): add the second unit test for insertNode function',
        'Related to #5'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Go 1.24 was released https://tip.golang.org/doc/go1.24',
        pull_request_url = 'https://github.com/labstack/echo/pull/2749'
    where id = _change_id;

    -- Echo PR pair 72: #2744 creates the Change, #2735 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Middleware'),
        'feat: add Forwarded header parsing and real IP extraction with tests',
        'For issue #2694

- Added parser for the "Forwarded" header to extract the "for" field.
- Implemented real IP extraction from the "Forwarded" headers.
- Added unit tests to validate header parsing and real IP extraction functionality.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Closed Echo pull request #2735 did not include a body.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2744'
    where id = _change_id;

    -- Echo PR pair 73: #2733 creates the Change, #2732 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Binder'),
        'Fix/only set request id if not exists',
        'A small optimization of not trying to reset request id incase one already exists'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'reject requests with 401 for non-preflight request with not matching origin header

fixes #2730',
        pull_request_url = 'https://github.com/labstack/echo/pull/2733'
    where id = _change_id;

    -- Echo PR pair 74: #2727 creates the Change, #2722 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Documentation'),
        'Fix error in README example code',
        'the example code is missing an `errors` package import'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Update golang.org/x/net dependency [GO-2024-3333](https://pkg.go.dev/vuln/GO-2024-3333)',
        pull_request_url = 'https://github.com/labstack/echo/pull/2727'
    where id = _change_id;

    -- Echo PR pair 75: #2721 creates the Change, #2719 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Maintenance'),
        'Update dependencies (dependabot reports https://pkg.go.dev/vuln/GO-2024-3321',
        'Update dependencies. dependabot reports https://pkg.go.dev/vuln/GO-2024-3321 / https://github.com/advisories/GHSA-v778-237x-gjrc

I do not see us directly affected us but dependabot reports are going to make people making issues/ticket in this repo. so it is better to prevent that.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'https://github.com/labstack/echo/pull/2717 which fixes https://github.com/labstack/echo/pull/2710',
        pull_request_url = 'https://github.com/labstack/echo/pull/2721'
    where id = _change_id;

    -- Echo PR pair 76: #2717 creates the Change, #2715 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Router'),
        'Add Conditions to Ensure Bind Succeeds with `Transfer-Encoding: chunked`',
        'In Go, when the length of the Body is unknown, `ContentLength` is set to -1. As demonstrated in #2716, this applies to cases where `Transfer-Encoding: chunked` is used.

```
	// ContentLength records the length of the associated content.
	// The value -1 indicates that the length is unknown.
	// Values >= 0 indicate that the given number of bytes may
	// be read from Body.
	//
	// For client requests, a value of 0 with a non-nil Body is
	// also treated as unknown.
	ContentLength [int64](https://pkg.go.dev/builtin#int64)
```

https://pkg.go.dev/net/http#Request

Previously, only `0` was excluded during Bind, but starting from #2710, `-1` was also excluded, making it impossible to Bind requests with `Transfer-Encoding: chunked`. I have fixed this issue.


Fixes #2716.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'echo.Static serves files but unsorted. To make them sorted, I''ve followed the same procedure as was done for [net/http](https://github.com/golang/go/issues/11879) (the fix they did is shown [here](https://github.com/golang/go/commit/25b00177af9f62f683ec68f1d697c2607d087ea6#diff-0661442fffb473f85dc4d4472172edbfb4b9b1837db3ab1a73e838bed3e6ab70R597)).',
        pull_request_url = 'https://github.com/labstack/echo/pull/2717'
    where id = _change_id;

    -- Echo PR pair 77: #2713 creates the Change, #2712 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Middleware'),
        'Update README.md',
        'Update to install and with version.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Next release `v4.13.0` will be Wednesday 2024.12.04.  This will upset probably quite a few people as we have breaking change. At least it is not on Friday :)

I will wait till December, after the Black Friday is over.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2713'
    where id = _change_id;

    -- Echo PR pair 78: #2711 creates the Change, #2710 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Binder'),
        'Shorten Github issue template and add test example',
        'Issue template is such a hassle to fill. hardly anyone does that correctly. it is better to concentrate people focus to provide test to demonstrate the problem than filling paragraphs that they will not fill, most of the time.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'I am writing a unit test and using httptest.NewRequestWithContext to create an http request, it will return a new http request with content-length = -1 with body as http.NoBody

```go
httpReq := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
httpReq.Header.Set(echo.HeaderContentType, "application/json")
```

https://github.com/golang/go/blob/master/src/net/http/httptest/httptest.go#L46C1-L77C3
```go
func NewRequestWithContext(ctx context.Context, method, target string, body io.Reader) *http.Request {
	if method == "" {
		method = "GET"
	}
	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(method + " " + target + " HTTP/1.0\r\n\r\n")))
	if err != nil {
		panic("invalid NewRequest arguments; " + err.Error())
	}
	req = req.WithContext(ctx)

	// HTTP/1.0 was used above to avoid needing a Host field. Change it to 1.1 here.
	req.Proto = "HTTP/1.1"
	req.ProtoMinor = 1
	req.Close = false

	if body != nil {
		switch v := body.(type) {
		case *bytes.Buffer:
			req.ContentLength = int64(v.Len())
		case *bytes.Reader:
			req.ContentLength = int64(v.Len())
		case *strings.Reader:
			req.ContentLength = int64(v.Len())
		default:
			req.ContentLength = -1
		}
		if rc, ok := body.(io.ReadCloser); ok {
			req.Body = rc
		} else {
			req.Body = io.NopCloser(body)
		}
	}
```
this is inconsistent in go-source-code as http.NewRequestWithContext with body as http.NoBody will have content-length as 0. i don''t know if this should be fixed in go-source-code. https://github.com/golang/go/issues/18117
https://github.com/golang/go/blob/master/src/net/http/request.go#L946-L951
```go
func NewRequestWithContext(ctx context.Context, method, url string, body io.Reader) (*Request, error) {
...
      if body != nil {
      ...
		case *bytes.Reader:
			req.ContentLength = int64(v.Len())
			snapshot := *v
			req.GetBody = func() (io.ReadCloser, error) {
				r := snapshot
				return io.NopCloser(&r), nil
			}
		case *strings.Reader:
			req.ContentLength = int64(v.Len())
			snapshot := *v
			req.GetBody = func() (io.ReadCloser, error) {
				r := snapshot
				return io.NopCloser(&r), nil
			}
		default:
			// This is where we''d set it to -1 (at least
			// if body != NoBody) to mean unknown, but
			// that broke people during the Go 1.8 testing
			// period. People depend on it being 0 I
			// guess. Maybe retry later. See Issue 18117.
```',
        pull_request_url = 'https://github.com/labstack/echo/pull/2711'
    where id = _change_id;

    -- Echo PR pair 79: #2709 creates the Change, #2705 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Documentation'),
        'CORS middleware should compile allowOrigin regexp at creation',
        'This change preserves previous behavior - invalid patterns are just ignored.

we can not add panics as this would cause runtime unrecovered panics (i.e. some people very angry as their applications crash at prod)

Reported by https://github.com/labstack/echo/issues/2708'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Reported by https://github.com/labstack/echo/issues/2703',
        pull_request_url = 'https://github.com/labstack/echo/pull/2709'
    where id = _change_id;

    -- Echo PR pair 80: #2702 creates the Change, #2701 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Maintenance'),
        'Fix issue #2694',
        'Fix Fix issue #2694 and add "Forwarded"'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'For #2699

[Seems consensus is to remove this middleware](https://github.com/labstack/echo/issues/2699#issuecomment-2464675851) and rely on the middleware for  https://github.com/labstack/echo-jwt .

This PR removes the dependencies and the middleware.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2702'
    where id = _change_id;

    -- Echo PR pair 81: #2700 creates the Change, #2698 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Router'),
        'dep: update golang-jwt to v4.5.1',
        'Fixes #2699

We want to avoid a known vulnerability in golang-jwt library is flagged as a security concern when using echo as a framework in our applications.

Tests are passing locally with the new version.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## CHANGE
use method, `echo.AcqurireContext` which defined.

```diff
-func (e *Echo) AcquireContext() Context {
-	return e.pool.Get().(Context)
+func (e *Echo) AcquireContext() *context {
+     return e.pool.Get().(*context)
}

func (e *Echo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Acquire context
-       c := e.pool.Get().(*context)
+	c := e.AcquireContext()
        ....
        // Release context
-	e.pool.Put(c)
+	e.ReleaseContext(c)
}
```',
        pull_request_url = 'https://github.com/labstack/echo/pull/2700'
    where id = _change_id;

    -- Echo PR pair 82: #2695 creates the Change, #2692 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Middleware'),
        'Refactor basic auth middleware to support multiple auth headers',
        'This is taken from `v5` branch + improved tests

reasoning for it:
multiple auth headers is something that can happen in environments like corporate test environments that are secured by application proxy servers where front facing proxy is configured to require own basic auth value + checks it and your application also requires basic auth headers from clients.  As Go standard library stores headers in map and keys are retrieved in random order the middleware may need to check multiple headers to match correct one.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## Summary
This PR formats `interface{}` -> `any`.
`any` is an alias for `interface{}`
```go
var any = interface{}
```
https://github.com/golang/go/blob/67f131485541f362c8e932cd254982a8ad2cfc09/src/builtin/builtin.go#L97

## What was changed
- `interface{}` -> `any`
  - it required `go >= 1.18`
- It has nothing to do with performance.

## Why the change was made
- IMO: `any` is easier than an empty interface literal.

## How it was tested
```sh
go test ./... -cover
```',
        pull_request_url = 'https://github.com/labstack/echo/pull/2695'
    where id = _change_id;

    -- Echo PR pair 83: #2691 creates the Change, #2690 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Binder'),
        'set cookie to request',
        'When setting a cookie to response also set it the request so that it can be retrieved in the same request.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Add TemplateRenderer struct to ease creating renderers for `html/template` and `text/template` packages.

Different take on #2673 ideas',
        pull_request_url = 'https://github.com/labstack/echo/pull/2691'
    where id = _change_id;

    -- Echo PR pair 84: #2688 creates the Change, #2684 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Documentation'),
        'Refactor TestBasicAuth to utilize table-driven test format',
        '### Summary
This PR refactors the `basic_auth_test.go` file to use a table-driven test approach. The new structure improves readability, simplifies the addition of new test cases, and makes it easier to maintain the tests as the codebase evolves.

### What was changed
- Refactored individual test cases within TestBasicAuth into a table-driven test.
- Moved repeated logic into the test table to reduce redundancy.
- No changes were made to production code.

### Why the change was made
Table-driven tests provide a more scalable way to manage and add test cases. This refactor ensures that future test cases can be added with minimal repetition.

### How it was tested
- From the ''middleware'' directory, ran `go test -v -run TestBasicAuth`. All tests passed successfully.
- Manually checked each test within TestBasicAuth to ensure that if test conditions were changed the test failed.
- No functional changes were made to the codebase, so there should be no impact on production.

### Follow on
Am willing to refactor more tests if the table-driven format is desirable'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '**Summary**
This PR introduces support for handling multipart requests that contain multiple files in the `bind` function. It extends the current functionality to allow seamless parsing and binding of multiple files uploaded through multipart form data.

**Changes**
- Modified the `bind` logic to handle multiple file uploads in a single request.
- Added support for accessing multiple files via `FormFile`.
- Updated the relevant documentation and comments for clarity.
- Introduced test cases to validate multi-file uploads.

**Testing**
- Added unit tests to ensure multi-file binding is working as expected.

**Additional Information**
This update improves the flexibility of the `bind` function when dealing with file uploads, making it easier to handle bulk file operations in a single request.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2688'
    where id = _change_id;

    -- Echo PR pair 85: #2683 creates the Change, #2682 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Maintenance'),
        'Improve error control & define limits',
        'In this PR:
- Improved error control
- Define a startup limit for arrays'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'improve `MultipartForm` to make it easier to know how to use it',
        pull_request_url = 'https://github.com/labstack/echo/pull/2683'
    where id = _change_id;

    -- Echo PR pair 86: #2675 creates the Change, #2673 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Router'),
        'Add Go 1.23 to CI',
        'https://tip.golang.org/doc/go1.23

So we support Go 1.20, 1.21, 1.22, 1.23 from now.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'I have added a pre-built templates function to render HTML easily.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2675'
    where id = _change_id;

    -- Echo PR pair 87: #2671 creates the Change, #2664 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Middleware'),
        'Improve Logger Middleware Documentation',
        'This pull request enhances the documentation for Echo''s Logger middleware. The improvements aim to provide clearer, more comprehensive information for users implementing logging in their Echo applications.

Key changes include:
- Expanded overview of the Logger middleware and its features
- More detailed configuration options with examples
- Advanced usage scenarios, including custom output and route-specific logging
- Best practices for using the Logger middleware effectively
- Performance considerations for high-traffic scenarios

These updates should help both new and experienced Echo users better understand and utilize the Logger middleware, leading to improved debugging and monitoring capabilities in their applications.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'github.com/golang-jwt/jwt v3 version is not getting updates anymore, updated to v5
all tests pass',
        pull_request_url = 'https://github.com/labstack/echo/pull/2671'
    where id = _change_id;

    -- Echo PR pair 88: #2660 creates the Change, #2659 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Binder'),
        'Implement Context interceptor for ServeHTTP',
        'I need to adjust the Context when using ServeHTTP, when not feasible with normal middleware.

Consider the following scenario.

When using WebSockets, I want to convert a received message to a http.Request.
Together with a custom ResponseWriter the request is handled by echo''s ServeHTTP.
But I want to adjust the Context and add my Socket reference to it.

There is no way to access the Context, except via middleware, but when handling middleware I have no actual access the the websocket anymore.

The easiest way to do this without interfering with other functionality or adding new functions, is checking if ResponseWriter conforms to a specific interface (ServeHTTPContextInterceptor) , if so, execute the interceptor and continue handling with the adjusted context.

Now a custom ResponseWriter can implement this interface and return a custom Context before `handle` and cleanup aftwards.

```
func (responseWriter *responseWriter) InterceptContext(ctx Echo.Context, handle func(ctx Echo.Context)) {

	handle(ctx)
}
```

Implemented and running in production at scale.

Another implemented usecase:

I have an Amazon AWS API Gateway routing all traffic to one Lambda Function, written in golang.

On receiving the call, I create a http.Request and ResponseWriter and perform ServeHTTP on echo, to handle the call.

Now I can create a Custom Context that hold various extra details about the request from API gateway.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'I need to adjust the context when using ServeHTTP, this is not feasible with normal middleware.

Consider the following scenario.

When using WebSockets, I want to convert a message to  http.Request.
Together with a custom ResponseWriter the request is handled by echo''s ServeHTTP.
But I want to adjust the Context and add my Socket reference to it.

There is no way to access the Context, except via middleware, but when handling middleware I have no actual access the the websocket anymore.

The easiest way to do this without interfering with other functionality or adding new functions, is checking if ResponseWriter conforms to a specific interface (ServeHTTPContextInterceptor) , if so, execute the interceptor and continue handling with the adjusted context.

Now a custom ResponseWriter can implement this interface and return a custom Context before `handle` and cleanup aftwards.

```
func (responseWriter *responseWriter) InterceptContext(ctx Echo.Context, handle func(ctx Echo.Context)) {

	handle(ctx)
}
```

Implemented and running in production at scale.

Another implemented usecase:

I have an Amazon AWS API Gateway routing all traffic to one Lambada Function, written in golang.

On receiving the call, I create a http.Request and ResponseWriter and perform ServeHTTP on echo, to handle the call.

Now I can create a Custom Context that hold various extra details about the request from API gateway.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2660'
    where id = _change_id;

    -- Echo PR pair 89: #2657 creates the Change, #2656 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Documentation'),
        'fix: Unclear behaviour of * in routes',
        'This should hopefully bring more clarity into this issue https://github.com/labstack/echo/issues/2619

Examples would explain this better so here''s similar to one from discussion

```
func main() {
	handler := func(c echo.Context) error {
		fmt.Printf("%s\n", c.Path())
		return nil
	}

	e := echo.New()
	v2 := e.Group("/v2")

	v2.DELETE("/*/blobs/:digest", handler)
	v2.GET("/*/blobs/:digest", handler)
	v2.HEAD("/*/blobs/:digest", handler)

	v2.DELETE("/*/manifests/:ref", handler)
	v2.GET("/*/manifests/:ref", handler)
	v2.HEAD("/*/manifests/:ref", handler)
	v2.PUT("/*/manifests/:ref", handler)

	v2.GET("/*/tags/list", handler)     // one wildcard
	v2.GET("/*/*/tags/list", handler)   // two wildcards
	v2.GET("/*/*/tags/list2*", handler) // two wildcards and trailing one
	v2.GET("/*/*/tags/list2", handler)  // two wildcards with fixed ending that may conflict with previous route

	v2.GET("/*/blobs/uploads/:ref", handler)
	v2.PATCH("/*/blobs/uploads/:ref", handler)
	v2.POST("/*/blobs/uploads", handler)
	v2.PUT("/*/blobs/uploads/:ref", handler)

	v2.GET("", handler)
	err := e.Start(":8080")
	if err != nil {
		panic(err)
	}
}
```

Example curl calls that should match those routes:

```
curl -ik  http://localhost:8080/v2/wildcard1/tags/list
# Matches /*/tags/list

curl -ik  http://localhost:8080/v2/wildcard1/wildcard2/tags/list
# Matches /*/*/tags/list

curl -ik  http://localhost:8080/v2/wildcard1/wildcard2/tags/list2wildcard3
# Matches /*/*/tags/list2*

curl -ik  http://localhost:8080/v2/wildcard1/wildcard2/tags/list2
# Matches /*/*/tags/list2
```


This one doesn''t match since you would need to register it as `/*/*/*/tags/list` and this probably better implemented as separate `**` wildcard
```
curl -ik  http://localhost:8080/v2/wildcard1/wildcard2/nowildcard/tags/list
{"message":"Method Not Allowed"}
```'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'This PR modifies the bindData function to preserve the pre v4.12.0 behavior for **map[string]interface{}** while supporting the new functionality:

- Single values are stored as strings
- Multiple values are stored as []string

This approach maintains compatibility with existing code that expects single string values, while allowing new code to take advantage of multiple values when present.

The change addresses the issue reported in https://github.com/labstack/echo/issues/2652, where the binding behavior for map[string]interface{} changed in v4.12.0, potentially breaking existing implementations.

Testing:
- Updated existing tests to reflect the new behavior

Please review and let me know if any further changes are needed.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2657'
    where id = _change_id;

    -- Echo PR pair 90: #2655 creates the Change, #2654 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Maintenance'),
        'Allow content type middleware',
        'Based on #2551

Adds a new middleware `AllowContentType` which restricts routes to only accepts requests with certain `Content-Type` header values.

It''s similar to the middleware of the same name available with Chi

The middleware also sets the `Accept` header field to the specified content types for all requests'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'just added http:// or https:// in front of the address when printed into the console to make it clickable.
probably better if i print it myself but this way its the last thing that''s printed in the console and it''s just easier',
        pull_request_url = 'https://github.com/labstack/echo/pull/2655'
    where id = _change_id;

    -- Echo PR pair 91: #2653 creates the Change, #2636 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Router'),
        'chore: fix typo',
        'Closed Echo pull request #2653 did not include a body.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## Pull requset to resolve #2632

Modified files with saved memory size:

| File name   | Saved size(bytes) |
| ----------- | ----------------- |
| binder.go   | 40                |
| echo.go     | 8                 |
| echo_fs.go  | 8                 |
| group.go    | 16                |
| ip.go       | 8                 |
| response.go | 16                |
| router.go   | 32                |
| context.go  | 16                |
| echo.go     | 8                 |


**Total saved size:** 144

There was potential to further optimize the `Echo` struct in the `echo.go` file, but I chose to respect the readability of the codebase.

@aldas a quick review from you would be awesome whenever you’re free. Thanks!',
        pull_request_url = 'https://github.com/labstack/echo/pull/2653'
    where id = _change_id;

    -- Echo PR pair 92: #2633 creates the Change, #2631 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Middleware'),
        'Update README.md',
        'A new tool that uses Echo as part of its core.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = '## What

replaced `interface{}` to  `any`.

## Why

golang 1.18 support any as type alias for interafce{}.
If echo follow golang support policy, the smarter using any than using interface{}.

## Concerns

Impact for an application run the environment that golang version  is < golang 1.18.
The application run at environment is  under golang 1.18  will be failed to run, since type alias any is supported after golang 1.18.
If echo support the application run at environment is under golang 1.18, this PR should be contained to v5.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2633'
    where id = _change_id;

    -- Echo PR pair 93: #2627 creates the Change, #2626 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Binder'),
        'When struct tag is not set, use actual field name for binding',
        'recently been using echo and even porting apps to echo from net/http and others.

i have added some tiny improvements in struct binding:
- if struct tag is missing then try to use field name as lookup key in data map (similar behavior as bson)
  - this feature is controlled by flag, `(*Echo).Binder` continues to work as it is
  - user can opt in to use this feature by setting `(*Echo).Binder = echo.BinderWithFallback()`)
- if the struct tag value is specified as a dash `-` it is skipped (similar behavior as encoding/json)
  - this feature is right there without a flag but it should not be a problem as reading certain value from `-` field is extremely unlikely and us gophers already treat it as a skipper (from well known behavior of json and others)'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Changelog for v4.12.0

I''ll tag it as minor version as we have quite a lot of different things here this time',
        pull_request_url = 'https://github.com/labstack/echo/pull/2627'
    where id = _change_id;

    -- Echo PR pair 94: #2625 creates the Change, #2624 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Documentation'),
        'Update golang.org/x/* deps',
        '`golang.org/x/net` needs to be updated

> Vulnerability #1: GO-2024-2687
    HTTP/2 CONTINUATION flood in net/http
  More info: https://pkg.go.dev/vuln/GO-2024-2687
  Module: golang.org/x/net
    Found in: golang.org/x/net@v0.[22](https://github.com/labstack/echo/actions/runs/8693604426/job/23840712643?pr=2624#step:6:23).0
    Fixed in: golang.org/x/net@v0.[23](https://github.com/labstack/echo/actions/runs/8693604426/job/23840712643?pr=2624#step:6:24).0
    Example traces found:'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Relates to #1172

Use `httputil.ReverseProxy` to proxy SSE requests as it has support for streaming responses. See:
https://github.com/golang/go/blob/b107d95b9a66bfe7150fd4f2915e9bb876a6999a/src/net/http/httputil/reverseproxy.go#L601

------------

can be tested with

1. create separate package and execute this code to start serving proxy application at port 8080 that proxies requests to localhost:8081

```go
package main

import (
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"net/url"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	tmpURL, err := url.Parse("http://localhost:8081")
	if err != nil {
		log.Fatal(err)
	}
	e.Use(middleware.Proxy(middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{{URL: tmpURL}})))

	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

```

2. Create application for serving SSE

Go file for application
```go
package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"time"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.File("/", "./index.html")

	e.GET("/sse", func(c echo.Context) error {
		log.Printf("SSE client connected, ip: %v", c.RealIP())

		w := c.Response()
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-c.Request().Context().Done():
				log.Printf("SSE client disconnected, ip: %v", c.RealIP())
				return nil
			case <-ticker.C:
				event := Event{
					Data: []byte("ping: " + time.Now().Format(time.RFC3339Nano)),
				}
				if err := event.WriteTo(w); err != nil {
					return err
				}
				w.Flush()
			}
		}
	})

	if err := e.Start(":8081"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

// Event structure is defined here: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events#event_stream_format
type Event struct {
	ID      []byte
	Data    []byte
	Event   []byte
	Retry   []byte
	Comment []byte
}

// WriteTo writes Event to given ResponseWriter
func (ev *Event) WriteTo(w http.ResponseWriter) error {
	// Marshalling part is taken from: https://github.com/r3labs/sse/blob/c6d5381ee3ca63828b321c16baa008fd6c0b4564/http.go#L16
	if len(ev.Data) == 0 && len(ev.Comment) == 0 {
		return nil
	}

	if len(ev.Data) > 0 {
		if _, err := fmt.Fprintf(w, "id: %s\n", ev.ID); err != nil {
			return err
		}

		sd := bytes.Split(ev.Data, []byte("\n"))
		for i := range sd {
			if _, err := fmt.Fprintf(w, "data: %s\n", sd[i]); err != nil {
				return err
			}
		}

		if len(ev.Event) > 0 {
			if _, err := fmt.Fprintf(w, "event: %s\n", ev.Event); err != nil {
				return err
			}
		}

		if len(ev.Retry) > 0 {
			if _, err := fmt.Fprintf(w, "retry: %s\n", ev.Retry); err != nil {
				return err
			}
		}
	}

	if len(ev.Comment) > 0 {
		if _, err := fmt.Fprintf(w, ": %s\n", ev.Comment); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprint(w, "\n"); err != nil {
		return err
	}

	return nil
}
```

in the same folder as app create index.html

```html
<!DOCTYPE html>
<html>
<body>

<h1>Getting server updates</h1>
<div id="result"></div>

<script>
  // Example taken from: https://www.w3schools.com/html/html5_serversentevents.asp
  if (typeof (EventSource) !== "undefined") {
    const source = new EventSource("/sse");
    source.onmessage = function (event) {
      document.getElementById("result").innerHTML += event.data + "<br>";
    };
  } else {
    document.getElementById("result").innerHTML = "Sorry, your browser does not support server-sent events...";
  }
</script>

</body>
</html>
```

3. Open http://localhost:8080 in your browser. You should see Ping messages streamed, assuming proxy middleware handles SSE requests as raw proxy',
        pull_request_url = 'https://github.com/labstack/echo/pull/2625'
    where id = _change_id;

    -- Echo PR pair 95: #2618 creates the Change, #2616 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Maintenance'),
        'fix(middleware): Fix Allow method of RateLimiterMemoryStore',
        '## What
Modify `(*RateLimiterMemoryStore).Allow` method in rate limiter middleware.

## Why
Currently, `Allow` method acts unexpected behavior that it denies the request nevertheless subtract of `lastSeen`  and `now` exceeds `expiresIn`.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'When route is registered with empty path it is normalized to `/`. Make sure that returned echo.Route structs reflect that behavior.  Internally router has changed `` (empty path) to `/` for a long time but Route that is returned did not reflect that. Is is problematic with `Reverse` function that uses empty string as "not found"

Related to #2615

Previous behavior can be seen with this test:
```go
func TestTest(t *testing.T) {
	e := echo.New()

	handler := func(c echo.Context) error {
		return c.String(http.StatusNotImplemented, "Nope")
	}
	r := e.GET("", handler) // path is registered as "" previously. After change `/` is registered
	r.Name = "test"

	existingEmpty := e.Reverse("test")
	assert.Equal(t, "", existingEmpty)

	notExistingEmpty := e.Reverse("not-existing")
	assert.Equal(t, "", notExistingEmpty)
}
```

whis this change `assert.Equal(t, "", existingEmpty)` shoulb be change to  `assert.Equal(t, "/", existingEmpty)`to pass the test',
        pull_request_url = 'https://github.com/labstack/echo/pull/2618'
    where id = _change_id;

    -- Echo PR pair 96: #2611 creates the Change, #2609 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Router'),
        'Remove maxparam dependence from Context',
        'This addresses race when Context is changing echo.Echo.maxParam value. This change does not address issues with Router not being safe when routes are added when server has already started and is serving incoming requests. This case is still unsafe (multiple goroutines writing+reading route tree)

Relates to https://github.com/labstack/echo/issues/1705'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'I have fixed the TargetHeader option of the RequestIDConfig, which was disabled.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2611'
    where id = _change_id;

    -- Echo PR pair 97: #2608 creates the Change, #2607 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Middleware'),
        'Default binder can bind pointer to slice as struct field. For example  `*[]string`',
        'Default binder can bind pointer to slice as struct field. For example `*[]string`

Related issue https://github.com/labstack/echo/issues/2381'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Default binder can use `UnmarshalParams(params []string) error` interface to bind multiple input values at one go.

Relates to https://github.com/labstack/echo/pull/2602

This allows developers to build fancy unmarsallers like related PR had. that turns `/?a=1,2,3&a=4,5,6` into `IntArrayB([]int{1, 2, 3, 4, 5, 6})`

```go
type IntArrayB []int

func (i *IntArrayB) UnmarshalParams(params []string) error {
	var numbers = make([]int, 0, len(params))

	for _, param := range params {
		var values = strings.Split(param, ",")
		for _, v := range values {
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return fmt.Errorf("''%s'' is not an integer", v)
			}
			numbers = append(numbers, int(n))
		}
	}

	*i = append(*i, numbers...)
	return nil
}

func TestBindUnmarshalParams(t *testing.T) {
	t.Run("ok, target is an alias to slice and is nil, append multiple inputs", func(t *testing.T) {
		e := New()
		req := httptest.NewRequest(http.MethodGet, "/?a=1,2,3&a=4,5,6", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		result := struct {
			V IntArrayB `query:"a"`
		}{}
		err := c.Bind(&result)

		assert.NoError(t, err)
		assert.Equal(t, IntArrayB([]int{1, 2, 3, 4, 5, 6}), result.V)
	})
}
```',
        pull_request_url = 'https://github.com/labstack/echo/pull/2608'
    where id = _change_id;

    -- Echo PR pair 98: #2606 creates the Change, #2605 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Binder'),
        'Change type definition blocks to single declarations. This helps copy…',
        'Change type definition blocks to single declarations. This helps copy/pasting Echo code in examples (for issues and discussions and for Echo website).

I admit this is yak shaving but answering issues with good examples is probably biggest time sink when it comes to maintaining Echo. And I feel that is important to answer with examples etc.   Also - website middleware part has block with middleware configurations. It is dedious to copy/paste middleware conf type there and have markdown intentation broken (blocks like that https://github.com/labstack/echox/blob/master/website/docs/middleware/csrf.md#configuration).'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Upgrade deps',
        pull_request_url = 'https://github.com/labstack/echo/pull/2606'
    where id = _change_id;

    -- Echo PR pair 99: #2604 creates the Change, #2603 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Documentation'),
        'Add SPDX licence comments to files.',
        'Add SPDX licence comments to files. See https://spdx.dev/learn/handling-license-info/  There have been cases when Echo code has been copied but copyright reference has been hard to find in these repos. With these comments it should easier to understand what parts derive from Echo. Assuming these are not removed - but in that case, this is out of our hands.

NB: year is 2015 as this is @vishr first commit year in this repo. copypright number does not need to be updated every year.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'Closed Echo pull request #2603 did not include a body.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2604'
    where id = _change_id;

    -- Echo PR pair 100: #2602 creates the Change, #2596 supplies pull_request_body.
    select public.fn_change_insert(
        _project_id,
        array['feature']::text[],
        (select id from public.epic where project_id = _project_id and name = 'Echo Maintenance'),
        'Aliased slices should be consistent with builtin slices',
        'In working with an API at work we had the use case on query params to accept the following syntaxes:

- `?foo=1&foo=2&foo=3`
- `?foo=1,2,3`

Initially we were using only the former syntax, but then a new library was developed that sent the latter syntax.  When we switched from defining our field as `[]int` to a custom type `Ints` that implemented `UnmarshalParam` we found that when there were multiple values for the same key, the `UnmarshalParam` would only receive the first value and others would be lost.

This change brings
consistency between the builtin slices and the aliased slices for query parameters.'
    ) into _change_id;
    update public.change
    set
        pull_request_body = 'I just fixed some typos.',
        pull_request_url = 'https://github.com/labstack/echo/pull/2602'
    where id = _change_id;

end;
$$;

insert into public.test_case (change_id, scenario, done)
select c.id, seed.scenario, seed.done
from public.change c
join (values
    ('Respect q=0 in gzip content negotiation', 'Created from the first Echo pull request in a paired seed import.', true),
    ('perf(json): pooled-buffer JSON deserialize', 'Pull request body was updated from the second Echo pull request in its pair.', true),
    ('backport PR 3016 from v5 to v4', 'Pull request URL points to the first Echo pull request in its pair.', false),
    ('Update CI action versions for v4 branch', 'Generated ref was allocated by fn_change_insert after demo1 last_ref was advanced.', false)
) as seed(title, scenario, done)
    on seed.title = c.title;

do
$$
declare
    _change record;
begin
    for _change in select id from public.change loop
        call public.sp_change_test_case_recalculate(_change.id);
    end loop;
end;
$$;

commit;
