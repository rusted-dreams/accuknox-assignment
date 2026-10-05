compile the filter.c file:
```bash
clang -O2 -g -target bpfel -c filter.c -o filter.o
```

build the go program:
```bash
go build -o portfilter.bin .
```

create a virtual network interface for testing
```bash
sudo ip netns add ak-client


sudo ip link add ak-host type veth peer name ak-peer
#`ip link add ak-host`: create an interface named `ak-host`.
#`type veth`: make it a virtual Ethernet interface.
#`peer name ak-peer`: create its connected partner named `ak-peer`.

# now move the client end:
sudo ip link set dev ak-peer netns ak-client

# assign ipv4 address to each end:
sudo ip addr add 10.200.1.1/30 dev ak-host

sudo ip -n ak-client addr add 10.200.1.2/30 dev ak-peer

#bring up both the host and client
sudo ip link set dev ak-host up
sudo ip -n ak-client link set dev ak-peer up

# test the virtual network:
sudo ip netns exec ak-client ping -c 3 -W 2 10.200.1.1

```