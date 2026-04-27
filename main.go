package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

type contextNamespaceTuple struct {
	k8sContext   string
	k8sNamespace string
}

var (
	kubeconfLocation = os.Getenv("HOME") + "/.kube/config"
	mergedConfig     *clientcmdapi.Config
)

// Tree data structures

type contextNode struct {
	name       string
	namespaces []string
	err        error
	expanded   bool
	isActive   bool
}

type model struct {
	contexts      []contextNode
	cursor        int // flat index into visible items
	selected      *contextNamespaceTuple
	activeContext string
	activeNs      string
	height        int // terminal height
	offset        int // scroll offset
	filter        string
	filtering     bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if m.filtering {
			switch msg.String() {
			case "esc":
				m.filtering = false
				m.filter = ""
				m.cursor = 0
				m.offset = 0
			case "enter", "up", "down":
				m.filtering = false
				// Fall through to normal key handling below
				return m.handleNavKey(msg.String())
			case "backspace":
				if len(m.filter) > 0 {
					m.filter = m.filter[:len(m.filter)-1]
					m.cursor = 0
					m.offset = 0
				}
			case "ctrl+c":
				return m, tea.Quit
			default:
				if len(msg.String()) == 1 {
					m.filter += msg.String()
					m.cursor = 0
					m.offset = 0
				}
			}
		} else {
			switch msg.String() {
			case "q", "ctrl+c", "esc":
				if m.filter != "" {
					m.filter = ""
					m.cursor = 0
					m.offset = 0
				} else {
					return m, tea.Quit
				}
			case "/":
				m.filtering = true
			default:
				return m.handleNavKey(msg.String())
			}
		}
	}

	// Keep cursor in view
	viewHeight := m.viewportHeight()
	if viewHeight > 0 {
		if m.cursor < m.offset {
			m.offset = m.cursor
		}
		if m.cursor >= m.offset+viewHeight {
			m.offset = m.cursor - viewHeight + 1
		}
	}

	return m, nil
}

func (m model) handleNavKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < m.filteredVisibleCount()-1 {
			m.cursor++
		}
	case "enter", " ":
		ctx, ns := m.filteredItemAtCursor()
		if ns == "" {
			for i := range m.contexts {
				if m.contexts[i].name == ctx {
					if m.contexts[i].err != nil {
						break
					}
					wasExpanded := m.contexts[i].expanded
					for j := range m.contexts {
						m.contexts[j].expanded = false
					}
					m.contexts[i].expanded = !wasExpanded
					break
				}
			}
		} else {
			m.selected = &contextNamespaceTuple{ctx, ns}
			return m, tea.Quit
		}
	}

	// Keep cursor in view
	viewHeight := m.viewportHeight()
	if viewHeight > 0 {
		if m.cursor < m.offset {
			m.offset = m.cursor
		}
		if m.cursor >= m.offset+viewHeight {
			m.offset = m.cursor - viewHeight + 1
		}
	}

	return m, nil
}

func (m model) viewportHeight() int {
	if m.height <= 2 {
		return 0
	}
	return m.height - 2 // reserve lines for help text
}

func (m model) visibleCount() int {
	count := 0
	for _, c := range m.contexts {
		count++ // context row
		if c.expanded {
			count += len(c.namespaces)
		}
	}
	return count
}

func (m model) fuzzyMatch(text string) bool {
	if m.filter == "" {
		return true
	}
	lower := strings.ToLower(text)
	pattern := strings.ToLower(m.filter)

	// 1. Exact substring — strongest signal
	if strings.Contains(lower, pattern) {
		return true
	}

	// 2. Subsequence match — handles abbreviations (e.g. "prd" matches "production")
	pi := 0
	for i := 0; i < len(lower) && pi < len(pattern); i++ {
		if lower[i] == pattern[pi] {
			pi++
		}
	}
	if pi == len(pattern) {
		return true
	}

	// 3. Edit distance — handles typos (Elasticsearch AUTO tiers)
	threshold := 0
	switch {
	case len(pattern) <= 2:
		threshold = 0
	case len(pattern) <= 5:
		threshold = 1
	default:
		threshold = 2
	}
	if threshold == 0 {
		return false
	}
	if len(pattern) > len(lower) {
		return levenshtein(lower, pattern) <= threshold
	}
	for i := 0; i <= len(lower)-len(pattern); i++ {
		if levenshtein(lower[i:i+len(pattern)], pattern) <= threshold {
			return true
		}
	}
	return false
}

func levenshtein(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)

	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min(curr[j-1]+1, min(prev[j]+1, prev[j-1]+cost))
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

func (m model) contextMatchesFilter(c contextNode) bool {
	if m.filter == "" {
		return true
	}
	if m.fuzzyMatch(c.name) {
		return true
	}
	for _, ns := range c.namespaces {
		if m.fuzzyMatch(ns) {
			return true
		}
	}
	return false
}

func (m model) filteredVisibleCount() int {
	count := 0
	for _, c := range m.contexts {
		if !m.contextMatchesFilter(c) {
			continue
		}
		count++
		if c.expanded {
			for _, ns := range c.namespaces {
				if m.filter == "" || m.fuzzyMatch(c.name) || m.fuzzyMatch(ns) {
					count++
				}
			}
		}
	}
	return count
}

func (m model) filteredItemAtCursor() (contextName, namespace string) {
	idx := 0
	for _, c := range m.contexts {
		if !m.contextMatchesFilter(c) {
			continue
		}
		if idx == m.cursor {
			return c.name, ""
		}
		idx++
		if c.expanded {
			for _, ns := range c.namespaces {
				if m.filter == "" || m.fuzzyMatch(c.name) || m.fuzzyMatch(ns) {
					if idx == m.cursor {
						return c.name, ns
					}
					idx++
				}
			}
		}
	}
	return "", ""
}

var (
	styleGreen     = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleRed       = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleTurquoise = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	styleCursor    = lipgloss.NewStyle().Bold(true).Reverse(true)
	styleFilter    = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
)

func (m model) View() string {
	var lines []string

	idx := 0
	for _, c := range m.contexts {
		if !m.contextMatchesFilter(c) {
			continue
		}

		prefix := "▸ "
		if c.expanded {
			prefix = "▾ "
		}

		label := c.name
		var style lipgloss.Style
		if c.err != nil {
			label = c.name + " (" + c.err.Error() + ")"
			style = styleRed
		} else if c.isActive {
			label = c.name + " (active)"
			style = styleGreen
		} else {
			style = styleTurquoise
		}

		line := prefix + label
		if idx == m.cursor {
			line = styleCursor.Render(line)
		} else {
			line = style.Render(line)
		}
		lines = append(lines, line)
		idx++

		if c.expanded {
			for _, ns := range c.namespaces {
				if m.filter != "" && !m.fuzzyMatch(c.name) && !m.fuzzyMatch(ns) {
					continue
				}
				nsLine := "    " + ns
				if idx == m.cursor {
					nsLine = styleCursor.Render(nsLine)
				} else if c.isActive && ns == m.activeNs {
					nsLine = styleGreen.Render(nsLine)
				}
				lines = append(lines, nsLine)
				idx++
			}
		}
	}

	// Apply viewport clipping
	viewHeight := m.viewportHeight()
	if viewHeight > 0 && len(lines) > viewHeight {
		end := m.offset + viewHeight
		if end > len(lines) {
			end = len(lines)
		}
		lines = lines[m.offset:end]
	}

	var footer string
	if m.filtering {
		footer = styleFilter.Render("/"+m.filter) + "▎"
	} else if m.filter != "" {
		footer = styleFilter.Render("/"+m.filter) + " (esc to clear)"
	} else {
		footer = "↑/↓ navigate • enter select • / filter • q quit"
	}

	return strings.Join(lines, "\n") + "\n\n" + footer + "\n"
}

func main() {
	var err error

	if len(os.Args) > 1 {
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			printUsage()
		}
	}

	if len(os.Getenv("KUBECONFIG")) > 0 {
		kubeconfLocation = os.Getenv("KUBECONFIG")
	}

	loadingRules := &clientcmd.ClientConfigLoadingRules{Precedence: strings.Split(kubeconfLocation, ":")}

	mergedConfig, err = loadingRules.Load()
	if err != nil {
		log.Fatalln(err)
	}

	if len(os.Args) > 1 {
		quickSwitch()
	}

	// Build tree data
	var contexts []contextNode
	initialCursor := 0
	flatIdx := 0

	for _, thisContextName := range mapKeysToSortedArray(mergedConfig.Contexts) {
		node := contextNode{name: thisContextName}

		namespacesInThisContextsCluster, err := getNamespacesInContextsCluster(thisContextName)
		if err != nil {
			node.err = err
		} else {
			for _, ns := range namespacesInThisContextsCluster {
				node.namespaces = append(node.namespaces, ns.Name)
			}
		}

		if thisContextName == mergedConfig.CurrentContext {
			node.isActive = true
			node.expanded = true
		}

		contexts = append(contexts, node)

		flatIdx++ // context row
		if node.expanded {
			for _, ns := range node.namespaces {
				if node.isActive && ns == mergedConfig.Contexts[thisContextName].Namespace {
					initialCursor = flatIdx
				}
				flatIdx++
			}
		}
	}

	m := model{
		contexts:      contexts,
		cursor:        initialCursor,
		activeContext: mergedConfig.CurrentContext,
		activeNs:      mergedConfig.Contexts[mergedConfig.CurrentContext].Namespace,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		log.Fatalln(err)
	}

	if result := finalModel.(model).selected; result != nil {
		switchContext(*result)
	}
}

func getNamespacesInContextsCluster(k8sContext string) ([]corev1.Namespace, error) {

	config, err := clientcmd.NewDefaultClientConfig(*mergedConfig, &clientcmd.ConfigOverrides{CurrentContext: k8sContext}).ClientConfig()
	if err != nil {
		log.Fatalln(err)
	}

	config.Timeout = time.Second

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln(err)
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		switch err.(type) {
		case *url.Error:
			return []corev1.Namespace{}, fmt.Errorf("unreachable")
		case *apierrors.StatusError:
			return []corev1.Namespace{}, fmt.Errorf("error from api: %w", err)
		default:
			return []corev1.Namespace{}, fmt.Errorf("error")
		}
	}

	return namespaces.Items, nil
}

func switchContext(rh contextNamespaceTuple) {
	mergedConfig.CurrentContext = rh.k8sContext
	mergedConfig.Contexts[rh.k8sContext].Namespace = rh.k8sNamespace

	removeStaleContextConfigs()

	configAccess := clientcmd.NewDefaultClientConfigLoadingRules()

	if err := clientcmd.ModifyConfig(configAccess, *mergedConfig, false); err != nil {
		log.Fatalln(err)
	}

	log.Printf("switched to %s/%s", rh.k8sContext, rh.k8sNamespace)
}

func quickSwitch() {
	if len(os.Args) == 2 {
		if !namespaceExists(mergedConfig.CurrentContext, os.Args[1]) {
			log.Fatalf("namespace %s not found in context %s\n", os.Args[1], mergedConfig.CurrentContext)
		}

		switchContext(contextNamespaceTuple{mergedConfig.CurrentContext, os.Args[1]})
		os.Exit(0)
	}

	if len(os.Args) == 3 && os.Args[2] == "." {
		if !contextExists(os.Args[1]) || !namespaceExists(os.Args[1], "default") {
			log.Fatalf("namespace %s not found in context %s\n", "default", os.Args[1])
		}

		switchContext(contextNamespaceTuple{os.Args[1], "default"})
		os.Exit(0)
	}

	if len(os.Args) == 3 {
		if !contextExists(os.Args[1]) || !namespaceExists(os.Args[1], os.Args[2]) {
			log.Fatalf("namespace %s not found in context %s\n", os.Args[2], os.Args[1])
		}

		switchContext(contextNamespaceTuple{os.Args[1], os.Args[2]})
		os.Exit(0)
	}
}

func removeStaleContextConfigs() {

	for _, configFilename := range strings.Split(kubeconfLocation, ":") {
		var output []string

		cfStat, err := os.Stat(configFilename)
		if err != nil {
			log.Fatalln("could not stat kubeconfig files")
		}

		cfFileMode := cfStat.Mode()

		cfContent, err := ioutil.ReadFile(configFilename)
		if err != nil {
			log.Fatalln("could not read kubeconfig files")
		}
		cfLines := strings.Split(string(cfContent), "\n")

		for _, line := range cfLines {
			if strings.Contains(line, "current-context:") {
				continue
			}

			output = append(output, line)
		}

		if err := ioutil.WriteFile(configFilename, []byte(strings.Join(output, "\n")), cfFileMode); err != nil {
			log.Fatalln("could not update kubeconfig files")
		}
	}
}

func contextExists(k8sContext string) bool {
	_, exists := mergedConfig.Contexts[k8sContext]
	return exists
}

func namespaceExists(k8sContext, k8sNamespace string) bool {
	namespacesInThisContextsCluster, err := getNamespacesInContextsCluster(k8sContext)
	if err != nil {
		log.Fatalln(err)
	}

	for _, ns := range namespacesInThisContextsCluster {
		if ns.Name == k8sNamespace {
			return true
		}
	}

	return false
}

func mapKeysToSortedArray(m map[string]*clientcmdapi.Context) []string {
	var s []string

	for k := range m {
		s = append(s, k)
	}

	sort.Strings(s)
	return s
}

func printUsage() {
	usageText := `usage:

./kubeswitch                          select context/namespace graphically
./kubeswitch <namespace>              switch to namespace in current context quickly
./kubeswitch <context> <namespace>    switch to namespace in context quickly
./kubeswitch <context>/<namespace>    switch to namespace in context quickly`

	fmt.Println(usageText)
	os.Exit(2)
}
