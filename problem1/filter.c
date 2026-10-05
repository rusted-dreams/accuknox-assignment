//go:build ignore

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>
#include <linux/if_ether.h>
#include <bpf/bpf_endian.h>
#include <linux/ip.h>
#include <linux/in.h>
#include <linux/tcp.h>

// #define BLOCKED_PORT 4040
struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, __u16);
} config SEC(".maps");


// Include both the More Fragments flag and fragment offset bits.
#define IPV4_FRAGMENT_MASK 0x3fff

SEC("xdp")
int filter(struct xdp_md* ctx) {
    void* data = (void*)(long)ctx->data;
    void* data_end = (void*)(long)ctx->data_end;

    struct ethhdr* eth = data;

    // The verifier needs bounds proof before each header read.
    if ((void*)(eth + 1) > data_end)
        return XDP_PASS;

    // Convert packet fields to host byte order before comparing.
    if (bpf_ntohs(eth->h_proto) != ETH_P_IP)
        return XDP_PASS;

    struct iphdr* ip = (void*)(eth + 1);

    if ((void*)(ip + 1) > data_end)
        return XDP_PASS;

    if (ip->version != 4 || ip->ihl < 5)
        return XDP_PASS;

    // IHL includes options and uses four-byte units.
    __u32 ip_header_len = (__u32)ip->ihl * 4;

    if ((void*)ip + ip_header_len > data_end)
        return XDP_PASS;

    if (ip->protocol != IPPROTO_TCP)
        return XDP_PASS;

    // Fragments may not contain a complete TCP header.
    if (bpf_ntohs(ip->frag_off) & IPV4_FRAGMENT_MASK)
        return XDP_PASS;

    __u32 ip_total_len = bpf_ntohs(ip->tot_len);

    // Do not mistake Ethernet padding for TCP bytes.
    if (ip_total_len < ip_header_len + sizeof(struct tcphdr))
        return XDP_PASS;

    if ((void*)ip + ip_total_len > data_end)
        return XDP_PASS;

    // TCP starts after the IP header. ihl is in 4-byte units and covers options.
    // This bounds check also proves the IP options fit in the packet.
    struct tcphdr* tcp = (void*)ip + ip_header_len;

    if ((void*)(tcp + 1) > data_end)
        return XDP_PASS;

    __u32 key = 0;
    __u16* blocked_port = bpf_map_lookup_elem(&config, &key);

    // Preserve fail-open behavior if configuration is unavailable.
    if (!blocked_port)
        return XDP_PASS;

    if (bpf_ntohs(tcp->dest) == *blocked_port)
        return XDP_DROP;

    return XDP_PASS;
}

char LICENSE[] SEC("license") = "Dual MIT/GPL";
