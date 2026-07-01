#!/usr/bin/env perl
use strict;
use warnings;
use File::Path qw(make_path);
use IPC::Open2 qw(open2);

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

#run_checked(qw(git add -A));
#run_checked("git", "commit", "-m", "Write PR for $change_name by agent");
#run_checked("git", "push", "origin", "changes/$change_name");

ensure_stage_is_ancestor();

my $change_file = "agent/changes/$change_name.md";
my $base = run_capture_checked(qw(git merge-base HEAD stage));
chomp $base;

my $pr_body = draft_pr_body($change_file, $base);
my $pr_title = extract_pr_title_from_content($pr_body, "generated PR body");

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
    "--body",
    $pr_body,
);
write_file($pr_url_file, $pr_url);

#run_checked(qw(git add -A));
#run_checked("git", "commit", "-m", "Write PR URL for $change_name by agent");
#run_checked("git", "push", "origin", "changes/$change_name");

sub draft_pr_body {
    my ($change_file, $base) = @_;

    my $change_content = read_file($change_file);
    my $diff_stat = run_capture_checked("git", "diff", "--stat", "$base..HEAD");
    my $diff = run_capture_checked("git", "diff", "--find-renames", "$base..HEAD");

    my $input = join(
        "",
        "Draft the final PR body for this branch. Do not create the PR.\n",
        "Use only the stdin-provided Change file, diff stat, and diff. Do not inspect files, run commands, edit files, commit, push, or call GitHub.\n",
        "Treat the Change file as the intended contract and the diff as the source of truth for what is actually in the branch.\n",
        "If the Change file and diff materially conflict, output only a short conflict report instead of a PR body.\n",
        "The first line must be \"# <Title>\" where <Title> exactly matches the Change title. Follow it with exactly one blank line.\n",
        "Keep the body concise, reviewer-focused, and specific. Mention only externally observable behavior, API/data contracts, database/seed changes, frontend/CLI/docs changes, and verification evidence that are present in the provided input.\n",
        "Do not include filler, implementation diary, generic praise, speculation, or verification claims not supported by the provided input.\n",
        "Output only markdown for the PR body.\n",
        "\n<change-file>\n",
        $change_content,
        "</change-file>\n\n<diff-stat>\n",
        $diff_stat,
        "</diff-stat>\n\n<diff>\n",
        $diff,
        "</diff>\n",
    );

    my $pid = open2(
        my $codex_out,
        my $codex_in,
        "codex",
        "exec",
        "-C",
        "/home/vito/go/src/project-manager",
        "--sandbox",
        "read-only",
        "--ephemeral",
        "-",
    );

    print {$codex_in} $input or fail("cannot write PR draft prompt to codex: $!");
    close $codex_in or fail("cannot close codex stdin: $!");

    local $/;
    my $body = <$codex_out>;
    close $codex_out or fail("cannot read codex output: $!");

    waitpid($pid, 0);
    $? == 0 or fail("codex exec failed");

    return $body;
}

sub extract_pr_title_from_content {
    my ($content, $source) = @_;
    my ($first_line) = split /\n/, $content, 2;
    defined $first_line or fail("PR body is empty: $source");
    $first_line =~ s/\r\z//;
    $first_line =~ /\A#\s+(.+?)\s*\z/
        or fail("first line of $source must be '# <Title>'");
    return $1;
}

sub read_file {
    my ($path) = @_;
    open my $fh, "<", $path or fail("cannot read $path: $!");
    local $/;
    my $content = <$fh>;
    close $fh or fail("cannot close $path: $!");
    return $content;
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
