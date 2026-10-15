package models

import "testing"

func TestNormalizeProductName(t *testing.T) {
	for in, want := range map[string]string{
		"  ค่า  ขนส่ง ": "ค่า ขนส่ง",
		"ค่าแรง":        "ค่าแรง",
		"\tน้ำแข็ง\n":   "น้ำแข็ง",
		"   ":           "",
	} {
		if got := NormalizeProductName(in); got != want {
			t.Errorf("NormalizeProductName(%q) = %q, want %q", in, got, want)
		}
	}
}
