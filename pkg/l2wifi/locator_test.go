package l2wifi

import (
	"testing"
)

func TestScanLocalWiFiNetworks(t *testing.T) {
	nets, err := ScanLocalWiFiNetworks()
	if err != nil {
		t.Logf("ScanLocalWiFiNetworks info: %v", err)
	} else {
		t.Logf("ScanLocalWiFiNetworks found %d networks", len(nets))
	}
}

func TestNormalizeBSSID(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"f4:69:42:6a:ae:a0", "f4:69:42:6a:ae:a0"},
		{"F4-69-42-6A-AE-A0", "f4:69:42:6a:ae:a0"},
		{"f469426aaea0", "f4:69:42:6a:ae:a0"},
		{"1:2:3:4:5:6", "01:02:03:04:05:06"},
	}

	for _, c := range cases {
		res := normalizeBSSID(c.input)
		if res != c.expected {
			t.Errorf("normalizeBSSID(%q) = %q; want %q", c.input, res, c.expected)
		}
	}
}
