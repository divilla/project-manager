# CLI Flow Architecture

## Flow

idea
code-init?                # optional user-provided implementation
spec-write
spec-review
code-write
code-check?               # optional user acceptance/refinement
spec-amend?               # if user extended/changed requirements
docs?                     # created only on user request
PR-write
code-review
code-fix                  # user approved and navigated code fixes
spec-update               # reconcile Spec and QA cases with final PR
docs-sync?                # required only if existing docs became stale
PR-update?                # required when body/evidence became stale
sync-base?                # explicit rebase when required
merge-check
merge

Key execution rules:

- code-init?: user-provided code becomes input to spec-write.
- spec-amend?: runs only when code-check changes requirements. Expanded requirements must loop through implementation again.
- docs?: creates or intentionally expands docs only when requested.
- PR-write: describes the complete diff, regardless of who wrote each change.
- code-fix: receives only user-approved findings.
- spec-update: mandatory final reconciliation with the authoritative PR.
- docs-sync?: runs whenever current repository docs became stale, even if the earlier optional docs stage was skipped.
- PR-update?: required when final Spec, docs, verification evidence, behavior, or scope changes make the PR body stale.
- merge-check: must inspect the latest post-fix diff, not merely reuse the earlier review result.

### merge-check:

Verifies
- The PR targets stage and is open, mergeable, and conflict-free.
- All intended commits are pushed; no relevant work remains only locally.
- The latest code-bearing diff was reviewed.
- Every user-approved finding was fixed and verified.
- Rejected findings require no further action.
- Required tests and PR checks passed for the current code.
- The final Spec and QA cases accurately represent the PR.
- Existing docs are not stale relative to PR behavior.
- The PR title, body, scope, and verification evidence are current.
- No unauthorized database or unrelated changes entered the PR.

Later derived-artifact commits are acceptable without another full code review when they modify only the Spec or docs and accurately follow already-reviewed code. Any later implementation or test change invalidates the previous review and routes back
to code-review.

On failure, merge-check routes to the responsible stage:

Unreviewed code       → code-review
Approved finding open → code-fix
Stale Spec/QA         → spec-update
Stale docs            → docs-sync
Stale PR body         → PR-update
Failing checks        → code-fix or code-write
Merge conflict        → conflict resolution

So its output should be binary:

`Ready to merge.`

or a precise list of blocking conditions.

---

### Blocking issues

- P1 — PR publication and head identity are missing. PR-write should explicitly mean push commits, create/update the PR, and record the reviewed head SHA. Checks and reviews must apply to the current PR head, not merely have passed sometime earlier.
- P1 — Base synchronization can invalidate review. A rebase, merge from stage, or conflict resolution can change the effective diff even without an obvious feature edit. After sync-base, compare the resulting diff against the reviewed diff. Any code
  or test difference routes back to code-review; conflict-resolution edits always do.

- P1 — spec-update could accidentally legitimize missing behavior. Reconciliation must not silently remove an approved requirement simply because the implementation omitted it. Material differences from the reviewed pre-code Spec need an explicit
  user-approved scope decision or must route back to implementation and review.

- P1 — Undecided review findings are not covered. Merge-check handles approved and rejected findings, but not findings awaiting a decision. Every finding should have a durable disposition: approved/fixed, rejected, or superseded. Undecided findings
  block merging.

- P2 — Post-review implementation loops need to be explicit. Any code or test change from code-fix, failing checks, conflict resolution, or late refinement must route through verification and then code-review again.
- P2 — code-check needs two distinct routes.
    - Behavior does not meet the existing Spec → code-write
    - Requirements changed or expanded → spec-amend → code-write

  Both routes should return through verification before PR creation or update.

### Recommended tightened flow

idea
code-init?
spec-write
spec-review ────────────────┐
code-write                  │
code-check?                 │
existing requirement gap ─┘
requirement change → spec-amend → spec-review → code-write

docs?
verify
push / PR-write
code-review
finding-disposition
code-fix? → verify → push → code-review

spec-update
docs-sync?
PR-update?
sync-base?
effective code/test diff changed → verify → code-review

merge-check
merge

### Tightened merge-check contract

In addition to your existing checks, require:

- PR head SHA equals the pushed local branch head.
- Review and required checks apply to that exact head or an accepted later derived-artifact-only commit.
- No finding remains without an explicit disposition.
- Final Spec changes did not silently remove previously approved requirements.
- Any post-review code or test change triggered a new review.
- Base synchronization did not change the reviewed effective code/test diff.
- Spec/docs-only commits received focused artifact validation and did not alter executable files.
- The PR is not a draft unless the merge process intentionally supports that state.

The binary output is right:

Ready to merge.

Otherwise, return only precise blockers with their routing stage, ideally including the affected file/check/finding and current PR head SHA.
