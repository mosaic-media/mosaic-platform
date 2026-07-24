// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

// Command modulesign is the publisher side of extension-module signing (ADR
// 0065): it generates a signing key, computes a binary's digest in the exact
// format the Platform verifies against, and signs a manifest. A module's release
// workflow runs it; the Platform never does — the Platform only verifies.
//
// The digest it prints comes from the same function the Platform hashes with
// (extension.FileDigest), so there is no second definition of the format to
// drift. Getting that format wrong is the one mistake that fails silently at the
// publisher and only surfaces as "signature does not verify" on the far side,
// which is why the tool owns it rather than a README.
//
//	modulesign genkey  -out <path>                  # writes <path> and <path>.pub
//	modulesign digest  <binary>                     # prints sha256:<hex>
//	modulesign sign    -key <path> <manifest.json>  # writes <manifest.json>.sig
//
// The private key is raw ed25519 seed bytes; the public key is the raw public
// key. Neither is armoured — a module publisher's key custody is their concern,
// and a raw key is the least ambiguous thing to hand a secret store.
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"os"

	"github.com/mosaic-media/platform/internal/adapters/extension"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "genkey":
		genkey(os.Args[2:])
	case "digest":
		digest(os.Args[2:])
	case "sign":
		sign(os.Args[2:])
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: modulesign genkey -out <path> | digest <binary> | sign -key <path> <manifest.json>")
	os.Exit(2)
}

func genkey(args []string) {
	out := flagValue(args, "-out")
	if out == "" {
		fail("genkey needs -out <path>")
	}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fail("generating key: %v", err)
	}
	// The private key file holds the seed, from which the full key is derived.
	// 0o600 because it is a secret; the tool refuses to be casual about that.
	if err := os.WriteFile(out, priv.Seed(), 0o600); err != nil {
		fail("writing private key: %v", err)
	}
	if err := os.WriteFile(out+".pub", pub, 0o644); err != nil { //nolint:gosec // a public key is public.
		fail("writing public key: %v", err)
	}
	fmt.Printf("wrote %s (private, keep secret) and %s.pub (trust this in the Platform)\n", out, out)
}

func digest(args []string) {
	if len(args) != 1 {
		fail("digest takes exactly one binary path")
	}
	d, err := extension.FileDigest(args[0])
	if err != nil {
		fail("%v", err)
	}
	fmt.Println(d)
}

func sign(args []string) {
	key := flagValue(args, "-key")
	manifest := lastNonFlag(args)
	if key == "" || manifest == "" {
		fail("sign needs -key <path> and a manifest path")
	}

	seed, err := os.ReadFile(key) //nolint:gosec // the operator names their own key file.
	if err != nil {
		fail("reading key: %v", err)
	}
	if len(seed) != ed25519.SeedSize {
		fail("key is not an ed25519 seed (%d bytes, want %d)", len(seed), ed25519.SeedSize)
	}
	priv := ed25519.NewKeyFromSeed(seed)

	data, err := os.ReadFile(manifest) //nolint:gosec // the operator names their own manifest.
	if err != nil {
		fail("reading manifest: %v", err)
	}
	// Sign the exact bytes, and refuse to sign a manifest that would not parse:
	// signing garbage produces a valid signature over garbage, which fails far
	// away with no clue why.
	if _, err := extension.ParseManifest(data); err != nil {
		fail("refusing to sign an invalid manifest: %v", err)
	}

	sig := ed25519.Sign(priv, data)
	out := manifest + ".sig"
	if err := os.WriteFile(out, sig, 0o644); err != nil { //nolint:gosec // a signature is public.
		fail("writing signature: %v", err)
	}
	fmt.Printf("wrote %s\n", out)
}

// flagValue returns the argument following name, or "".
func flagValue(args []string, name string) string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == name {
			return args[i+1]
		}
	}
	return ""
}

// lastNonFlag returns the last argument that is not a flag or a flag's value —
// the positional manifest path.
func lastNonFlag(args []string) string {
	for i := len(args) - 1; i >= 0; i-- {
		if i > 0 && args[i-1] == "-key" {
			continue
		}
		if args[i] == "-key" {
			continue
		}
		return args[i]
	}
	return ""
}

func fail(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "modulesign: "+format+"\n", a...)
	os.Exit(1)
}
