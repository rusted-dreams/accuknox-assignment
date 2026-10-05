# AccuKnox assignment


## Requirements

For Problems 1 and 2:

- Native Linux or a Linux VM with permission to load and attach eBPF programs.
  The implementation was tested on Linux x86-64.
- Go **1.27.1 or later**, as declared in the Go modules.
- Clang/LLVM with a BPF target, Linux userspace headers, and libbpf development
  headers (`bpf/bpf_helpers.h` and `bpf/bpf_endian.h`).
- `ip` from iproute2, `curl`, Python 3, and sudo/root access. `ping` is used for
  the Problem 1 control test; `bpftool` is used to inspect Problem 2 attachments.
- Problem 1 needs generic XDP and network-namespace/veth support.
- Problem 2 needs cgroup v2, a running systemd user session for its test scope,
  IPv4/IPv6 connection hooks, and the `bpf_strncmp` helper. Enable IPv6 loopback
  for the IPv6 tests.

Go downloads the pinned cilium/ebpf dependencies during the first build, so that
build needs internet access. The tests themselves use local traffic only.
Problem 3 needs only Go; it does not require root or eBPF.

## Build and test

Start at the repository root and follow the relevant guide:

- [Problem 1: IPv4 ingress TCP port filter](problem1/readme.md) — build, create
  the test network, test default/custom ports, and detach.
- [Problem 2: named-client TCP connection policy](problem2/readme.md) — build,
  create a test cgroup, test IPv4/IPv6 and the process/scope boundaries.
- [Problem 3: Go worker explanation](problem3/problem3_solution.md) — run the original
  snippet and understand the workers, channel buffer, and missing output.

Run loaders from their problem directory because they read the BPF objects
using relative paths. Stop each loader with Ctrl+C to detach its programs.
Do not commit generated `.o` files, binaries, or the copied `myprocess` client.

Use only the documented test interface/cgroup. The loaders do not configure the
host firewall; establish successful baseline connections before testing policy.
These are limited learning examples, not production security controls.
