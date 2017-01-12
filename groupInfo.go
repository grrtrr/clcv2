// Recursive, parallel processing of group trees.
// Code closely based on the following examples:
// https://godoc.org/golang.org/x/sync/errgroup (parallel MD5 summation)
// https://golang.org/src/path/filepath/path.go#L393 (recursive filesystem walk)

package clcv2

import (
	"context"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// GroupInfo aggregates information to print the group tree
type GroupInfo struct {
	ID, Name string     // UUID and name of the HW Group
	Parent   string     // UUID of the parent group
	Type     string     // type of the group (e.g. "default")
	Servers  []string   // list of server IDs
	Groups   GroupInfos // list of children
}

// LT defines the sort order for GroupInfo elements
// - special (non-default) groups like 'Archive' or 'Templates' always come first
// - otherwise, groups are sorted in case-insensitive lexicographical order
func (g *GroupInfo) LT(o *GroupInfo) bool {
	if g.Type == o.Type {
		return strings.ToUpper(g.Name) < strings.ToUpper(o.Name)
	}
	return o.Type == "default"
}

// GroupInfos is an array of pointers to GroupInfo
type GroupInfos []*GroupInfo

// GroupInfos implements sort.Interface
func (g GroupInfos) Len() int           { return len(g) }
func (g GroupInfos) Swap(i, j int)      { g[i], g[j] = g[j], g[i] }
func (g GroupInfos) Less(i, j int) bool { return g[i].LT(g[j]) }

// WalkFn processes a single Group entry
type WalkFn func(*Group) error

// WalkGroupTree is the equivalent of filepath.Walk for nested clcv2 Group trees.
func WalkGroupTree(root *Group, w WalkFn) error {
	err := w(root)
	if err != nil {
		return err
	}
	for idx := range root.Groups {
		if err := WalkGroupTree(&root.Groups[idx], w); err != nil {
			return err
		}
	}
	return nil
}

// WalkGroupHierarchy converts the tree at @root into a processed GroupInfo tree
// @ctx:  cancellation context
// @root: root of the tree to start at
// @cb:   callback function to process a single GroupInfo entry
func WalkGroupHierarchy(
	ctx context.Context,
	root *Group,
	cb func(context.Context, *GroupInfo) error) (*GroupInfo, error) {
	var ret *GroupInfo
	var nodes = make(chan GroupInfo)

	g, ctx := errgroup.WithContext(ctx)

	// Do a depth-first tree traversal, serializing extracted @nodes
	g.Go(func() error {
		defer close(nodes)
		return WalkGroupTree(root, func(g *Group) error {
			node := GroupInfo{
				ID:   g.Id,
				Name: g.Name,
				Type: g.Type,
			}

			for _, l := range g.Links {
				if l.Rel == "server" {
					node.Servers = append(node.Servers, l.Id)
				} else if l.Rel == "parentGroup" {
					node.Parent = l.Id
				}
			}

			select {
			case nodes <- node:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	})

	// Process nodes in parallel, using a fixed number of goroutines.
	processedNodes := make(chan GroupInfo)
	const numDigesters = 20
	for i := 0; i < numDigesters; i++ {
		g.Go(func() error {
			for node := range nodes {
				if cb != nil {
					if err := cb(ctx, &node); err != nil {
						return err
					}
				}
				select {
				case processedNodes <- node:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	}
	go func() {
		g.Wait()
		close(processedNodes)
	}()

	// Re-assemble the processed group tree, using a hash map indexed by Group UUID
	// Groups are inserted in order: (a) special groups first, then (b) in lexicographical order
	m := make(map[string]*GroupInfo)
	for node := range processedNodes {
		node := node
		m[node.ID] = &node
	}

	for _, node := range m {
		if node.Parent == "" { // root node
			ret = node
		} else if parent, ok := m[node.Parent]; !ok {
			return nil, errors.Errorf("no parent found for node %s/%s", node.Name, node.ID)
		} else if l := len(parent.Groups); l == 0 {
			parent.Groups = append(parent.Groups, node)
		} else if idx := sort.Search(l, func(i int) bool { return node.LT(parent.Groups[i]) }); idx == l {
			parent.Groups = append(parent.Groups, node) // not found, value is greatest in the order
		} else {
			parent.Groups = append(parent.Groups[:idx], append([]*GroupInfo{node}, parent.Groups[idx:]...)...)
		}
	}

	// If any of the above goroutines returned an error, propagate it to the caller.
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return ret, nil
}
