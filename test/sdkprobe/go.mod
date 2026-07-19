// This is a deliberately separate Go module. It stands in for a Module built
// by someone who is not the Platform team, and it exists to prove ADR 0016's
// property: the published surface (contracts/platform/v1) is importable from
// outside the Platform's own module, and self-contained — an external module
// cannot import anything under internal/, so if a public signature referenced
// an internal type, this module would not compile.
//
// The replace points at the repository root so the probe builds against the
// working tree. Nothing in the main module imports this one; it is compiled
// only by test/sdkboundary.
module example.com/mosaic-sdk-probe

go 1.25.0

require github.com/mosaic-media/mosaic-platform v0.0.0

replace github.com/mosaic-media/mosaic-platform => ../..
