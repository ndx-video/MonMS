package documents

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/monms/monms/internal/schema"
	"github.com/pocketbase/pocketbase/core"
)

// CollectionTree is one markdown-bound collection and its folder hierarchy.
type CollectionTree struct {
	Name    string
	Root    string
	Folders []FolderNode
	Orphans []DocumentLeaf
}

// FolderNode is a nested folder in a document tree.
type FolderNode struct {
	Name     string
	Path     string
	Children []FolderNode
	Docs     []DocumentLeaf
}

// DocumentLeaf is one markdown document in the tree.
type DocumentLeaf struct {
	Title      string
	Slug       string
	Path       string
	ID         string
	Collection string
}

// Forest is the full documents browser payload.
type Forest struct {
	Collections []CollectionTree
	OrphanCount int
}

// BuildForest groups synced markdown records into navigable folder trees.
func BuildForest(app core.App, siteAbs string) (Forest, error) {
	bindings, err := schema.LoadMarkdownBindings(filepath.Join(siteAbs, "schema"))
	if err != nil {
		return Forest{}, err
	}

	diff, err := DiffOrphans(app, siteAbs)
	if err != nil {
		return Forest{}, err
	}

	var collections []CollectionTree
	for _, binding := range bindings {
		tree, err := buildCollectionTree(app, binding)
		if err != nil {
			return Forest{}, err
		}
		collections = append(collections, tree)
	}

	sort.Slice(collections, func(i, j int) bool {
		return collections[i].Name < collections[j].Name
	})

	return Forest{
		Collections: collections,
		OrphanCount: len(diff.Orphans),
	}, nil
}

func buildCollectionTree(app core.App, binding schema.CollectionMeta) (CollectionTree, error) {
	records, err := app.FindAllRecords(binding.Name)
	if err != nil {
		return CollectionTree{}, err
	}

	root := &folderBuilder{}
	for _, rec := range records {
		leaf := DocumentLeaf{
			Title:      rec.GetString("title"),
			Slug:       rec.GetString("slug"),
			Path:       rec.GetString("path"),
			ID:         rec.Id,
			Collection: binding.Name,
		}
		if leaf.Title == "" {
			leaf.Title = leaf.Slug
		}
		if leaf.Slug == "" && leaf.Path != "" {
			leaf.Slug = leaf.Path
		}

		path := strings.Trim(rec.GetString("path"), "/")
		if path == "" {
			root.docs = append(root.docs, leaf)
			continue
		}
		root.insert(path, leaf)
	}

	return CollectionTree{
		Name:    binding.Name,
		Root:    binding.Monms.Root,
		Folders: root.children(""),
		Orphans: sortLeaves(root.docs),
	}, nil
}

type folderBuilder struct {
	subfolders map[string]*folderBuilder
	docs       []DocumentLeaf
}

func (f *folderBuilder) insert(path string, leaf DocumentLeaf) {
	path = strings.Trim(path, "/")
	if path == "" {
		f.docs = append(f.docs, leaf)
		return
	}

	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		f.docs = append(f.docs, leaf)
		return
	}

	if f.subfolders == nil {
		f.subfolders = make(map[string]*folderBuilder)
	}

	head := parts[0]
	tail := strings.Join(parts[1:], "/")
	child, ok := f.subfolders[head]
	if !ok {
		child = &folderBuilder{}
		f.subfolders[head] = child
	}
	child.insert(tail, leaf)
}

func (f *folderBuilder) children(prefix string) []FolderNode {
	if f == nil || len(f.subfolders) == 0 {
		return nil
	}

	names := make([]string, 0, len(f.subfolders))
	for name := range f.subfolders {
		names = append(names, name)
	}
	sort.Strings(names)

	out := make([]FolderNode, 0, len(names))
	for _, name := range names {
		child := f.subfolders[name]
		nodePath := name
		if prefix != "" {
			nodePath = prefix + "/" + name
		}
		out = append(out, FolderNode{
			Name:     name,
			Path:     nodePath,
			Children: child.children(nodePath),
			Docs:     sortLeaves(child.docs),
		})
	}
	return out
}

func sortLeaves(leaves []DocumentLeaf) []DocumentLeaf {
	if len(leaves) == 0 {
		return nil
	}
	sort.Slice(leaves, func(i, j int) bool {
		return leaves[i].Slug < leaves[j].Slug
	})
	return leaves
}
