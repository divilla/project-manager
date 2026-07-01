#!/usr/bin/env perl
use strict;
use warnings;

sub fail {
    my ($message) = @_;
    die "$message\n";
}

@ARGV == 0 or fail("usage: scripts/change-stage.pl");

my $branch = qx{git branch --show-current};
chomp $branch;

my $change_name = $branch;
$change_name =~ s/^changes\///;

$branch eq "changes/$change_name"
    or fail("current branch is not a changes/<change-name> branch: $branch");
$change_name =~ /\A[A-Za-z0-9][A-Za-z0-9._-]*\z/
    or fail("invalid change name from current branch: $change_name");

my $change_branch = "changes/$change_name";

ensure_clean_worktree();
run_checked(qw(git fetch origin));
my $original_commit = trim(run_capture_checked(qw(git rev-parse HEAD)));
my $origin_change_commit = remote_head_commit($change_branch);
$original_commit eq $origin_change_commit
    or fail("local $change_branch $original_commit does not match origin/$change_branch $origin_change_commit");
ensure_pr_exists($change_branch);
ensure_stage_is_ancestor();

my $squashed_commit = $original_commit;
if (!is_squashed_change_commit($original_commit, $change_name)) {
    $squashed_commit = create_squash_commit($change_name);
    run_checked(
        "git",
        "push",
        "--force-with-lease=refs/heads/$change_branch:$origin_change_commit",
        "origin",
        "$squashed_commit:refs/heads/$change_branch",
    );
    run_checked("git", "update-ref", "refs/heads/$change_branch", $squashed_commit, $original_commit);
}

ensure_pr_exists($change_branch);

run_checked("git", "checkout", "stage");
run_checked("git", "pull", "--ff-only", "origin", "stage");
run_checked("git", "merge", "--ff-only", $change_branch);
my $stage_commit = trim(run_capture_checked(qw(git rev-parse HEAD)));
$stage_commit eq $squashed_commit
    or fail("stage HEAD $stage_commit does not match squashed change commit $squashed_commit");
run_checked("git", "push", "-u", "origin", "stage");
my $origin_stage_commit = remote_head_commit("stage");
$origin_stage_commit eq $squashed_commit
    or fail("origin/stage $origin_stage_commit does not match squashed change commit $squashed_commit");
run_checked("git", "push", "origin", "--delete", $change_branch);

sub ensure_clean_worktree {
    my $status = run_capture_checked(qw(git status --short));
    trim($status) eq "" or fail("uncommitted changes");
}

sub run_capture_checked {
    my (@command) = @_;
    open my $fh, "-|", @command
        or fail(join(" ", @command) . " failed to start: $!");
    local $/;
    my $output = <$fh>;
    $output = "" if !defined $output;
    close $fh or fail(join(" ", @command) . " failed");
    return $output;
}

sub trim {
    my ($value) = @_;
    $value =~ s/\A\s+//;
    $value =~ s/\s+\z//;
    return $value;
}

sub remote_head_commit {
    my ($branch_name) = @_;
    my $output = trim(run_capture_checked("git", "ls-remote", "--heads", "origin", $branch_name));
    $output =~ /\A([0-9a-f]{40})\s+refs\/heads\/\Q$branch_name\E\z/
        or fail("cannot verify origin/$branch_name");
    return $1;
}

sub is_squashed_change_commit {
    my ($commit, $change_name) = @_;
    my $parents = trim(run_capture_checked("git", "rev-list", "--parents", "-n", "1", $commit));
    my @fields = split /\s+/, $parents;
    shift @fields;
    return 0 if @fields != 1;

    my $stage_commit = trim(run_capture_checked(qw(git rev-parse origin/stage)));
    return 0 if $fields[0] ne $stage_commit;

    my $subject = trim(run_capture_checked("git", "log", "-1", "--format=%s", $commit));
    return $subject eq "Implement change $change_name";
}

sub create_squash_commit {
    my ($change_name) = @_;
    my $tree = trim(run_capture_checked(qw(git rev-parse HEAD^{tree})));
    return trim(run_capture_checked(
        "git",
        "commit-tree",
        $tree,
        "-p",
        "origin/stage",
        "-m",
        "Implement change $change_name",
    ));
}

sub ensure_pr_exists {
    my ($change_branch) = @_;
    my $state = trim(run_capture_checked(
        "gh",
        "pr",
        "view",
        $change_branch,
        "--json",
        "state",
        "--jq",
        ".state",
    ));
    $state eq "OPEN"
        or fail("PR for $change_branch is not open");
}

sub run_checked {
    my (@command) = @_;
    system @command;
    $? == 0 or fail(join(" ", @command) . " failed");
}

sub ensure_stage_is_ancestor {
    system(qw(git merge-base --is-ancestor origin/stage HEAD));
    return if $? == 0;
    fail("rebase needed: origin/stage is not an ancestor of HEAD");
}
