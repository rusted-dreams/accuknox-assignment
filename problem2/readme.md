# Problem 2: Process TCP connection policy

## Assumptions

- Treat `myprocess` as a client: restrict new outbound TCP connections, not
  incoming connections, existing connections, or individual packets.
- Within the selected cgroup, allow `myprocess` to connect only to destination
  port `4040`, for IPv4 and IPv6. Other names and non-TCP traffic are unaffected.
- Match Linux thread name (`comm`), not executable path or command-line arguments.
  Names are mutable, thread-specific, and limited to 15 characters. This is a
  demo identity check, not a production security boundary.
- Create test sockets after entering the cgroup. Passed-in or previously
  created sockets are outside the demonstrated assumptions.

## Approach

Attach `cgroup/connect4` and `cgroup/connect6` programs to an isolated test
cgroup. Each checks the protocol, reads `comm`, matches `myprocess`, and compares
the destination port in host byte order. Return `1` to allow or `0` to reject;
name-read failure allows the attempt (fail-open).

Go loads `policy.o`, attaches both programs, and waits for SIGINT/SIGTERM.
Deferred cleanup detaches both links before closing the collection. The two
C functions intentionally duplicate the short policy; keep their rules aligned.

## Setup and run

Requires Linux with cgroup v2 and the required BPF hooks/helpers, systemd,
Go compatible with `go.mod`, Clang with BPF support, Linux/libbpf headers,
and privileges to load/attach BPF. Tests use Python 3 and curl.

Create a test shell in a separate terminal and leave it open:

```sh
systemd-run --user --scope --unit=ak-problem2 bash
cat /proc/self/cgroup
```

From `problem2/` in a normal terminal, build and run. Set `-cgroup` to
`/sys/fs/cgroup` followed by the path printed after `0::` above; do not target
the cgroup root or your editor's group. Example for the tested user:

```sh
clang -O2 -g -target bpfel -c policy.c -o policy.o
go build -o processpolicy.bin .
sudo ./processpolicy.bin -cgroup /sys/fs/cgroup/user.slice/user-1000.slice/user@1000.service/app.slice/ak-problem2.scope
```

Inside the test shell, go to `problem2/` and create the named TCP test client:

```sh
cp -i "$(command -v curl)" ./myprocess
```

With reachable test servers, `myprocess` should connect to TCP `4040` but fail
on `4041`; ordinary curl should reach both. For IPv6 use `-6` and URLs such as
`http://[::1]:4040/`. Outside the test cgroup, `myprocess` remains unrestricted
by this policy. Manual IPv4/IPv6 TCP, connected UDP, scope-boundary, and cleanup
tests passed; these are not an automated test suite.

Stop the loader with Ctrl+C before exiting the test shell. Stop test servers
as well. Keep generated `policy.o`, `processpolicy.bin`, and `myprocess` out of Git.
