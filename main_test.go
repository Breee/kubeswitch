package main

import (
	"testing"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestMapKeysToSortedArray_Multiple(t *testing.T) {
	m := map[string]*clientcmdapi.Context{
		"gamma": {},
		"alpha": {},
		"beta":  {},
	}

	got := mapKeysToSortedArray(m)

	if len(got) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(got))
	}

	expected := []string{"alpha", "beta", "gamma"}
	for i, v := range expected {
		if got[i] != v {
			t.Errorf("index %d: expected %q, got %q", i, v, got[i])
		}
	}
}

func TestMapKeysToSortedArray_Empty(t *testing.T) {
	m := map[string]*clientcmdapi.Context{}
	got := mapKeysToSortedArray(m)

	if len(got) != 0 {
		t.Fatalf("expected 0 keys, got %d", len(got))
	}
}

func TestMapKeysToSortedArray_Single(t *testing.T) {
	m := map[string]*clientcmdapi.Context{
		"only": {},
	}

	got := mapKeysToSortedArray(m)

	if len(got) != 1 || got[0] != "only" {
		t.Errorf("expected [\"only\"], got %v", got)
	}
}

func TestContextExists_Found(t *testing.T) {
	mergedConfig = &clientcmdapi.Config{
		Contexts: map[string]*clientcmdapi.Context{
			"prod": {Namespace: "default"},
			"dev":  {Namespace: "default"},
		},
	}

	if !contextExists("prod") {
		t.Error("expected contextExists(\"prod\") to be true")
	}
}

func TestContextExists_NotFound(t *testing.T) {
	mergedConfig = &clientcmdapi.Config{
		Contexts: map[string]*clientcmdapi.Context{
			"prod": {Namespace: "default"},
		},
	}

	if contextExists("staging") {
		t.Error("expected contextExists(\"staging\") to be false")
	}
}

func TestContextExists_EmptyConfig(t *testing.T) {
	mergedConfig = &clientcmdapi.Config{
		Contexts: map[string]*clientcmdapi.Context{},
	}

	if contextExists("anything") {
		t.Error("expected contextExists(\"anything\") to be false on empty config")
	}
}
