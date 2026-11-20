//go:build !darwin && !linux && !windows

package sys

import (
	"fmt"
	"runtime"
)

type otherBackend struct{}

func newPlatformBackend(runner CommandRunner, fs HostFS) PlatformBackend {
	return &otherBackend{}
}

func (b *otherBackend) Name() string {
	return runtime.GOOS
}

func (b *otherBackend) Inspect() (FingerprintStatus, error) {
	return FingerprintStatus{
		OS:           runtime.GOOS,
		Supported:    false,
		Configurable: false,
		Limitations: []string{
			fmt.Sprintf("Fingerprint terminal authentication setup is not supported on OS %q (supported targets: darwin, linux, windows).", runtime.GOOS),
		},
	}, nil
}

func (b *otherBackend) PlanEnable() (ActionPlan, error) {
	return ActionPlan{}, fmt.Errorf("fingerprint setup is not supported on %s", runtime.GOOS)
}

func (b *otherBackend) PlanDisable() (ActionPlan, error) {
	return ActionPlan{}, fmt.Errorf("fingerprint setup is not supported on %s", runtime.GOOS)
}

func (b *otherBackend) Verify() error {
	return fmt.Errorf("fingerprint verification is not supported on %s", runtime.GOOS)
}
