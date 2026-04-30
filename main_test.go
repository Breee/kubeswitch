package main

import (
	"bytes"
	"encoding/json"
	"os"
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

func newTestModel() model {
        return model{
                contexts: []contextNode{
                        {name: "alpha", namespaces: []string{"ns1", "ns2"}, expanded: false},
                        {name: "beta", namespaces: []string{"ns3"}, expanded: true},
                },
                cursor: 0,
                height: 20,
        }
}

func TestHandleNavKey_RightExpandsCluster(t *testing.T) {
        m := newTestModel()
        // cursor on "alpha" (collapsed)
        m.cursor = 0
        result, _ := m.handleNavKey("right")
        rm := result.(model)
        if !rm.contexts[0].expanded {
                t.Error("expected alpha to be expanded after right arrow")
        }
        // beta should be collapsed (only one expanded at a time)
        if rm.contexts[1].expanded {
                t.Error("expected beta to be collapsed after expanding alpha")
        }
}

func TestHandleNavKey_LeftCollapsesCluster(t *testing.T) {
        m := newTestModel()
        // cursor on "beta" which is expanded; beta is at flat index 1 (alpha collapsed)
        m.cursor = 1
        result, _ := m.handleNavKey("left")
        rm := result.(model)
        if rm.contexts[1].expanded {
                t.Error("expected beta to be collapsed after left arrow")
        }
}

func TestHandleNavKey_LeftOnNamespaceJumpsToCluster(t *testing.T) {
        m := newTestModel()
        // beta is expanded with ns3; beta is at flat index 1, ns3 at flat index 2
        m.cursor = 2
        result, _ := m.handleNavKey("left")
        rm := result.(model)
        if rm.cursor != 1 {
                t.Errorf("expected cursor to jump to beta (index 1), got %d", rm.cursor)
        }
}

func TestHandleNavKey_RightOnNamespaceDoesNothing(t *testing.T) {
        m := newTestModel()
        // cursor on ns3 (flat index 2)
        m.cursor = 2
        result, _ := m.handleNavKey("right")
        rm := result.(model)
        if rm.cursor != 2 {
                t.Errorf("expected cursor to stay at 2, got %d", rm.cursor)
        }
        if rm.contexts[1].expanded != true {
                t.Error("expected beta to remain expanded")
        }
}

func TestHandleNavKey_RightOnAlreadyExpandedDoesNothing(t *testing.T) {
        m := newTestModel()
        // cursor on "beta" which is already expanded (flat index 1)
        m.cursor = 1
        result, _ := m.handleNavKey("right")
        rm := result.(model)
        if !rm.contexts[1].expanded {
                t.Error("expected beta to remain expanded")
        }
}

func TestHandleNavKey_HAndLKeysWork(t *testing.T) {
        m := newTestModel()
        m.cursor = 0
        result, _ := m.handleNavKey("l")
        rm := result.(model)
        if !rm.contexts[0].expanded {
                t.Error("expected alpha to be expanded after 'l' key")
        }

        // Now collapse it with 'h'
        rm.cursor = 0
        result2, _ := rm.handleNavKey("h")
        rm2 := result2.(model)
        if rm2.contexts[0].expanded {
                t.Error("expected alpha to be collapsed after 'h' key")
        }
}

func TestFilteredContextIndex(t *testing.T) {
        m := newTestModel()
        // alpha is at index 0, beta is at index 1 (alpha not expanded)
        if got := m.filteredContextIndex("alpha"); got != 0 {
                t.Errorf("expected alpha at index 0, got %d", got)
        }
        if got := m.filteredContextIndex("beta"); got != 1 {
                t.Errorf("expected beta at index 1, got %d", got)
        }

        // Expand alpha: alpha=0, ns1=1, ns2=2, beta=3
        m.contexts[0].expanded = true
        if got := m.filteredContextIndex("beta"); got != 3 {
                t.Errorf("expected beta at index 3 when alpha expanded, got %d", got)
        }
}

func TestDebugLog_Enabled(t *testing.T) {
        // Capture stderr
        oldStderr := os.Stderr
        r, w, _ := os.Pipe()
        os.Stderr = w

        oldDebugMode := debugMode
        debugMode = true
        defer func() {
                debugMode = oldDebugMode
                os.Stderr = oldStderr
        }()

        debugLog("test %s %d", "hello", 42)
        w.Close()

        var buf bytes.Buffer
        buf.ReadFrom(r)
        output := buf.String()

        if output != "[DEBUG] test hello 42\n" {
                t.Errorf("expected debug output %q, got %q", "[DEBUG] test hello 42\n", output)
        }
}

func TestDebugLog_Disabled(t *testing.T) {
        oldStderr := os.Stderr
        r, w, _ := os.Pipe()
        os.Stderr = w

        oldDebugMode := debugMode
        debugMode = false
        defer func() {
                debugMode = oldDebugMode
                os.Stderr = oldStderr
        }()

        debugLog("should not appear")
        w.Close()

        var buf bytes.Buffer
        buf.ReadFrom(r)
        output := buf.String()

        if output != "" {
                t.Errorf("expected no output when debug disabled, got %q", output)
	}
}
