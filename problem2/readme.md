# Problem 2: Process TCP connection policy

Shared tools and privileges are listed in the [root README](../README.md#requirements).

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

From `problem2/` in a normal terminal, build and identify the scope:

```sh
clang -O2 -g -target bpfel -c policy.c -o policy.o
go build -o processpolicy.bin .
POLICY_CGROUP="/sys/fs/cgroup$(systemctl --user show ak-problem2.scope -p ControlGroup --value)"
printf '%s\n' "$POLICY_CGROUP"
```

The printed path must match `/sys/fs/cgroup` plus the test shell's `0::` path
and end in `ak-problem2.scope`. Stop if it resolves to `/sys/fs/cgroup` alone;
never attach to the host root or your editor's group.

Inside the test shell, navigate to this repository's `problem2/` and create
the named TCP client. If prompted, replace only an existing demo copy:

```sh
cp -i "$(command -v curl)" ./myprocess
```

## Manual tests

Start these servers in four separate normal terminals:

```sh
python3 -m http.server 4040 --bind 127.0.0.1
# Another terminal:
python3 -m http.server 4041 --bind 127.0.0.1
# Another terminal:
python3 -m http.server 4040 --bind ::1
# Another terminal:
python3 -m http.server 4041 --bind ::1
```

Before attachment, run these inside the scoped shell; all must return HTTP 200:

```sh
./myprocess -4 --noproxy '*' --connect-timeout 3 --max-time 5 -I http://127.0.0.1:4040/
./myprocess -4 --noproxy '*' --connect-timeout 3 --max-time 5 -I http://127.0.0.1:4041/
./myprocess -6 --noproxy '*' --connect-timeout 3 --max-time 5 -I 'http://[::1]:4040/'
./myprocess -6 --noproxy '*' --connect-timeout 3 --max-time 5 -I 'http://[::1]:4041/'
```

In the normal loader terminal, start the policy and leave it running:

```sh
sudo ./processpolicy.bin -cgroup "$POLICY_CGROUP"
```

Repeat the four requests inside the scope: both 4040 requests should succeed;
both 4041 requests should fail immediately, normally with curl error 7.
Verify that a different name remains unaffected in that same shell:

```sh
curl -4 --noproxy '*' --connect-timeout 3 --max-time 5 -I http://127.0.0.1:4041/
curl -6 --noproxy '*' --connect-timeout 3 --max-time 5 -I 'http://[::1]:4041/'
```

Both should return HTTP 200. From a normal terminal outside the test scope,
run the two `./myprocess` 4041 requests: these should also return HTTP 200.
Check `cat /proc/self/cgroup` to confirm this terminal is outside the scope.

In another normal terminal, set `POLICY_CGROUP` to the same verified path and
inspect the hooks:

```sh
sudo bpftool cgroup show "$POLICY_CGROUP"
```

While running, both connect4/connect6 rows should be present. Stop the loader
with Ctrl+C: its rows should disappear and both previously denied requests
should return HTTP 200 from the still-open scoped shell.

Manual IPv4/IPv6 TCP, connected UDP, scope-boundary, and cleanup tests were
previously reported passing; the steps above cover TCP, not the UDP tests.

## Cleanup

Stop the loader with Ctrl+C before exiting the test shell. Stop test servers
as well. Keep generated `policy.o`, `processpolicy.bin`, and `myprocess` out of Git.
