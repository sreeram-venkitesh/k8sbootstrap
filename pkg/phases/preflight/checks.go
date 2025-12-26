package preflight

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sreeram-venkitesh/k8sbootstrap/pkg/utils/initsystem"
)

func CheckRoot() (errorList []error) {
	if os.Getuid() != 0 {
		return []error{fmt.Errorf("User is not running as root")}
	}
	return nil
}

func CheckSwap() (errorList []error) {
	bytes, err := exec.Command("swapoff", "-a").Output()
	if err != nil {
		return []error{err}
	}

	if len(bytes) > 0 {
		return []error{fmt.Errorf("swap is enabled, please disable it")}
	}
	return nil
}

func CheckPorts() (errorList []error) {
	requiredPorts := []string{"6443", "2379", "2380", "10250", "10251", "10252"}

	for _, port := range requiredPorts {
		cmd := exec.Command("sh", "-c", fmt.Sprintf("netstat -tuln | grep ':%s ' || ss -tuln | grep ':%s '", port, port))
		output, _ := cmd.Output()

		if len(output) > 0 {
			return []error{fmt.Errorf("Port %s is already in use", port)}
		}
	}
	return nil
}

func CheckKernelModules() (errorList []error) {
	modules := []string{"br_netfilter", "overlay"}

	for _, module := range modules {
		cmd := exec.Command("lsmod")
		output, err := cmd.Output()
		if err != nil {
			return []error{err}
		}

		if !strings.Contains(string(output), module) {
			return []error{fmt.Errorf("kernel module %s is not loaded", module)}
		}
	}
	return nil
}

func CheckContainerRuntime() (errorList []error) {
	initSystem := initsystem.SystemdInitSystem{}
	isActive := initSystem.ServiceIsActive("containerd")
	if !isActive {
		return []error{fmt.Errorf("containerd is not running")}
	}
	return nil
}

func CheckIPForwarding() (errorList []error) {
	output, err := os.ReadFile("/proc/sys/net/ipv4/ip_forward")
	if err != nil {
		return []error{err}
	}

	if strings.TrimSpace(string(output)) != "1" {
		return []error{fmt.Errorf("IP forwarding is not enabled")}
	}
	return nil
}

func CheckKubelet() (errorList []error) {
	if err := ServiceCheck("kubelet"); err != nil {
		return err
	}
	return nil
}

func ServiceCheck(service string) (errorList []error) {
	// TODO instead of hardcoding systemd, implement a GetInitSystem method
	initSystem := initsystem.SystemdInitSystem{}

	if !initSystem.ServiceExists(service) {
		return []error{fmt.Errorf("%s service does not exist", service)}
	}

	if !initSystem.ServiceIsActive(service) {
		errorList = append(errorList,
			fmt.Errorf("%s service is not active, please run 'systemctl start %s.service'", service, service))
	}

	return errorList
}
