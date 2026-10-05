# Problem 1: TCP port filter

Drops incoming IPv4 TCP packets for a destination port using generic XDP.
Go supplies the port through a one-entry BPF array map.

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
sudo ip -n ak-client link set ak-peer up
```

Start HTTP servers on host ports `4040` and `4041` in separate terminals:

```sh
python3 -m http.server 4040 --bind 10.200.1.1
# In another terminal:
python3 -m http.server 4041 --bind 10.200.1.1
```

Test from the namespace; repeat with port `4041`:

```sh
sudo ip netns exec ak-client curl --noproxy '*' --connect-timeout 3 --max-time 5 -I http://10.200.1.1:4040/
```

Without XDP, both ports should return `200 OK`. With XDP, only the selected
port should time out. After Ctrl+C, both should work again. Ensure the host
firewall allows test traffic before attributing timeouts to XDP.

Manual tests reported passing: default/custom TCP filtering, invalid-port
rejection, ICMP, UDP, and shutdown recovery. Parser edge cases remain untested.

## Cleanup

Stop the loader and test servers. Delete only UFW exceptions added for this
demo (skip rules already removed):

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
