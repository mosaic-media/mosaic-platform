// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

// Command licenseheader adds (or verifies) the SPDX license header on Go source
// files. It exists so the header is applied uniformly by a tool rather than by
// each contributor's editor.
//
// Usage:
//
//	go run ./tools/licenseheader                 # add the header to any repo file missing it
//	go run ./tools/licenseheader -check          # list files missing it; exit non-zero if any
//	go run ./tools/licenseheader file.go ...     # operate on only the named files
//
// With no file arguments it walks the tree from -root (default "."). With file
// arguments it processes exactly those (the pre-commit hook passes staged
// files). The -check form is what CI runs to keep the header from drifting. The
// tool is idempotent: a file that already carries the SPDX marker in its leading
// comment block is left untouched.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// header is the SPDX block prepended to each file. AGPL-3.0-only matches the
// repository's stated license (AGPL version 3); change it here — the one place —
// to adjust the notice for this repository. The linking exception has no
// registered SPDX identifier, so it is referenced by file rather than a WITH
// clause.
const header = `// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.
`

const marker = "SPDX-License-Identifier"

func main() {
	check := flag.Bool("check", false, "report files missing the header and exit non-zero; do not modify")
	root := flag.String("root", ".", "directory to walk when no explicit files are given")
	flag.Parse()

	targets, err := collect(*root, flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "licenseheader:", err)
		os.Exit(2)
	}

	var missing, added []string
	for _, path := range targets {
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "licenseheader:", err)
			os.Exit(2)
		}
		if hasHeader(content) {
			continue
		}
		if *check {
			missing = append(missing, path)
			continue
		}
		if err := os.WriteFile(path, prepend(content), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "licenseheader:", err)
			os.Exit(2)
		}
		added = append(added, path)
	}

	if *check {
		if len(missing) > 0 {
			fmt.Fprintf(os.Stderr, "%d file(s) missing the SPDX header:\n", len(missing))
			for _, m := range missing {
				fmt.Fprintln(os.Stderr, "  "+filepath.ToSlash(m))
			}
			fmt.Fprintln(os.Stderr, "\nFix: run 'go run ./tools/licenseheader' (no flags) to add it, then commit.")
			os.Exit(1)
		}
		return
	}

	for _, a := range added {
		fmt.Println("added header:", filepath.ToSlash(a))
	}
	fmt.Printf("licenseheader: %d file(s) updated\n", len(added))
}

// collect returns the .go files to process. When files are named explicitly it
// returns those (filtered to .go, and to ones that exist — a staged rename can
// leave a path that is gone). Otherwise it walks root, skipping .git, vendor and
// testdata.
func collect(root string, files []string) ([]string, error) {
	if len(files) > 0 {
		var targets []string
		for _, f := range files {
			if !strings.HasSuffix(f, ".go") {
				continue
			}
			if info, err := os.Stat(f); err != nil || info.IsDir() {
				continue
			}
			targets = append(targets, f)
		}
		return targets, nil
	}

	var targets []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "vendor", "testdata":
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") {
			targets = append(targets, path)
		}
		return nil
	})
	return targets, err
}

// hasHeader reports whether the SPDX marker already appears in the file's
// leading comment block — the run of blank and // lines before the first line
// of code. Scoping to that block avoids matching the marker where it appears
// later in a file (such as in this tool's own header constant).
func hasHeader(content []byte) bool {
	for _, line := range bytes.Split(content, []byte("\n")) {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			continue
		}
		if !bytes.HasPrefix(trimmed, []byte("//")) {
			return false
		}
		if bytes.Contains(trimmed, []byte(marker)) {
			return true
		}
	}
	return false
}

// prepend places the header and a single blank line before the file's existing
// content. The blank line keeps a following package-doc comment a distinct
// comment group, so it is not absorbed into the header.
func prepend(content []byte) []byte {
	var b bytes.Buffer
	b.WriteString(header)
	b.WriteByte('\n')
	b.Write(content)
	return b.Bytes()
}
