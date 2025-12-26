package preflight

import "fmt"

func RunPreflightChecks() error {
	checks := []struct {
		name string
		fn   func() []error
	}{
		{"Checking if running as root", CheckRoot},
		{"Checking if swap is enabled", CheckSwap},
		{"Checking if ports are available", CheckPorts},
		{"Checking for kernel modules", CheckKernelModules},
		{"Checking container runtime", CheckContainerRuntime},
		{"Checking IP forwarding", CheckIPForwarding},
	}

	fmt.Println("[preflight] Running preflight checks")

	for _, check := range checks {
		fmt.Printf("[preflight] %s\n", check.name)
		if err := check.fn(); err != nil {
			return fmt.Errorf("%s failed: %s", check.name, err)
		}
	}

	fmt.Println("[preflight] All preflight checks passed!")
	return nil
}
