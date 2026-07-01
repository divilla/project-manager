#!/usr/bin/env perl
use strict;
use warnings;

sub fail {
    my ($message) = @_;
    die "$message\n";
}

@ARGV == 0 or fail("usage: scripts/change-master.pl");

my $branch = trim(run_capture_checked(qw(git branch --show-current)));
$branch eq "stage"
    or fail("Please checkout stage branch.");

ensure_clean_worktree();
run_checked(qw(git fetch origin));

run_checked("git", "checkout", "stage");
run_checked("git", "pull", "--ff-only", "origin", "stage");
my $stage_commit = trim(run_capture_checked(qw(git rev-parse HEAD)));
my $origin_stage_commit = remote_head_commit("stage");
$stage_commit eq $origin_stage_commit
    or fail("stage HEAD $stage_commit does not match origin/stage $origin_stage_commit");

my $origin_master_commit = remote_head_commit("master");
my $master_commit = local_branch_commit("master");
if (defined $master_commit) {
    ensure_commit_is_ancestor(
        $master_commit,
        $origin_master_commit,
        "Local master contains commits that are not on origin/master. Refusing to promote until master is reconciled.",
    );
}

ensure_master_is_ancestor($origin_master_commit, $stage_commit);
ensure_stage_is_current($stage_commit);
run_checked(
    "git",
    "push",
    "--atomic",
    "--force-with-lease=refs/heads/master:$origin_master_commit",
    "--force-with-lease=refs/heads/stage:$stage_commit",
    "origin",
    "$stage_commit:refs/heads/master",
    "$stage_commit:refs/heads/stage",
);
ensure_stage_is_current($stage_commit);

my $new_origin_master_commit = remote_head_commit("master");
$new_origin_master_commit eq $stage_commit
    or fail("origin/master $new_origin_master_commit does not match stage commit $stage_commit");
run_checked("git", "fetch", "origin", "refs/heads/master:refs/heads/master");

my $new_master_commit = local_branch_commit("master");
defined $new_master_commit && $new_master_commit eq $stage_commit
    or fail("local master does not match promoted commit $stage_commit");
my $current_branch = trim(run_capture_checked(qw(git branch --show-current)));
$current_branch eq "stage"
    or fail("current branch is $current_branch after promotion; expected stage");
print "Promoted master to $stage_commit; current branch is stage.\n";

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

sub local_branch_commit {
    my ($branch_name) = @_;
    system("git", "show-ref", "--verify", "--quiet", "refs/heads/$branch_name");
    return undef if $? != 0;
    return trim(run_capture_checked("git", "rev-parse", "refs/heads/$branch_name"));
}

sub run_checked {
    my (@command) = @_;
    system @command;
    $? == 0 or fail(join(" ", @command) . " failed");
}

sub ensure_stage_is_current {
    my ($stage_commit) = @_;
    my $origin_stage_commit = remote_head_commit("stage");
    $origin_stage_commit eq $stage_commit
        or fail("origin/stage moved from $stage_commit to $origin_stage_commit");
}

sub ensure_master_is_ancestor {
    my ($master_commit, $stage_commit) = @_;
    ensure_commit_is_ancestor(
        $master_commit,
        $stage_commit,
        "cannot fast-forward master to stage: master is not an ancestor of $stage_commit",
    );
}

sub ensure_commit_is_ancestor {
    my ($ancestor_commit, $descendant_commit, $message) = @_;
    system("git", "merge-base", "--is-ancestor", $ancestor_commit, $descendant_commit);
    return if $? == 0;
    fail($message);
}
