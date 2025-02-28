package client

// This module provides functionality for constructing a Merkle tree of uploadable inputs.
import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	cpb "github.com/bazelbuild/remote-apis-sdks/go/api/command"
	"github.com/bazelbuild/remote-apis-sdks/go/pkg/command"
	"github.com/bazelbuild/remote-apis-sdks/go/pkg/digest"
	"github.com/bazelbuild/remote-apis-sdks/go/pkg/filemetadata"
	"github.com/bazelbuild/remote-apis-sdks/go/pkg/uploadinfo"
	"github.com/pkg/errors"

	repb "github.com/bazelbuild/remote-apis/build/bazel/remote/execution/v2"
	log "github.com/golang/glog"
)

// treeNode represents a file tree, which is an intermediate representation used to encode a Merkle
// tree later. It corresponds roughly to a *repb.Directory, but with pointers, not digests, used to
// refer to other nodes.
type treeNode struct {
	leaves   map[string]*fileSysNode
	children map[string]*treeNode
}

type fileNode struct {
	ue           *uploadinfo.Entry
	isExecutable bool
}

type symlinkNode struct {
	target string
}

type fileSysNode struct {
	file                 *fileNode
	emptyDirectoryMarker bool
	symlink              *symlinkNode
	nodeProperties       *cpb.NodeProperties
}

// TreeStats contains various stats/metadata of the constructed Merkle tree.
// Note that these stats count the overall input tree, even if some parts of it are not unique.
// For example, if a file "foo" of 10 bytes occurs 5 times in the tree, it will be counted as 5
// InputFiles and 50 TotalInputBytes.
type TreeStats struct {
	// The total number of input files.
	InputFiles int
	// The total number of input directories.
	InputDirectories int
	// The total number of input symlinks
	InputSymlinks int
	// The overall number of bytes from all the inputs.
	TotalInputBytes int64
	// TODO(olaola): number of FileMetadata cache hits/misses go here.
}

// TreeSymlinkOpts controls how symlinks are handled when constructing a tree.
type TreeSymlinkOpts struct {
	// By default, a symlink is converted into its targeted file.
	// If true, preserve the symlink.
	Preserved bool
	// If true, the symlink target (if not dangling) is followed.
	FollowsTarget bool
	// If true, overrides Preserved=true for symlinks that point outside the
	// exec root, converting them into their targeted files while preserving
	// symlinks that point to files within the exec root.  Has no effect if
	// Preserved=false, as all symlinks are materialized.
	MaterializeOutsideExecRoot bool
}

// DefaultTreeSymlinkOpts returns a default DefaultTreeSymlinkOpts object.
func DefaultTreeSymlinkOpts() *TreeSymlinkOpts {
	return &TreeSymlinkOpts{
		FollowsTarget: true,
	}
}

// treeSymlinkOpts returns a TreeSymlinkOpts object based on the given SymlinkBehaviorType.
func treeSymlinkOpts(opts *TreeSymlinkOpts, sb command.SymlinkBehaviorType) *TreeSymlinkOpts {
	if opts == nil {
		opts = DefaultTreeSymlinkOpts()
	}
	switch sb {
	case command.ResolveSymlink:
		opts.Preserved = false
	case command.PreserveSymlink:
		opts.Preserved = true
	}
	return opts
}

// shouldIgnore returns whether a given input should be excluded based on the given InputExclusions,
func shouldIgnore(inp string, t command.InputType, excl []*command.InputExclusion) bool {
	for _, r := range excl {
		if r.Type != command.UnspecifiedInputType && r.Type != t {
			continue
		}
		if m, _ := regexp.MatchString(r.Regex, inp); m {
			return true
		}
	}
	return false
}

func getRelPath(base, path string) (string, error) {
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path %v is not under %v", path, base)
	}
	return rel, nil
}

// getTargetRelPath returns two versions of targetPath, the first is relative to execRoot
// and the second is relative to the directory of symlinkRelPath.
// symlinkRelPath must be relative to execRoot.
// targetPath must either be absolute or relative to the directory of symlinkRelPath.
// If targetPath is not a descendant of execRoot, an error is returned.
func getTargetRelPath(execRoot, symlinkRelPath string, targetPath string) (relExecRoot string, relSymlinkDir string, err error) {
	symlinkAbsDir := filepath.Join(execRoot, filepath.Dir(symlinkRelPath))
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(symlinkAbsDir, targetPath)
	}

	relExecRoot, err = getRelPath(execRoot, targetPath)
	if err != nil {
		return "", "", err
	}

	relSymlinkDir, err = filepath.Rel(symlinkAbsDir, targetPath)
	return relExecRoot, relSymlinkDir, err
}

// getRemotePath generates a remote path for a given local path
// by replacing workingDir component with remoteWorkingDir
func getRemotePath(path, workingDir, remoteWorkingDir string) (string, error) {
	workingDirRelPath, err := filepath.Rel(workingDir, path)
	if err != nil {
		return "", fmt.Errorf("getRemotePath failed while trying to get working dir relative path of %q, err: %v", path, err)
	}
	remotePath := filepath.Join(remoteWorkingDir, workingDirRelPath)
	return remotePath, nil
}

// getExecRootRelPaths returns local and remote exec-root-relative paths for a given local absolute path
func getExecRootRelPaths(absPath, execRoot, workingDir, remoteWorkingDir string) (relPath string, remoteRelPath string, err error) {
	if relPath, err = getRelPath(execRoot, absPath); err != nil {
		return "", "", err
	}
	if remoteWorkingDir == "" || remoteWorkingDir == workingDir {
		return relPath, relPath, nil
	}
	if remoteRelPath, err = getRemotePath(relPath, workingDir, remoteWorkingDir); err != nil {
		return relPath, "", err
	}
	log.V(3).Infof("getExecRootRelPaths(%q, %q, %q, %q)=(%q, %q)", absPath, execRoot, workingDir, remoteWorkingDir, relPath, remoteRelPath)
	return relPath, remoteRelPath, nil
}

// evalParentSymlinks replaces each parent element in relPath with its target if it's a symlink.
//
// Returns the evaluated path with a list of parent symlinks if any. All are relative to execRoot, but not necessarily descendents of it.
// Returned paths may not be filepath.Clean.
// The basename of relPath is not resolved. It remains a symlink if it is one.
// Any errors would be from accessing files.
// Example: execRoot=/a relPath=b/c/d/e.go, b->bb, evaledPath=/a/bb/c/d/e.go, symlinks=[a/b]
func evalParentSymlinks(execRoot, relPath string, materializeOutsideExecRoot bool, fmdCache filemetadata.Cache) (string, []string, error) {
	var symlinks []string
	evaledPathBuilder := strings.Builder{}
	// targetPathBuilder captures the absolute path to the evaluated target so far.
	// It is effectively what relative symlinks are relative to. If materialization
	// is enabled, the materialized path may represent a different tree which makes it
	// unusable with relative symlinks.
	targetPathBuilder := strings.Builder{}
	targetPathBuilder.WriteString(execRoot)
	targetPathBuilder.WriteRune(filepath.Separator)

	ps := strings.Split(relPath, string(filepath.Separator))
	lastIndex := len(ps) - 1
	for i, p := range ps {
		if i != 0 {
			evaledPathBuilder.WriteRune(filepath.Separator)
			targetPathBuilder.WriteRune(filepath.Separator)
		}
		if i == lastIndex {
			// Do not resolve basename.
			evaledPathBuilder.WriteString(p)
			break
		}

		relP := evaledPathBuilder.String() + p
		absP := filepath.Join(execRoot, relP)
		fmd := fmdCache.Get(absP)
		if fmd.Symlink == nil {
			// Not a symlink.
			evaledPathBuilder.WriteString(p)
			targetPathBuilder.WriteString(p)
			continue
		}

		if filepath.IsAbs(fmd.Symlink.Target) {
			targetPathBuilder.Reset()
		}
		targetPathBuilder.WriteString(fmd.Symlink.Target)
		// log.V(5).Infof("eval: relPath=%s, relP=%s, absP=%s, targetPath=%s", relPath, relP, absP, targetPathBuilder.String())

		_, targetRelSymlinkDir, err := getTargetRelPath(execRoot, relP, targetPathBuilder.String())
		if err != nil {
			if materializeOutsideExecRoot {
				evaledPathBuilder.WriteString(p)
				continue
			}
			return "", nil, err
		}
		evaledPathBuilder.WriteString(targetRelSymlinkDir)
		symlinks = append(symlinks, relP)
	}
	return evaledPathBuilder.String(), symlinks, nil
}

// loadFiles reads all files specified by the given InputSpec (descending into subdirectories
// recursively), and loads their contents into the provided map.
func loadFiles(execRoot, localWorkingDir, remoteWorkingDir string, excl []*command.InputExclusion, filesToProcess []string, fs map[string]*fileSysNode, cache filemetadata.Cache, opts *TreeSymlinkOpts, nodeProperties map[string]*cpb.NodeProperties) error {
	if opts == nil {
		opts = DefaultTreeSymlinkOpts()
	}

	var seenAncestorSymlinkRel map[string]bool
	if opts.Preserved {
		// Remember ancestor symlinks to avoid redundant processing for shared ancestors.
		// This ensures that an ancestor symlink is only ever added once to the list.
		seenAncestorSymlinkRel = make(map[string]bool, len(filesToProcess))
	}
	for len(filesToProcess) != 0 {
		path := filesToProcess[0]
		filesToProcess = filesToProcess[1:]

		if path == "" {
			return errors.New("empty Input, use \".\" for entire exec root")
		}
		var absPath string
		if opts.Preserved {
			evaledPath, parentSymlinks, err := evalParentSymlinks(execRoot, path, opts.MaterializeOutsideExecRoot, cache)
			log.V(3).Infof("loadFiles: path=%s, evaled=%s, parentSymlinks=%v, err=%v", path, evaledPath, parentSymlinks, err)
			if err != nil {
				return err
			}
			absPath = filepath.Join(execRoot, evaledPath)
			for _, p := range parentSymlinks {
				if _, ok := seenAncestorSymlinkRel[p]; ok {
					continue
				}
				seenAncestorSymlinkRel[p] = true
				filesToProcess = append(filesToProcess, p)
			}
		} else {
			absPath = filepath.Join(execRoot, path)
		}
		normPath, remoteNormPath, err := getExecRootRelPaths(absPath, execRoot, localWorkingDir, remoteWorkingDir)
		if err != nil {
			return err
		}
		np := nodeProperties[remoteNormPath]
		meta := cache.Get(absPath)

		// An implication of this is that, if a path is a symlink to a
		// directory, then the symlink attribute takes precedence.
		if meta.Symlink != nil && meta.Symlink.IsDangling && !opts.Preserved {
			// For now, we do not treat a dangling symlink as an error. In the case
			// where the symlink is not preserved (i.e. needs to be converted to a
			// file), we simply ignore this path in the finalized tree.
			continue
		} else if meta.Symlink != nil && opts.Preserved {
			if shouldIgnore(absPath, command.SymlinkInputType, excl) {
				continue
			}
			targetExecRoot, targetSymDir, err := getTargetRelPath(execRoot, normPath, meta.Symlink.Target)
			if err != nil {
				// The symlink points to a file outside the exec root. This is an
				// error unless materialization of symlinks pointing outside the
				// exec root is enabled.
				if !opts.MaterializeOutsideExecRoot {
					return errors.Wrapf(err, "failed to determine the target of symlink %q as a child of %q", normPath, execRoot)
				}
				if meta.Symlink.IsDangling {
					return errors.Errorf("failed to materialize dangling symlink %q with target %q", normPath, meta.Symlink.Target)
				}
				goto processNonSymlink
			}

			fs[remoteNormPath] = &fileSysNode{
				// We cannot directly use meta.Symlink.Target, because it could be
				// an absolute path. Since the remote worker will map the exec root
				// to a different directory, we must strip away the local exec root.
				// See https://github.com/bazelbuild/remote-apis-sdks/pull/229#discussion_r524830458
				symlink:        &symlinkNode{target: targetSymDir},
				nodeProperties: np,
			}

			followsTarget := opts.FollowsTarget
			if seenAncestorSymlinkRel[normPath] {
				// Do not follow target if this symlink was only an ancestor and not an explicit input.
				// Otherwise, the entire tree of the target will be included.
				followsTarget = false
				// If the target is also specified as input, it will appear again in the list. Remove the mark here to ensure it gets followed.
				seenAncestorSymlinkRel[normPath] = false
			}

			if !meta.Symlink.IsDangling && followsTarget {
				// getTargetRelPath validates this target is under execRoot,
				// and the iteration loop will get the relative path to execRoot,
				filesToProcess = append(filesToProcess, targetExecRoot)
			}

			// Done processing this symlink, a subsequent iteration will process
			// the targeted file if necessary.
			continue
		}

	processNonSymlink:
		if meta.IsDirectory {
			if shouldIgnore(absPath, command.DirectoryInputType, excl) {
				continue
			} else if meta.Err != nil {
				return meta.Err
			}

			f, err := os.Open(absPath)
			if err != nil {
				return err
			}

			files, err := f.Readdirnames(-1)
			f.Close()
			if err != nil {
				return err
			}

			if len(files) == 0 {
				if normPath != "." {
					fs[remoteNormPath] = &fileSysNode{emptyDirectoryMarker: true, nodeProperties: np}
				}
				continue
			}
			for _, f := range files {
				filesToProcess = append(filesToProcess, filepath.Join(normPath, f))
			}
		} else {
			if shouldIgnore(absPath, command.FileInputType, excl) {
				continue
			} else if meta.Err != nil {
				return meta.Err
			}

			fs[remoteNormPath] = &fileSysNode{
				file: &fileNode{
					ue:           uploadinfo.EntryFromFile(meta.Digest, absPath),
					isExecutable: meta.IsExecutable,
				},
				nodeProperties: np,
			}
		}
	}
	return nil
}

// ComputeMerkleTree packages an InputSpec into uploadable inputs, returned as uploadinfo.Entrys
func (c *Client) ComputeMerkleTree(ctx context.Context, execRoot, workingDir, remoteWorkingDir string, is *command.InputSpec, cache filemetadata.Cache) (root digest.Digest, inputs []*uploadinfo.Entry, stats *TreeStats, err error) {
	stats = &TreeStats{}
	fs := make(map[string]*fileSysNode)
	for _, i := range is.VirtualInputs {
		if i.Path == "" {
			return digest.Empty, nil, nil, errors.New("empty Path in VirtualInputs")
		}
		absPath := filepath.Join(execRoot, i.Path)
		normPath, remoteNormPath, err := getExecRootRelPaths(absPath, execRoot, workingDir, remoteWorkingDir)
		if err != nil {
			return digest.Empty, nil, nil, err
		}
		np := is.InputNodeProperties[remoteNormPath]
		if i.IsEmptyDirectory {
			if normPath != "." {
				fs[remoteNormPath] = &fileSysNode{emptyDirectoryMarker: true, nodeProperties: np}
			}
			continue
		}
		fs[remoteNormPath] = &fileSysNode{
			file: &fileNode{
				ue:           uploadinfo.EntryFromBlob(i.Contents),
				isExecutable: i.IsExecutable,
			},
			nodeProperties: np,
		}
	}
	if err := loadFiles(execRoot, workingDir, remoteWorkingDir, is.InputExclusions, is.Inputs, fs, cache, treeSymlinkOpts(c.TreeSymlinkOpts, is.SymlinkBehavior), is.InputNodeProperties); err != nil {
		return digest.Empty, nil, nil, err
	}
	ft, err := buildTree(fs)
	if err != nil {
		return digest.Empty, nil, nil, err
	}
	var blobs map[digest.Digest]*uploadinfo.Entry
	var tree map[string]digest.Digest
	if log.V(5) {
		tree = make(map[string]digest.Digest)
	}
	root, blobs, err = packageTree(ft, stats, "", tree)
	if log.V(5) {
		if s, ok := ctx.Value("cl_tree").(*string); ok {
			treePaths := make([]string, 0, len(tree))
			for p := range tree {
				treePaths = append(treePaths, p)
			}
			sort.Strings(treePaths)
			sb := strings.Builder{}
			for _, p := range treePaths {
				sb.WriteString(fmt.Sprintf("  %s: %s\n", p, tree[p]))
			}
			*s = sb.String()
		}
	}
	if err != nil {
		return digest.Empty, nil, nil, err
	}
	for _, ue := range blobs {
		inputs = append(inputs, ue)
	}
	return root, inputs, stats, nil
}

func buildTree(files map[string]*fileSysNode) (*treeNode, error) {
	root := &treeNode{}
	for name, fn := range files {
		segs := strings.Split(name, string(filepath.Separator))
		// The last segment is the filename, so split it off.
		segs, base := segs[0:len(segs)-1], segs[len(segs)-1]

		node := root
		for _, s := range segs {
			if node.children == nil {
				node.children = make(map[string]*treeNode)
			}
			child := node.children[s]
			if child == nil {
				child = &treeNode{}
				node.children[s] = child
			}
			node = child
		}

		if fn.emptyDirectoryMarker {
			if node.children == nil {
				node.children = make(map[string]*treeNode)
			}
			if node.children[base] == nil {
				node.children[base] = &treeNode{}
			}
			continue
		}
		if node.leaves == nil {
			node.leaves = make(map[string]*fileSysNode)
		}
		node.leaves[base] = fn
	}
	return root, nil
}

// If tree is not nil, it will be populated with a flattened tree of path->digest.
// prefix should always be provided as an empty string which will be used to accumolate path prefixes during recursion.
func packageTree(t *treeNode, stats *TreeStats, prefix string, tree map[string]digest.Digest) (root digest.Digest, blobs map[digest.Digest]*uploadinfo.Entry, err error) {
	dir := &repb.Directory{}
	blobs = make(map[digest.Digest]*uploadinfo.Entry)

	var path string
	for name, child := range t.children {
		if tree != nil {
			path = prefix + "/" + name
		}

		dg, childBlobs, err := packageTree(child, stats, path, tree)
		if err != nil {
			return digest.Empty, nil, err
		}

		if tree != nil {
			tree[path] = dg
		}

		dir.Directories = append(dir.Directories, &repb.DirectoryNode{Name: name, Digest: dg.ToProto()})
		for d, b := range childBlobs {
			blobs[d] = b
		}
	}
	sort.Slice(dir.Directories, func(i, j int) bool { return dir.Directories[i].Name < dir.Directories[j].Name })

	for name, n := range t.leaves {
		// A node can have exactly one of file/symlink/emptyDirectoryMarker.
		if n.file != nil {
			dg := n.file.ue.Digest
			dir.Files = append(dir.Files, &repb.FileNode{Name: name, Digest: dg.ToProto(), IsExecutable: n.file.isExecutable, NodeProperties: command.NodePropertiesToAPI(n.nodeProperties)})
			blobs[dg] = n.file.ue
			stats.InputFiles++
			stats.TotalInputBytes += dg.Size
			if tree != nil {
				tree[prefix+"/"+name] = dg
			}
			continue
		}
		if n.symlink != nil {
			dir.Symlinks = append(dir.Symlinks, &repb.SymlinkNode{Name: name, Target: n.symlink.target, NodeProperties: command.NodePropertiesToAPI(n.nodeProperties)})
			stats.InputSymlinks++
		}
	}

	sort.Slice(dir.Files, func(i, j int) bool { return dir.Files[i].Name < dir.Files[j].Name })
	sort.Slice(dir.Symlinks, func(i, j int) bool { return dir.Symlinks[i].Name < dir.Symlinks[j].Name })

	ue, err := uploadinfo.EntryFromProto(dir)
	if err != nil {
		return digest.Empty, nil, err
	}
	dg := ue.Digest
	if tree != nil {
		tree[prefix] = dg
	}
	blobs[dg] = ue
	stats.TotalInputBytes += dg.Size
	stats.InputDirectories++
	return dg, blobs, nil
}

// TreeOutput represents a leaf output node in a nested directory structure (a file, a symlink, or an empty directory).
type TreeOutput struct {
	Digest           digest.Digest
	Path             string
	IsExecutable     bool
	IsEmptyDirectory bool
	SymlinkTarget    string
	NodeProperties   *repb.NodeProperties
}

// FlattenTree takes a Tree message and calculates the relative paths of all the files to
// the tree root. Note that only files/symlinks/empty directories are included in the returned slice,
// not the intermediate directories. Directories containing only other directories will be omitted.
func (c *Client) FlattenTree(tree *repb.Tree, rootPath string) (map[string]*TreeOutput, error) {
	root, err := digest.NewFromMessage(tree.Root)
	if err != nil {
		return nil, err
	}
	dirs := make(map[digest.Digest]*repb.Directory)
	dirs[root] = tree.Root
	for _, ue := range tree.Children {
		dg, e := digest.NewFromMessage(ue)
		if e != nil {
			return nil, e
		}
		dirs[dg] = ue
	}
	return flattenTree(root, rootPath, dirs)
}

func flattenTree(root digest.Digest, rootPath string, dirs map[digest.Digest]*repb.Directory) (map[string]*TreeOutput, error) {
	// Create a queue of unprocessed directories, along with their flattened
	// path names.
	type queueElem struct {
		d digest.Digest
		p string
	}
	queue := []*queueElem{}
	queue = append(queue, &queueElem{d: root, p: rootPath})

	// Process the queue, recording all flattened TreeOutputs as we go.
	flatFiles := make(map[string]*TreeOutput)
	for len(queue) > 0 {
		flatDir := queue[0]
		queue = queue[1:]

		dir, ok := dirs[flatDir.d]
		if !ok {
			return nil, fmt.Errorf("couldn't find directory %s with digest %s", flatDir.p, flatDir.d)
		}

		// Check whether this is an empty directory.
		if len(dir.Files)+len(dir.Directories)+len(dir.Symlinks) == 0 {
			flatFiles[flatDir.p] = &TreeOutput{
				Path:             flatDir.p,
				Digest:           digest.Empty,
				IsEmptyDirectory: true,
				NodeProperties:   dir.NodeProperties,
			}
			continue
		}
		// Add files to the set to return
		for _, file := range dir.Files {
			out := &TreeOutput{
				Path:           filepath.Join(flatDir.p, file.Name),
				Digest:         digest.NewFromProtoUnvalidated(file.Digest),
				IsExecutable:   file.IsExecutable,
				NodeProperties: file.NodeProperties,
			}
			flatFiles[out.Path] = out
		}

		// Add symlinks to the set to return
		for _, sm := range dir.Symlinks {
			out := &TreeOutput{
				Path:           filepath.Join(flatDir.p, sm.Name),
				SymlinkTarget:  sm.Target,
				NodeProperties: sm.NodeProperties,
			}
			flatFiles[out.Path] = out
		}

		// Add subdirectories to the queue
		for _, subdir := range dir.Directories {
			digest := digest.NewFromProtoUnvalidated(subdir.Digest)
			name := filepath.Join(flatDir.p, subdir.Name)
			queue = append(queue, &queueElem{d: digest, p: name})
		}
	}
	return flatFiles, nil
}

func packageDirectories(t *treeNode) (root *repb.Directory, files map[digest.Digest]*uploadinfo.Entry, treePb *repb.Tree, err error) {
	root = &repb.Directory{}
	files = make(map[digest.Digest]*uploadinfo.Entry)
	childDirs := make([]string, 0, len(t.children))
	treePb = &repb.Tree{}

	for name := range t.children {
		childDirs = append(childDirs, name)
	}
	sort.Strings(childDirs)

	for _, name := range childDirs {
		child := t.children[name]
		chRoot, childFiles, chTree, err := packageDirectories(child)
		if err != nil {
			return nil, nil, nil, err
		}
		ue, err := uploadinfo.EntryFromProto(chRoot)
		if err != nil {
			return nil, nil, nil, err
		}
		dg := ue.Digest
		root.Directories = append(root.Directories, &repb.DirectoryNode{Name: name, Digest: dg.ToProto()})
		for d, b := range childFiles {
			files[d] = b
		}
		treePb.Children = append(treePb.Children, chRoot)
		treePb.Children = append(treePb.Children, chTree.Children...)
	}
	sort.Slice(root.Directories, func(i, j int) bool { return root.Directories[i].Name < root.Directories[j].Name })

	for name, n := range t.leaves {
		// A node can have exactly one of file/symlink/emptyDirectoryMarker.
		if n.file != nil {
			dg := n.file.ue.Digest
			root.Files = append(root.Files, &repb.FileNode{Name: name, Digest: dg.ToProto(), IsExecutable: n.file.isExecutable, NodeProperties: command.NodePropertiesToAPI(n.nodeProperties)})
			files[dg] = n.file.ue
			continue
		}
		if n.symlink != nil {
			root.Symlinks = append(root.Symlinks, &repb.SymlinkNode{Name: name, Target: n.symlink.target, NodeProperties: command.NodePropertiesToAPI(n.nodeProperties)})
		}
	}
	sort.Slice(root.Files, func(i, j int) bool { return root.Files[i].Name < root.Files[j].Name })
	sort.Slice(root.Symlinks, func(i, j int) bool { return root.Symlinks[i].Name < root.Symlinks[j].Name })

	return root, files, treePb, nil
}

// ComputeOutputsToUpload transforms the provided local output paths into uploadable Chunkers.
// The paths have to be relative to execRoot.
// It also populates the remote ActionResult, packaging output directories as trees where required.
func (c *Client) ComputeOutputsToUpload(execRoot, workingDir string, paths []string, cache filemetadata.Cache, sb command.SymlinkBehaviorType, nodeProperties map[string]*cpb.NodeProperties) (map[digest.Digest]*uploadinfo.Entry, *repb.ActionResult, error) {
	outs := make(map[digest.Digest]*uploadinfo.Entry)
	resPb := &repb.ActionResult{}
	for _, path := range paths {
		absPath := filepath.Join(execRoot, workingDir, path)
		if _, err := getRelPath(execRoot, absPath); err != nil {
			return nil, nil, err
		}
		meta := cache.Get(absPath)
		if meta.Err != nil {
			if e, ok := meta.Err.(*filemetadata.FileError); ok && e.IsNotFound {
				continue // Ignore missing outputs.
			}
			return nil, nil, meta.Err
		}
		normPath, err := filepath.Rel(filepath.Join(execRoot, workingDir), absPath)
		if err != nil {
			return nil, nil, err
		}
		if !meta.IsDirectory {
			// A regular file.
			ue := uploadinfo.EntryFromFile(meta.Digest, absPath)
			outs[meta.Digest] = ue
			resPb.OutputFiles = append(resPb.OutputFiles, &repb.OutputFile{Path: normPath, Digest: meta.Digest.ToProto(), IsExecutable: meta.IsExecutable, NodeProperties: command.NodePropertiesToAPI(nodeProperties[normPath])})
			continue
		}
		// A directory.
		fs := make(map[string]*fileSysNode)
		if e := loadFiles(absPath, "", "", nil, []string{"."}, fs, cache, treeSymlinkOpts(c.TreeSymlinkOpts, sb), nodeProperties); e != nil {
			return nil, nil, e
		}
		ft, err := buildTree(fs)
		if err != nil {
			return nil, nil, err
		}

		rootDir, files, treePb, err := packageDirectories(ft)
		if err != nil {
			return nil, nil, err
		}
		ue, err := uploadinfo.EntryFromProto(rootDir)
		if err != nil {
			return nil, nil, err
		}
		outs[ue.Digest] = ue
		treePb.Root = rootDir
		ue, err = uploadinfo.EntryFromProto(treePb)
		if err != nil {
			return nil, nil, err
		}
		outs[ue.Digest] = ue
		for _, ue := range files {
			outs[ue.Digest] = ue
		}
		resPb.OutputDirectories = append(resPb.OutputDirectories, &repb.OutputDirectory{Path: normPath, TreeDigest: ue.Digest.ToProto()})
		// Upload the child directories individually as well
		ueRoot, _ := uploadinfo.EntryFromProto(treePb.Root)
		outs[ueRoot.Digest] = ueRoot
		for _, child := range treePb.Children {
			ueChild, _ := uploadinfo.EntryFromProto(child)
			outs[ueChild.Digest] = ueChild
		}
	}
	return outs, resPb, nil
}
