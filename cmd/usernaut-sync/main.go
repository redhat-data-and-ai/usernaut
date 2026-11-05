/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// usernaut-sync renders Group CRs from a dataproduct-config checkout and
// applies them with server-side apply.
//
// It exists because the CI job used to spawn ~24k processes (2x yq, ls, sort,
// uniq per group) and one `oc apply` per group: 2,169 applies for 6 real
// changes, 29 minutes of API round trips.
//
// Usage, post-merge (only what the merge touched):
//
//	usernaut-sync -root . -namespace usernaut -changed-from "$CI_COMMIT_BEFORE_SHA"
//
// Scheduled drift control (whole tree, deleting Groups whose file is gone):
//
//	usernaut-sync -root . -namespace usernaut -prune
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	usernautdevv1alpha1 "github.com/redhat-data-and-ai/usernaut/api/v1alpha1"
)

const (
	fieldOwner   = "usernaut-sync"
	createdByKey = "app.kubernetes.io/created-by"
	createdByVal = "usernaut"
)

var (
	root        = flag.String("root", ".", "dataproduct-config checkout")
	namespace   = flag.String("namespace", "", "namespace to apply into (required unless -print)")
	changedFrom = flag.String("changed-from", "", "only reconcile groups touched since this git ref (empty = full sync)")
	workers     = flag.Int("workers", 16, "concurrent apply workers")
	kubeconfig  = flag.String("kubeconfig", "", "kubeconfig path (default: $KUBECONFIG, then ~/.kube/config, then in-cluster)")
	serverDry   = flag.Bool("server-dry-run", false, "validate against the API server without persisting")
	strict      = flag.Bool("strict", false, "fail on unknown group-file fields (enable in MR validation)")
	prune       = flag.Bool("prune", false, "full sync only: delete Groups whose file no longer exists")
	printOnly   = flag.Bool("print", false, "render manifests to stdout and exit")
	verbose     = flag.Bool("v", false, "log every applied group")
)

// tree indexes the checkout.
type tree struct {
	groupFiles []string          // every dataproducts/*/*/groups/*.yaml
	byBase     map[string]string // "<kind>/<product>/<file basename>" -> path
	byName     map[string]string // spec.group_name -> path (duplicates reported)
	duplicates map[string][]string
	accessReqs map[string][]string // "<kind>/<product>/<group dir>" -> usernames
}

func main() {
	flag.Parse()
	if *namespace == "" && !*printOnly {
		fmt.Fprintln(os.Stderr, "-namespace is required")
		flag.Usage()
		os.Exit(2)
	}
	if *workers < 1 {
		fmt.Fprintln(os.Stderr, "-workers must be >= 1")
		os.Exit(2)
	}
	if *prune && *changedFrom != "" {
		fmt.Fprintln(os.Stderr, "-prune only applies to a full sync (drop -changed-from)")
		os.Exit(2)
	}
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	t, err := loadTree()
	if err != nil {
		return err
	}
	files, err := selectGroupFiles(t)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "no group or access-request changes; nothing to do")
		return nil
	}

	groups := make([]*usernautdevv1alpha1.Group, 0, len(files))
	rendered := map[string]string{}
	for _, f := range files {
		g, err := renderGroup(f, t)
		if err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
		// a second file claiming the same group_name would fight over the CR
		if other, ok := t.duplicates[g.Name]; ok {
			return fmt.Errorf("group_name %q is rendered by %s", g.Name, strings.Join(other, ", "))
		}
		rendered[g.Name] = f
		groups = append(groups, g)
	}
	fmt.Fprintf(os.Stderr, "rendered %d Group CRs from %d files\n", len(groups), len(files))

	if *printOnly {
		for i, g := range groups {
			if i > 0 {
				fmt.Println("---")
			}
			b, err := yaml.Marshal(g)
			if err != nil {
				return err
			}
			os.Stdout.Write(b)
		}
		return nil
	}

	c, err := newClient()
	if err != nil {
		return err
	}
	if err := applyAll(c, groups); err != nil {
		return err
	}
	if *prune {
		return pruneGroups(c, rendered)
	}
	return nil
}

// ---------------------------------------------------------------- discovery

func loadTree() (*tree, error) {
	t := &tree{
		byBase:     map[string]string{},
		byName:     map[string]string{},
		duplicates: map[string][]string{},
		accessReqs: map[string][]string{},
	}
	files, err := filepath.Glob(filepath.Join(*root, "dataproducts", "*", "*", "groups", "*.yaml"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	t.groupFiles = files
	for _, f := range files {
		scope, err := scopeOf(f)
		if err != nil {
			return nil, err
		}
		t.byBase[scope] = f

		name, err := groupNameOf(f)
		if err != nil {
			return nil, err
		}
		if name == "" {
			continue
		}
		if prev, ok := t.byName[name]; ok && prev != f {
			t.duplicates[name] = append(t.duplicates[name], prev, f)
			continue
		}
		t.byName[name] = f
	}

	dirs, err := filepath.Glob(filepath.Join(*root, "dataproducts", "*", "*", "access-requests", "groups", "*"))
	if err != nil {
		return nil, err
	}
	for _, dir := range dirs {
		scope, err := scopeOf(dir)
		if err != nil {
			return nil, err
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() || filepath.Ext(e.Name()) != ".yaml" {
				continue
			}
			t.accessReqs[scope] = append(t.accessReqs[scope], strings.TrimSuffix(e.Name(), ".yaml"))
		}
	}
	return t, nil
}

// scopeOf maps ".../<kind>/<product>/groups/<name>.yaml" and
// ".../<kind>/<product>/access-requests/groups/<name>" to "<kind>/<product>/<name>",
// so an access request can only ever join a group in the same dataproduct.
func scopeOf(path string) (string, error) {
	parts := strings.Split(filepath.ToSlash(filepath.Clean(path)), "/")
	for i := 2; i+1 < len(parts); i++ {
		if parts[i] != "groups" {
			continue
		}
		name := strings.TrimSuffix(parts[i+1], ".yaml")
		return parts[i-2] + "/" + parts[i-1] + "/" + name, nil
	}
	return "", fmt.Errorf("%s: not under a groups/ directory", path)
}

func groupNameOf(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var doc struct {
		GroupName string `json:"group_name"`
	}
	if err := yaml.Unmarshal(b, &doc); err != nil {
		return "", err
	}
	return doc.GroupName, nil
}

// selectGroupFiles returns the group YAMLs to reconcile. Empty -changed-from
// means the whole tree (scheduled sync).
func selectGroupFiles(t *tree) ([]string, error) {
	if *changedFrom == "" {
		return t.groupFiles, nil
	}
	if isZeroSHA(*changedFrom) {
		fmt.Fprintln(os.Stderr, "before-sha is empty/zero; falling back to a full sync")
		return t.groupFiles, nil
	}
	out, err := exec.Command("git", "-C", *root, "-c", "core.quotePath=false",
		"diff", "-z", "--name-only", "--relative", *changedFrom+"...HEAD").Output()
	if err != nil {
		// shallow clone, force push, or unknown ref: a partial sync is not
		// safe to guess, so re-assert everything.
		fmt.Fprintf(os.Stderr, "warning: git diff %s...HEAD failed (%v); falling back to a full sync\n", *changedFrom, err)
		return t.groupFiles, nil
	}

	seen := map[string]bool{}
	var files []string
	add := func(f string) {
		if f != "" && !seen[f] {
			seen[f] = true
			files = append(files, f)
		}
	}
	for _, c := range strings.Split(string(out), "\x00") {
		if c == "" {
			continue
		}
		if strings.HasPrefix(c, "scripts/") || c == ".gitlab-ci.yml" {
			return t.groupFiles, nil
		}
		scope, err := scopeOf(c)
		if err != nil {
			continue
		}
		if f, ok := t.byBase[scope]; ok {
			add(f)
			continue
		}
		// access request: dir may be named after group_name, not the file
		if f, ok := t.byName[strings.SplitN(scope, "/", 3)[2]]; ok {
			if sameScope(f, scope) {
				add(f)
			}
		}
	}
	sort.Strings(files)
	return files, nil
}

func sameScope(groupFile, scope string) bool {
	s, err := scopeOf(groupFile)
	if err != nil {
		return false
	}
	// compare "<kind>/<product>" only
	return strings.Join(strings.SplitN(s, "/", 3)[:2], "/") == strings.Join(strings.SplitN(scope, "/", 3)[:2], "/")
}

func isZeroSHA(s string) bool {
	if s == "" {
		return true
	}
	return strings.Trim(s, "0") == ""
}

// ------------------------------------------------------------------ render

func renderGroup(path string, t *tree) (*usernautdevv1alpha1.Group, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var spec usernautdevv1alpha1.GroupSpec
	if *strict {
		// approvers is review metadata, not part of the CRD
		var doc struct {
			usernautdevv1alpha1.GroupSpec `json:",inline"`
			Approvers                     []string `json:"approvers,omitempty"`
		}
		if err := yaml.UnmarshalStrict(b, &doc); err != nil {
			return nil, err
		}
		spec = doc.GroupSpec
	} else {
		if err := yaml.Unmarshal(b, &spec); err != nil {
			return nil, err
		}
		var raw map[string]any
		if err := yaml.Unmarshal(b, &raw); err == nil {
			for k := range raw {
				if !knownKeys[k] {
					fmt.Fprintf(os.Stderr, "warning: %s: unknown field %q will be pruned\n", path, k)
				}
			}
		}
	}
	if spec.GroupName == "" {
		return nil, fmt.Errorf("missing group_name")
	}

	// access-request directories sit next to the groups/ dir of the same
	// dataproduct, and may be named after the file or after group_name.
	kindProduct := strings.Join(strings.SplitN(mustScope(path), "/", 3)[:2], "/")
	arUsers := t.accessReqs[kindProduct+"/"+spec.GroupName]
	if base := strings.TrimSuffix(filepath.Base(path), ".yaml"); base != spec.GroupName {
		arUsers = append(arUsers, t.accessReqs[kindProduct+"/"+base]...)
	}
	spec.Members.Users = mergeUsers(spec.Members.Users, arUsers)

	g := &usernautdevv1alpha1.Group{
		TypeMeta: metav1.TypeMeta{
			APIVersion: usernautdevv1alpha1.GroupVersion.String(),
			Kind:       "Group",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      spec.GroupName,
			Namespace: *namespace,
			Labels:    map[string]string{createdByKey: createdByVal},
		},
		Spec: spec,
	}
	if strings.Contains(filepath.ToSlash(path), "/unstructured/") {
		g.Labels["app.kubernetes.io/part-of"] = "unstructured"
	}
	return g, nil
}

// knownKeys are the group-file fields we understand, used for the lenient
// path only (-strict decodes against the typed spec instead).
var knownKeys = map[string]bool{
	"group_name":   true,
	"approvers":    true,
	"members":      true,
	"backends":     true,
	"group_params": true,
}

func mustScope(path string) string {
	s, err := scopeOf(path)
	if err != nil {
		return ""
	}
	return s
}

func mergeUsers(a, b []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(a)+len(b))
	for _, s := range append(append([]string{}, a...), b...) {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, dup := seen[s]; dup {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	// the previous shell implementation emitted a sorted list; keeping the
	// order stable avoids a generation bump on every Group
	sort.Strings(out)
	return out
}

// ------------------------------------------------------------------- apply

func applyAll(c client.Client, groups []*usernautdevv1alpha1.Group) error {
	ctx := context.Background()
	opts := []client.PatchOption{client.FieldOwner(fieldOwner), client.ForceOwnership}
	if *serverDry {
		opts = append(opts, client.DryRunAll)
	}

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		ok       int
		failures []string
		ch       = make(chan *usernautdevv1alpha1.Group)
	)
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for g := range ch {
				u, convErr := toApplyObject(g)
				err := convErr
				if err == nil {
					err = c.Patch(ctx, u, client.Apply, opts...)
				}
				mu.Lock()
				if err != nil {
					failures = append(failures, fmt.Sprintf("%s: %v", g.Name, err))
				} else {
					ok++
					if *verbose {
						fmt.Fprintln(os.Stderr, "applied", g.Name)
					}
				}
				mu.Unlock()
			}
		}()
	}
	for _, g := range groups {
		ch <- g
	}
	close(ch)
	wg.Wait()

	for _, f := range failures {
		fmt.Fprintln(os.Stderr, "FAILED", f)
	}
	if len(failures) > 0 {
		return fmt.Errorf("%d/%d groups failed", len(failures), len(groups))
	}
	fmt.Fprintf(os.Stderr, "applied %d groups to %s (server-dry-run=%t)\n", ok, *namespace, *serverDry)
	return nil
}

// pruneGroups deletes Groups we own that no longer exist in the checkout.
func pruneGroups(c client.Client, rendered map[string]string) error {
	ctx := context.Background()
	var list usernautdevv1alpha1.GroupList
	if err := c.List(ctx, &list,
		client.InNamespace(*namespace),
		client.MatchingLabels{createdByKey: createdByVal}); err != nil {
		return fmt.Errorf("list groups: %w", err)
	}
	deleted := 0
	for i := range list.Items {
		g := &list.Items[i]
		if _, keep := rendered[g.Name]; keep {
			continue
		}
		if err := c.Delete(ctx, g); err != nil {
			return fmt.Errorf("delete %s: %w", g.Name, err)
		}
		deleted++
		fmt.Fprintln(os.Stderr, "pruned", g.Name)
	}
	fmt.Fprintf(os.Stderr, "pruned %d groups\n", deleted)
	return nil
}

// toApplyObject renders the typed Group as unstructured with the empty status
// removed: sending status:{} from a second field manager would fight the
// controller's status writes.
func toApplyObject(g *usernautdevv1alpha1.Group) (*unstructured.Unstructured, error) {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(g)
	if err != nil {
		return nil, err
	}
	delete(obj, "status")
	return &unstructured.Unstructured{Object: obj}, nil
}

func newClient() (client.Client, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	if *kubeconfig != "" {
		rules.ExplicitPath = *kubeconfig
	}
	// ClientConfig() uses the kubeconfig when present (including $KUBECONFIG,
	// where `oc login` writes), and only falls back to in-cluster config when
	// there is no kubeconfig at all.
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		rules, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, err
	}
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := usernautdevv1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return client.New(cfg, client.Options{Scheme: scheme})
}
