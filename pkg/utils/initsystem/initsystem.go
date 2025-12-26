package initsystem

import (
	"fmt"
	"os/exec"
	"strings"
)

type InitSystem interface {
	EnableCommand(service string) string

	ServiceStart(service string) error

	ServiceStop(service string) error

	ServiceRestart(service string) error

	ServiceExists(service string) bool

	ServiceIsActive(service string) bool
}

type SystemdInitSystem struct{}

func (s SystemdInitSystem) EnableCommand(service string) string {
	return fmt.Sprintf("systemctl enable %s.service", service)
}

func (s SystemdInitSystem) reloadSystemd() error {
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("Failed to reload systemd: %w", err)
	}
	return nil
}

func (s SystemdInitSystem) ServiceStart(service string) error {
	if err := s.reloadSystemd(); err != nil {
		return err
	}
	args := []string{"start", service}
	return exec.Command("systemctl", args...).Run()
}

func (s SystemdInitSystem) ServiceStop(service string) error {
	args := []string{"stop", service}
	return exec.Command("systemctl", args...).Run()
}

func (s SystemdInitSystem) ServiceRestart(service string) error {
	if err := s.reloadSystemd(); err != nil {
		return err
	}
	args := []string{"restart", service}
	return exec.Command("systemctl", args...).Run()
}

func (s SystemdInitSystem) ServiceExists(service string) bool {
	args := []string{"status", service}
	bytes, _ := exec.Command("systemctl", args...).Output()
	output := string(bytes)
	return !strings.Contains(output, "Loaded: not-found") && !strings.Contains(output, "could not be found")
}

func (s SystemdInitSystem) ServiceIsActive(service string) bool {
	args := []string{"is-active", service}
	bytes, _ := exec.Command("systemctl", args...).Output()
	output := string(bytes)
	if strings.TrimSpace(output) == "active" {
		return true
	}
	return false
}
