#!/usr/bin/env perl
use strict;
use warnings;

sub fail {
    my ($message) = @_;
    die "$message\n";
}

my $dry_run = 0;
if (@ARGV == 1 && $ARGV[0] eq "--dry-run") {
    $dry_run = 1;
} elsif (@ARGV != 0) {
    fail("usage: scripts/rename-branches.pl [--dry-run]");
}

run_capture_checked(qw(git rev-parse --git-dir));

my @local_branches = local_branches_with_prefix("changes/");
my @remotes = remotes();
my %remote_branches_by_remote;
my %remote_targets_by_remote;

for my $remote (@remotes) {
    $remote_branches_by_remote{$remote} = remote_branches_with_prefix($remote, "changes/");
    $remote_targets_by_remote{$remote} = remote_branches_with_prefix($remote, "change/");
}

validate_local_targets(@local_branches);
validate_remote_targets(\%remote_branches_by_remote, \%remote_targets_by_remote);

if (!@local_branches && !remote_rename_count(\%remote_branches_by_remote)) {
    print "No local or remote branches start with changes/.\n";
    exit 0;
}

for my $remote (@remotes) {
    my $branches = $remote_branches_by_remote{$remote};
    next if !%{$branches};

    my @command = ("git", "push", "--atomic");
    for my $old_branch (sort keys %{$branches}) {
        push @command, "--force-with-lease=refs/heads/$old_branch:$branches->{$old_branch}";
    }
    push @command, $remote;
    for my $old_branch (sort keys %{$branches}) {
        my $new_branch = renamed_branch($old_branch);
        push @command, "$branches->{$old_branch}:refs/heads/$new_branch";
        push @command, ":refs/heads/$old_branch";
    }

    if ($dry_run) {
        print dry_run_command(@command), "\n";
    } else {
        print "Renaming remote branches on $remote.\n";
        run_checked(@command);
    }
}

for my $old_branch (@local_branches) {
    my $new_branch = renamed_branch($old_branch);
    my $remote = trim(run_capture("git", "config", "--get", "branch.$old_branch.remote"));
    my $merge = trim(run_capture("git", "config", "--get", "branch.$old_branch.merge"));

    if ($dry_run) {
        print dry_run_command("git", "branch", "-m", $old_branch, $new_branch), "\n";
        if ($merge eq "refs/heads/$old_branch") {
            print dry_run_command("git", "config", "branch.$new_branch.merge", "refs/heads/$new_branch"), "\n";
        }
        next;
    }

    print "Renaming local branch $old_branch to $new_branch.\n";
    run_checked("git", "branch", "-m", $old_branch, $new_branch);
    if ($remote ne "" && $merge eq "refs/heads/$old_branch") {
        run_checked("git", "config", "branch.$new_branch.merge", "refs/heads/$new_branch");
    }
}

if (!$dry_run) {
    for my $remote (@remotes) {
        next if !%{$remote_branches_by_remote{$remote}};
        run_checked("git", "fetch", "--prune", $remote);
    }
}

sub local_branches_with_prefix {
    my ($prefix) = @_;
    my $output = run_capture_checked("git", "for-each-ref", "--format=%(refname:strip=2)", "refs/heads/$prefix");
    my @branches = grep { $_ ne "" } split /\n/, $output;
    return sort @branches;
}

sub remotes {
    my $output = run_capture_checked(qw(git remote));
    my @remotes = grep { $_ ne "" } split /\n/, $output;
    return sort @remotes;
}

sub remote_branches_with_prefix {
    my ($remote, $prefix) = @_;
    my $output = run_capture_checked("git", "ls-remote", "--heads", $remote, "$prefix*");
    my %branches;
    for my $line (grep { $_ ne "" } split /\n/, $output) {
        $line =~ /\A([0-9a-f]+)\s+refs\/heads\/(.+)\z/
            or fail("cannot parse remote branch line from $remote: $line");
        $branches{$2} = $1;
    }
    return \%branches;
}

sub validate_local_targets {
    my (@branches) = @_;
    for my $old_branch (@branches) {
        my $new_branch = renamed_branch($old_branch);
        system("git", "show-ref", "--verify", "--quiet", "refs/heads/$new_branch");
        $? != 0
            or fail("local target branch already exists: $new_branch");
    }
}

sub validate_remote_targets {
    my ($sources_by_remote, $targets_by_remote) = @_;
    for my $remote (sort keys %{$sources_by_remote}) {
        my $sources = $sources_by_remote->{$remote};
        my $targets = $targets_by_remote->{$remote};
        for my $old_branch (sort keys %{$sources}) {
            my $new_branch = renamed_branch($old_branch);
            exists $targets->{$new_branch}
                and fail("remote target branch already exists on $remote: $new_branch");
        }
    }
}

sub remote_rename_count {
    my ($branches_by_remote) = @_;
    my $count = 0;
    for my $remote (keys %{$branches_by_remote}) {
        $count += scalar keys %{$branches_by_remote->{$remote}};
    }
    return $count;
}

sub renamed_branch {
    my ($branch) = @_;
    $branch =~ s/\Achanges\//change\//;
    return $branch;
}

sub run_capture {
    my (@command) = @_;
    open my $fh, "-|", @command
        or fail(join(" ", @command) . " failed to start: $!");
    local $/;
    my $output = <$fh>;
    $output = "" if !defined $output;
    close $fh;
    return $output;
}

sub run_capture_checked {
    my (@command) = @_;
    my $output = run_capture(@command);
    $? == 0 or fail(join(" ", @command) . " failed");
    return $output;
}

sub run_checked {
    my (@command) = @_;
    system @command;
    $? == 0 or fail(join(" ", @command) . " failed");
}

sub trim {
    my ($value) = @_;
    $value =~ s/\A\s+//;
    $value =~ s/\s+\z//;
    return $value;
}

sub dry_run_command {
    my (@command) = @_;
    return "dry-run: " . join(" ", map { shell_quote($_) } @command);
}

sub shell_quote {
    my ($value) = @_;
    return "''" if $value eq "";
    return $value if $value =~ /\A[A-Za-z0-9_+=:,.\/-]+\z/;
    $value =~ s/'/'"'"'/g;
    return "'$value'";
}
