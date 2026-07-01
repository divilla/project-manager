#!/usr/bin/env perl
use strict;
use warnings;

sub fail {
    my ($message) = @_;
    die "$message\n";
}

@ARGV == 1 or fail("usage: scripts/change-new.pl <change-name>");

my $change_name = $ARGV[0];
$change_name =~ /\A[A-Za-z0-9][A-Za-z0-9._-]*\z/
    or fail("invalid change name: $change_name");

my $branch = "changes/$change_name";
my $checkout_output = run_capture("git", "checkout", $branch);
my $checkout_status = $?;

if ($checkout_status != 0) {
    my $stage_output = run_capture(qw(git checkout stage));
    my $stage_status = $?;
    $stage_status == 0
        or fail("git checkout stage failed:\n$stage_output");
    index($stage_output, "Your branch is up to date with 'origin/stage'.") >= 0
        or fail("git checkout stage did not report an up-to-date origin/stage branch:\n$stage_output");

    run_checked("git", "checkout", "-b", $branch);
} elsif (length $checkout_output) {
    print $checkout_output;
}

my $change_path = "agent/changes/$change_name.md";
run_checked("touch", $change_path);

sub run_capture {
    my (@command) = @_;
    return qx{@command 2>&1};
}

sub run_checked {
    my (@command) = @_;
    my $output = run_capture(@command);
    $? == 0 or fail(join(" ", @command) . " failed:\n$output");
    print $output if length $output;
}
