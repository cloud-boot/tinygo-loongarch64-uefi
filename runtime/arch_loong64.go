// This file is the proposed addition to TinyGo's
// src/runtime/arch_loong64.go. It mirrors arch_riscv64.go (same
// intent, same fields, LoongArch LP64D-specific values).
//
// Go 1.21+ ships GOARCH=loong64 for LoongArch 64-bit. Once TinyGo
// has this file, our UEFI target spec
// targets/uefi-loongarch64.json (goos=linux, goarch=loong64,
// gc=leaking, scheduler=none, libc="") will compile and produce a
// PE/COFF object suitable for go-coff/peln to link into
// BOOTLOONGARCH64.EFI.

//go:build loong64

package runtime

const GOARCH = "loong64"

// TargetBits is the address width of the CPU. 64 for LoongArch64,
// matching arm64 / amd64 / riscv64.
const TargetBits = 64

// deferExtraRegs is the number of extra callee-saved registers a defer
// frame needs to spill on this arch. arm64 and riscv64 save 0 (the
// existing callee-saved set is already captured by the runtime), and
// the LoongArch LP64D ABI exposes the same callee-saved register
// class (s0..s8 + fp) that TinyGo's defer mechanism doesn't need to
// spill separately.
const deferExtraRegs = 0

// callInstSize is the size in bytes of a single call instruction. The
// runtime walks back from the return address by this many bytes to
// find the start of the call. LoongArch uses `bl` (a single 32-bit
// instruction, ±128 MiB) for direct calls; far calls use the
// `pcalau12i + jirl` pair (8 bytes) but the return address still
// points immediately after `jirl`, so the conservative value of 4
// matches the existing riscv64 convention.
const callInstSize = 4

// Linux signal / mmap constants. LoongArch Linux uses the *generic*
// values (asm-generic/mman.h, asm-generic/signal.h) — same as arm64 /
// x86_64 / riscv64 for these particular constants.
const (
	linux_MAP_ANONYMOUS = 0x20
	linux_SIGBUS        = 7
	linux_SIGILL        = 4
	linux_SIGSEGV       = 11
)

// align rounds ptr up to the next 16-byte boundary. LoongArch ELF
// psABI v2 §4 "Stack Alignment" mandates 16-byte stack alignment at
// function entry, same as arm64 and RV64.
func align(ptr uintptr) uintptr {
	return (ptr + 15) &^ 15
}

// getCurrentStackPointer returns the current SP register value. Uses
// LLVM's @llvm.stacksave intrinsic via TinyGo's stacksave() shim —
// arch-agnostic at the IR level, so the loong64 backend lowers it to
// a trivial `move result, $sp`.
func getCurrentStackPointer() uintptr {
	return uintptr(stacksave())
}
