# Problem 1: TCP port filter

Drops incoming IPv4 TCP packets for a destination port using generic XDP.
Go supplies the port through a one-entry BPF array map.

Shared tools and privileges are listed in the [root README](../README.md#requirements).
## Scope

- Default port: `4040`. Override with `-port` (`1–65535`); restart to change it.
- Ingress only, on the attached interface; no egress or process-based filtering.
- UDP, ICMP, IPv6, VLAN-tagged traffic, and IPv4 fragments pass through.
- Parser-check or map-lookup failures pass traffic (fail-open).
- Learning demo, not a production firewall.

## Build and run

Requires Linux with BPF/generic XDP support, Go, Clang with BPF support,
Linux/libbpf headers, and privileges to load BPF.

From `problem1/`, with the test interface already created:

```sh
clang -O2 -g -target bpfel -c filter.c -o filter.o
go build -o portfilter.bin .
sudo ./portfilter.bin                  # block TCP 4040
# Alternatively:
sudo ./portfilter.bin -port 4041       # block TCP 4041
```

Run only one loader at a time. Ctrl+C detaches the filter. Run from this
directory because the loader reads `filter.o` using a relative path.
Do not commit `filter.o` or `portfilter.bin`.

## Test network

The demo uses `ak-host` and `ak-peer` as an isolated veth pair. The current Go
loader hard-codes `ak-host`; using another interface requires changing that
lookup in `main.go` and rebuilding. There is no interface CLI option yet.

Create once; skip if already configured:

```sh
sudo ip netns add ak-client
sudo ip link add ak-host type veth peer name ak-peer
sudo ip link set ak-peer netns ak-client
sudo ip addr add 10.200.1.1/30 dev ak-host
sudo ip -n ak-client addr add 10.200.1.2/30 dev ak-peer
sudo ip link set ak-host up
sudo ip -n ak-client link set lo up
sudo ip -n ak-client link set ak-peer up
```

Start HTTP servers on host ports `4040` and `4041` in separate terminals:

```sh
python3 -m http.server 4040 --bind 10.200.1.1
# In another terminal:
python3 -m http.server 4041 --bind 10.200.1.1
```

Before starting the loader, test both servers from the namespace:

```sh
sudo ip netns exec ak-client curl -4 --noproxy '*' --connect-timeout 3 --max-time 5 -I http://10.200.1.1:4040/
sudo ip netns exec ak-client curl -4 --noproxy '*' --connect-timeout 3 --max-time 5 -I http://10.200.1.1:4041/
```

Both must return HTTP 200 before continuing. If UFW is active and blocks this
test, inspect its rules and add only the missing test exception:

```sh
sudo ufw status
sudo ufw allow in on ak-host proto tcp from 10.200.1.2 to 10.200.1.1 port 4040,4041
```

Do not disable the firewall. Repeat the baseline after fixing reachability.

In the loader terminal, run `sudo ./portfilter.bin`. Repeat both curl commands:
4040 should time out (curl error 28), while 4041 returns HTTP 200. Check that
non-TCP traffic still works:

```sh
sudo ip netns exec ak-client ping -c 1 -W 2 10.200.1.1
```

Press Ctrl+C in the loader terminal; 4040 should return HTTP 200 again.
Restart with `sudo ./portfilter.bin -port 4041`: now 4040 should work and 4041
should time out. Stop the loader and confirm 4041 recovers too.

Manual tests reported passing: default/custom TCP filtering, invalid-port
rejection, ICMP, UDP, and shutdown recovery. Parser edge cases remain untested.

## Cleanup

Stop the loader and test servers. Delete only UFW exceptions added for this
demo (skip rules you did not add or that are already removed):

```sh
sudo ufw delete allow in on ak-host proto udp from 10.200.1.2 to 10.200.1.1 port 4040
sudo ufw delete allow in on ak-host proto tcp from 10.200.1.2 to 10.200.1.1 port 4040,4041
```

When no longer needed, stop namespace processes and remove only this demo's
network resources:

```sh
sudo ip link delete ak-host
sudo ip netns delete ak-client
```
