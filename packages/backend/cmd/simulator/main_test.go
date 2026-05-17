package main

import "testing"

func TestDecideLaunchMode_DefaultsToServer(t *testing.T) {
	mode := decideLaunchMode(false, "", "")
	if mode != launchModeServer {
		t.Fatalf("expected server mode by default, got %q", mode)
	}
}

func TestDecideLaunchMode_UsesCLIWhenDeviceSpecified(t *testing.T) {
	mode := decideLaunchMode(false, "SENSOR-001", "")
	if mode != launchModeCLI {
		t.Fatalf("expected cli mode when sensor device specified, got %q", mode)
	}

	mode = decideLaunchMode(false, "", "ACT-001")
	if mode != launchModeCLI {
		t.Fatalf("expected cli mode when actuator device specified, got %q", mode)
	}
}

func TestDecideLaunchMode_ServerFlagStillWins(t *testing.T) {
	mode := decideLaunchMode(true, "SENSOR-001", "ACT-001")
	if mode != launchModeServer {
		t.Fatalf("expected server mode when --server is set, got %q", mode)
	}
}
