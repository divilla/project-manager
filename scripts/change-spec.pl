#!/usr/bin/env perl
use strict;
use warnings;

sub fail {
    my ($message) = @_;
    die "$message\n";
}

@ARGV == 0 or fail("usage: scripts/change-spec.pl");

my $branch = qx{git branch --show-current};
chomp $branch;

my ($change_name) = $branch =~ m{^change/([0-9]+-[0-9A-Za-z_-]+)$}
    or fail("current branch is not a change/<change-slug> branch: $branch");

run_checked(qw(git add -A));
run_checked("git", "commit", "-m", "Spec for $change_name by agent");
run_checked("git", "push", "origin", "change/$change_name");

sub run_checked {
    my (@command) = @_;
    system @command;
    $? == 0 or fail(join(" ", @command) . " failed");
}
