#!/usr/bin/env perl
use strict;
use warnings;

sub fail {
    my ($message) = @_;
    die "$message\n";
}

@ARGV == 0 or fail("usage: scripts/change-docs.pl");

my $branch = qx{git branch --show-current};
chomp $branch;

my $change_name = $branch;
$change_name =~ s/^changes\///;

$branch eq "changes/$change_name"
    or fail("current branch is not a changes/<change-name> branch: $branch");
$change_name =~ /\A[A-Za-z0-9][A-Za-z0-9._-]*\z/
    or fail("invalid change name from current branch: $change_name");

run_checked(qw(git add -A));
run_checked("git", "commit", "-m", "Write docs for $change_name by agent");
run_checked("git", "push", "origin", "changes/$change_name");

sub run_checked {
    my (@command) = @_;
    system @command;
    $? == 0 or fail(join(" ", @command) . " failed");
}
