//go:build ignore

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>
#include <linux/in.h>
#include <bpf/bpf_endian.h>

SEC("cgroup/connect4")
int allow_connect4(struct bpf_sock_addr* ctx) {
    if (ctx->protocol != IPPROTO_TCP)
        return 1;

    char comm[16] = { 0 };

    // Avoid restricting an unidentified caller.
    if (bpf_get_current_comm(comm, sizeof(comm)) != 0)
        return 1;
    // Leave other process names unaffected.
    if (bpf_strncmp(comm, sizeof(comm), "myprocess") != 0)
        return 1;

    // Restrict the target's outbound TCP destination port.
    return bpf_ntohs(ctx->user_port) == 4040;
}

char LICENSE[] SEC("license") = "Dual MIT/GPL";