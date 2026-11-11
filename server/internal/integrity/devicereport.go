package integrity

import (
	"context"
	"slices"
)

const PurposeClientReport = "client_report"

var clientDeviceSignals = []string{"root", "test_keys", "emulator", "jailbreak"}

type DeviceReports struct {
	policy  Policy
	signals SignalSink
}

func NewDeviceReports(policy Policy, signals SignalSink) *DeviceReports {
	return &DeviceReports{policy: policy, signals: signals}
}

func (d *DeviceReports) Observe(ctx context.Context, account string, names []string) {
	known := knownDeviceSignals(names)
	if len(known) == 0 || d.policy.DeviceRiskAction == DeviceRiskWarn {
		return
	}
	d.signals.RecordIntegrity(ctx, account, PurposeClientReport, SignalDeviceRisk.String(), d.policy.DeviceRiskWeight, known)
}

func knownDeviceSignals(names []string) []string {
	var known []string
	for _, signal := range clientDeviceSignals {
		if slices.Contains(names, signal) {
			known = append(known, signal)
		}
	}
	return known
}
