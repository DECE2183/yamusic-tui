package config

import "testing"

func TestKeyContainsLayoutIndependent(t *testing.T) {
	k := NewKey("c")
	if !k.Contains("c") {
		t.Errorf("Latin 'c' should match")
	}
	if !k.Contains("с") { // Cyrillic es on the physical C key
		t.Errorf("Cyrillic 'с' (RU layout, physical C) should match binding 'c'")
	}
	if !k.Contains("С") { // uppercase Cyrillic
		t.Errorf("Cyrillic 'С' should match binding 'c'")
	}
	if k.Contains("s") || k.Contains("ы") {
		t.Errorf("unrelated keys should not match")
	}
}

func TestContainsCyrillicLayout(t *testing.T) {
	k := NewKey("a") // physical A -> Cyrillic Ф
	if !k.Contains("ф") {
		t.Errorf("Cyrillic 'ф' (physical A) should match binding 'a'")
	}
	if k.Contains("с") { // physical C -> Latin c
		t.Errorf("Cyrillic 'с' should not match binding 'a'")
	}
}
