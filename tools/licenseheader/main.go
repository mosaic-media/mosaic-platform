// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

// Command licenseheader adds (or verifies) the SPDX license header on every Go
// source file in the repository. It exists so the header is applied uniformly
// by a tool rather than by each contributor's editor.
//
// Usage:
//
//	go run ./tools/licenseheader          # add the header to any file missing it
//	go run ./tools/licenseheader -check   # list files missing it; exit non-zero if any
//
// The -check form is what a pre-commit hook or CI step runs to keep the header
// from silently drifting. The tool is idempotent: a file that already carries
// the SPDX marker in its leading comment block is left untouched.
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
	root := flag.String("root", ".", "directory to walk")
	flag.Parse()

	var missing, added []string
	err := filepath.WalkDir(*root, func(path string, d fs.DirEntry, err error) error {
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
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if hasHeader(content) {
			return nil
		}
		if *check {
			missing = append(missing, path)
			return nil
		}
		if err := os.WriteFile(path, prepend(content), 0o644); err != nil {
			return err
		}
		added = append(added, path)
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "licenseheader:", err)
		os.Exit(2)
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
