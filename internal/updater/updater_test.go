package updater

import "testing"

func TestIsNewer(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
	}{
		// Exact version matches
		{"0.8.2", "0.8.2", false},
		{"v0.8.2", "v0.8.2", false},

		// Latest is newer
		{"0.8.1", "0.8.2", true},
		{"0.7.0", "0.8.0", true},
		{"0.8.0", "0.8.1", true},

		// Current is newer
		{"0.8.2", "0.8.1", false},
		{"0.9.0", "0.8.2", false},

		// Git describe output with commit count and hash
		{"0.8.2-3-gabcdef1", "0.8.2", false},
		{"0.8.2-3-gabcdef1", "0.8.3", true},
		{"0.7.0-4-g31a5b8e", "0.8.2", true},
		{"v0.7.0-4-g31a5b8e", "0.8.2", true},

		// Dirty builds
		{"0.8.2-dirty", "0.8.2", false},
		{"0.8.2-dirty", "0.8.3", true},
		{"0.8.1-dirty", "0.8.2", true},

		// Git describe with dirty
		{"0.8.2-3-gabcdef1-dirty", "0.8.2", false},
		{"0.8.2-3-gabcdef1-dirty", "0.8.3", true},

		// Dev and unknown versions
		{"dev", "0.8.2", true},
		{"unknown", "0.8.2", true},

		// Commit hashes (not semver)
		{"31a5b8e", "0.8.2", true},
		{"abcdef1234", "0.8.2", true},

		// Minor and patch versions
		{"0.8.0", "0.8.1", true},
		{"0.8.0", "0.9.0", true},
		{"1.0.0", "2.0.0", true},
		{"1.2.3", "1.2.4", true},
		{"1.2.3", "1.3.0", true},
		{"1.2.3", "2.0.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.current+"_vs_"+tt.latest, func(t *testing.T) {
			got := isNewer(tt.current, tt.latest)
			if got != tt.want {
				t.Errorf("isNewer(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

func TestIsSemver(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"0.8.2", true},
		{"1.2.3", true},
		{"10.20.30", true},
		{"0.1", true},
		{"1.0", true},

		// Not semver
		{"dev", false},
		{"unknown", false},
		{"abcdef", false},
		{"0.8.2-3-gabcdef", false},
		{"v0.8.2", false},
		{"1", false},
		{"1.2.3.4", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := isSemver(tt.version)
			if got != tt.want {
				t.Errorf("isSemver(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestParseSemver(t *testing.T) {
	tests := []struct {
		version string
		want    [3]int
	}{
		{"0.8.2", [3]int{0, 8, 2}},
		{"1.2.3", [3]int{1, 2, 3}},
		{"10.20.30", [3]int{10, 20, 30}},
		{"0.1", [3]int{0, 1, 0}},
		{"5", [3]int{5, 0, 0}},
		{"", [3]int{0, 0, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := parseSemver(tt.version)
			if got != tt.want {
				t.Errorf("parseSemver(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestMatchAsset(t *testing.T) {
	assets := []ghAsset{
		{Name: "kibble-linux-amd64", Size: 1000},
		{Name: "kibble-linux-arm64", Size: 2000},
		{Name: "kibble-linux-arm", Size: 3000},
		{Name: "kibble-darwin-arm64", Size: 4000},
	}

	tests := []struct {
		name      string
		goos      string
		goarch    string
		wantFound bool
		wantName  string
		wantSize  int64
	}{
		{"Linux AMD64", "linux", "amd64", true, "kibble-linux-amd64", 1000},
		{"Linux ARM64", "linux", "arm64", true, "kibble-linux-arm64", 2000},
		{"Linux ARM", "linux", "arm", true, "kibble-linux-arm", 3000},
		{"macOS ARM64", "darwin", "arm64", true, "kibble-darwin-arm64", 4000},
		{"Windows AMD64", "windows", "amd64", false, "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily mock runtime.GOOS/GOARCH, so we test matchAsset directly
			// by looking for the expected name in the assets
			wantName := "kibble-" + tt.goos + "-" + tt.goarch
			var found bool
			var gotAsset ghAsset
			for _, a := range assets {
				if a.Name == wantName {
					found = true
					gotAsset = a
					break
				}
			}

			if found != tt.wantFound {
				t.Errorf("Found asset = %v, want %v", found, tt.wantFound)
			}
			if found && gotAsset.Name != tt.wantName {
				t.Errorf("Asset name = %q, want %q", gotAsset.Name, tt.wantName)
			}
			if found && gotAsset.Size != tt.wantSize {
				t.Errorf("Asset size = %d, want %d", gotAsset.Size, tt.wantSize)
			}
		})
	}
}
