package main

import (
	"encoding/json"
	"runtime"
	"testing"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestVersionDefault(t *testing.T) {
	if version != "dev" {
		t.Errorf("expected default version to be %q, got %q", "dev", version)
	}
}

func TestGitCommitDefault(t *testing.T) {
	if gitCommit != "unknown" {
		t.Errorf("expected default gitCommit to be %q, got %q", "unknown", gitCommit)
	}
}

func TestBuildDateDefault(t *testing.T) {
	if buildDate != "unknown" {
		t.Errorf("expected default buildDate to be %q, got %q", "unknown", buildDate)
	}
}

func TestGetVersionInfo(t *testing.T) {
	info := getVersionInfo()
	if info.Version != version {
		t.Errorf("expected Version %q, got %q", version, info.Version)
	}
	if info.GitCommit != gitCommit {
		t.Errorf("expected GitCommit %q, got %q", gitCommit, info.GitCommit)
	}
	if info.BuildDate != buildDate {
		t.Errorf("expected BuildDate %q, got %q", buildDate, info.BuildDate)
	}
	if info.GoVersion != runtime.Version() {
		t.Errorf("expected GoVersion %q, got %q", runtime.Version(), info.GoVersion)
	}
	if info.Compiler != runtime.Compiler {
		t.Errorf("expected Compiler %q, got %q", runtime.Compiler, info.Compiler)
	}
	expectedPlatform := runtime.GOOS + "/" + runtime.GOARCH
	if info.Platform != expectedPlatform {
		t.Errorf("expected Platform %q, got %q", expectedPlatform, info.Platform)
	}
}

func TestVersionInfoJSON(t *testing.T) {
	info := getVersionInfo()
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("failed to marshal VersionInfo: %v", err)
	}

	var parsed VersionInfo
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal VersionInfo: %v", err)
	}

	if parsed.Version != info.Version {
		t.Errorf("JSON round-trip: expected Version %q, got %q", info.Version, parsed.Version)
	}
	if parsed.GitCommit != info.GitCommit {
		t.Errorf("JSON round-trip: expected GitCommit %q, got %q", info.GitCommit, parsed.GitCommit)
	}
	if parsed.BuildDate != info.BuildDate {
		t.Errorf("JSON round-trip: expected BuildDate %q, got %q", info.BuildDate, parsed.BuildDate)
	}
	if parsed.GoVersion != info.GoVersion {
		t.Errorf("JSON round-trip: expected GoVersion %q, got %q", info.GoVersion, parsed.GoVersion)
	}
	if parsed.Compiler != info.Compiler {
		t.Errorf("JSON round-trip: expected Compiler %q, got %q", info.Compiler, parsed.Compiler)
	}
	if parsed.Platform != info.Platform {
		t.Errorf("JSON round-trip: expected Platform %q, got %q", info.Platform, parsed.Platform)
	}
}

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
