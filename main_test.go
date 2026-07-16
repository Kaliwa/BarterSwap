package main

import "testing"

func TestEnv(t *testing.T) {
	t.Setenv("BARTERSWAP_TEST_KEY", "valeur")
	if got := env("BARTERSWAP_TEST_KEY", "défaut"); got != "valeur" {
		t.Errorf("env = %q, attendu la valeur définie", got)
	}
	if got := env("BARTERSWAP_TEST_ABSENT", "défaut"); got != "défaut" {
		t.Errorf("env = %q, attendu le repli", got)
	}
}
