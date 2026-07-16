//go:build !aix && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris

package flow

import "os/exec"

func configureExecProcessGroup(_ *exec.Cmd) {}
