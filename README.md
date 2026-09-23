# tinygo-loongarch64-uefi

Adds the missing TinyGo runtime support needed to compile a
`goarch=loong64` baremetal UEFI binary — the prerequisite for ever
shipping `BOOTLOONGARCH64.EFI` out of the
[`go-coff/stub`](https://github.com/go-coff/stub) project.

This project does **not** build TinyGo from source. It ships one Go
file that, once installed into TinyGo's runtime, lets `tinygo build`
succeed for the
[`uefi-loongarch64.json`](https://github.com/go-coff/stub/blob/main/targets/uefi-loongarch64.json)
target. The intent is that this file lands upstream in TinyGo — see
[Upstreaming](#upstreaming) below.

## Status

| Stage                                 | State                                                            |
| ------------------------------------- | ---------------------------------------------------------------- |
| TinyGo runtime compile (Go side)      | ✅ resolved by `runtime/arch_loong64.go` in this project          |
| go-coff stub `tinygo build` succeeds  | ☐ pending verification (toolchain not available in this session) |
| `BOOTLOONGARCH64.EFI` link from COFF  | ✅ via `go-coff/peln` (`0x6264`); `lld-link` bypassed             |
| QEMU boot test                        | ☐ pending stub Taskfile wire-up + `qemu-system-loongarch64`      |

> **Note** — the `lld-link → BOOTLOONGARCH64.EFI` step is **not** the
> blocker. `lld-link`'s COFF driver has no `/machine:loongarch64`, but
> the cloud-boot toolchain doesn't use it: the
> [`go-coff/peln`](https://github.com/go-coff/peln) library is our pure-Go COFF/PE
> linker, supporting amd64 (`0x8664`), arm64 (`0xAA64`), riscv64
> (`0x5064`), and loongarch64 (`0x6264`), with the LoongArch
> relocation table (`R_LARCH_NONE`, `_32`, `_64`, `_RELATIVE`,
> `_MARK_LA`, `_MARK_PCREL`, `_B16`, `_B21`, `_B26`, `_ABS_HI20`,
> `_ABS_LO12`, `_ABS64_LO20`, `_ABS64_HI12`, `_PCALA_HI20`,
> `_PCALA_LO12`, `_PCALA64_LO20`, `_PCALA64_HI12`, `_RELAX`, `_ALIGN`)
> unit-tested in
> [peln/linker/reloc_loongarch64_test.go](https://github.com/go-coff/peln/blob/main/linker/reloc_loongarch64_test.go).
> Invoke it via `pectl link --machine loongarch64 …`.

## The TinyGo gap

TinyGo's `src/runtime/` carries one `arch_<GOARCH>.go` per supported
standard Go architecture (`arch_amd64.go`, `arch_arm64.go`,
`arch_riscv64.go`, …). The file declares per-arch constants (`GOARCH`,
`TargetBits`, `callInstSize`, `deferExtraRegs`), Linux ABI numbers
(`linux_MAP_ANONYMOUS`, `linux_SIG*`), and two helpers (`align`,
`getCurrentStackPointer`). The rest of the runtime
(`gc_leaking.go`, `os_linux.go`, `panic.go`, `print.go`,
`runtime_unix.go`) uses these symbols unconditionally; without the
per-arch file they show up as `undefined`.

Go 1.21+ ships `GOARCH=loong64` (LoongArch 64-bit), but TinyGo as of
mid-2026 doesn't carry a matching `arch_loong64.go`. Our UEFI target
spec uses the standard `goarch=loong64`, so it picks no
`arch_*.go` from TinyGo's tree → `undefined` symbols cascade across
the whole runtime.

[`runtime/arch_loong64.go`](runtime/arch_loong64.go) fills that gap.
It mirrors `arch_riscv64.go` (same intent, same fields, LoongArch
LP64D-specific values).

## Constants — sourcing

| Symbol | LoongArch LP64D value | Source |
| --- | --- | --- |
| `TargetBits` | `64` | LoongArch ELF psABI v2 §1 |
| `callInstSize` | `4` | `bl` opcode is a single 32-bit insn |
| `deferExtraRegs` | `0` | Same callee-saved class as RV64/AArch64 |
| `linux_MAP_ANONYMOUS` | `0x20` | `asm-generic/mman.h` (LoongArch follows generic) |
| `linux_SIGBUS` / `_SIGILL` / `_SIGSEGV` | 7 / 4 / 11 | `asm-generic/signal.h` |

`align(ptr)` rounds up to 16 bytes — per the LoongArch psABI §4
"Stack Alignment", the SP must be 16-byte aligned at function entry,
same as RV64 and AArch64.

`getCurrentStackPointer()` uses TinyGo's `stacksave()` shim, which
the LoongArch LLVM backend lowers to a trivial `move result, $sp`.

## Usage

Same flow as the riscv64 sibling project:

```sh
# from inside this directory
task patch:apply     # copy arch_loong64.go into the local tinygo cached goroot
task verify          # apply, run `tinygo build` from go-coff/stub, restore
task patch:restore   # undo the patch from the cache
task patch:diff      # emit a unified diff suitable for a TinyGo upstream PR
```

The TinyGo cached goroot path is auto-discovered via `tinygo info`.

## Upstreaming

The end state is to land `arch_loong64.go` in TinyGo proper
(`src/runtime/arch_loong64.go`). The shape mirrors the existing
`arch_riscv64.go` exactly — same fields, LoongArch-specific values.
Once accepted, this repo becomes a historical archive.

PR target: <https://github.com/tinygo-org/tinygo>.

Open question for the PR: should `getCurrentStackPointer` use the
`stacksave()` IR intrinsic (as we do) or hand-written inline
assembly (`move ${result}, $sp`)? The IR-intrinsic path matches what
`arch_riscv64.go` does in the same project, so the proposed PR
matches the existing convention.
