package config

import "testing"

func TestVODProfile(t *testing.T) {
	if VOD.SegmentDuration != 5 {
		t.Errorf("expected VOD segment duration to be 5, got %d", VOD.SegmentDuration)
	}
	if VOD.LowLatency != false {
		t.Errorf("expected VOD low latency to be false, got %v", VOD.LowLatency)
	}
}

func TestLIVEProfile(t *testing.T) {
	if LIVE.SegmentDuration != 2 {
		t.Errorf("expected LIVE segment duration to be 2, got %d", LIVE.SegmentDuration)
	}
	if LIVE.LowLatency != true {
		t.Errorf("expected LIVE low latency to be true, got %v", LIVE.LowLatency)
	}
}
