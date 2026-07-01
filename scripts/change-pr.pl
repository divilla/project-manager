#!/usr/bin/env perl
use strict;
use warnings;
use File::Path qw(make_path);

sub fail {
    my ($message) = @_;
    die "$message\n";
}

@ARGV == 0 or fail("usage: scripts/change-pr.pl");

my $branch = qx{git branch --show-current};
chomp $branch;

my $change_name = $branch;
$change_name =~ s/^changes\///;

$branch eq "changes/$change_name"
    or fail("current branch is not a changes/<change-name> branch: $branch");
$change_name =~ /\A[A-Za-z0-9][A-Za-z0-9._-]*\z/
    or fail("invalid change name from current branch: $change_name");

run_checked(qw(git add -A));
run_checked("git", "commit", "-m", "Write PR for $change_name by agent");
run_checked("git", "push", "origin", "changes/$change_name");

ensure_stage_is_ancestor();

my $pr_body_file = "agent/prs/$change_name.md";
my $pr_title = extract_pr_title($pr_body_file);
my $pr_url_file = "agent/prurls/$change_name";

make_path("agent/prurls");
my $pr_url = run_capture_checked(
    "gh",
    "pr",
    "create",
    "--base",
    "stage",
    "--head",
    "changes/$change_name",
    "--title",
    $pr_title,
    "--body-file",
    $pr_body_file,
);
write_file($pr_url_file, $pr_url);

run_checked(qw(git add -A));
run_checked("git", "commit", "-m", "Write PR URL for $change_name by agent");
run_checked("git", "push", "origin", "changes/$change_name");

sub extract_pr_title {
    my ($path) = @_;
    open my $fh, "<", $path or fail("cannot read PR body file $path: $!");
    my $first_line = <$fh>;
    defined $first_line or fail("PR body file is empty: $path");
    chomp $first_line;
    $first_line =~ s/\r\z//;
    $first_line =~ /\A#\s+(.+?)\s*\z/
        or fail("first line of $path must be '# <Title>'");
    return $1;
}

sub write_file {
    my ($path, $content) = @_;
    open my $fh, ">", $path or fail("cannot write $path: $!");
    print {$fh} $content or fail("cannot write $path: $!");
    close $fh or fail("cannot close $path: $!");
}

sub run_capture_checked {
    my (@command) = @_;
    open my $fh, "-|", @command
        or fail(join(" ", @command) . " failed to start: $!");
    local $/;
    my $output = <$fh>;
    close $fh or fail(join(" ", @command) . " failed");
    return $output;
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
