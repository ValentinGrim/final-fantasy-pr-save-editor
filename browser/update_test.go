package browser

import "testing"

func TestNewComparable(t *testing.T) {
	tests := []struct {
		in    string
		want  comp
		found bool
	}{
		{"1.4.0", comp{version: [3]int{1, 4, 0}}, true},
		{"v1.4.1", comp{version: [3]int{1, 4, 1}}, true},
		{"v1.10.0", comp{version: [3]int{1, 10, 0}}, true},
		{"1.4", comp{version: [3]int{1, 4, 0}}, true},
		{"1.4.0-rc1", comp{version: [3]int{1, 4, 0}, pre: true}, true},
		{"v1.5.0-beta.2", comp{version: [3]int{1, 5, 0}, pre: true}, true},
		{"1.4.0+build.7", comp{version: [3]int{1, 4, 0}}, true},
		{"1.4.0.1", comp{}, false},
		{"v1.4.0_linux", comp{}, false},
		{"1", comp{}, false},
		{"", comp{}, false},
		{"latest", comp{}, false},
		{"1.x.0", comp{}, false},
		{"1..0", comp{}, false},
	}
	for _, tt := range tests {
		got, found := newComparable(tt.in)
		if found != tt.found || got != tt.want {
			t.Errorf("newComparable(%q) = %+v, %v; want %+v, %v", tt.in, got, found, tt.want, tt.found)
		}
	}
}

func TestIsNewer(t *testing.T) {
	tests := []struct {
		current, other string
		want           bool
	}{
		{"1.4.0", "v1.4.1", true},
		{"1.4.1", "1.4.0", false},
		{"1.4.0", "1.4.0", false},
		{"1.9.9", "v1.10.0", true},
		{"v1.10.0", "1.9.9", false},
		{"1.4.0", "2.0.0", true},
		{"1.4.0", "1.4.0-rc1", false}, // a prerelease never beats its release
		{"1.4.0-rc1", "1.4.0", true},
		{"1.4.0-rc1", "1.4.0-rc2", false},
		{"1.4.0", "1.4.0+build.7", false},
	}
	for _, tt := range tests {
		c, ok := newComparable(tt.current)
		if !ok {
			t.Fatalf("newComparable(%q) not comparable", tt.current)
		}
		o, ok := newComparable(tt.other)
		if !ok {
			t.Fatalf("newComparable(%q) not comparable", tt.other)
		}
		if got := IsNewer(c, o); got != tt.want {
			t.Errorf("IsNewer(%q, %q) = %v; want %v", tt.current, tt.other, got, tt.want)
		}
	}
}
