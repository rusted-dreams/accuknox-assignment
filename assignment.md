Dear ______,

Pleasure e-meeting and thank you for your interest in Accuknox.

Accuknox's back-end product/platform uses core Linux kernel primitives such as eBPF, LSMs, seccomp, and iptables/Netfilter to achieve core product goals. It is expected that the members of this team have a flair for using Linux and problem-solving. With that in mind, would you be open to solving a problem statement?

## Problem statement 1: Drop packets using eBPF

Write an eBPF code to drop the TCP packets on a port (def: 4040). Additionally, if you can make the port number configurable from the userspace, that will be a big plus.

I understand that you are new to eBPF, but you can take all the help from your favorite search engine. The aim should be to demo the working code.
Expected Timeline: No timeline as such but will appreciate it if you can revert within a week or two.

## Problem statement 2: Drop packets only for a given process

Write an eBPF code to allow traffic only at a specific TCP port (default 4040) for a given process name (for e.g, "myprocess"). All the traffic to all other ports for only that process should be dropped.

## Problem Statement 3: Explain the code snippet

Explain what the following code is attempting to do? You can explain by:
Explaining how the highlighted constructs work?
Giving use-cases of what these constructs could be used for.
What is the significance of the for loop with 4 iterations?
What is the significance of make(chan func(), 10)?
Why is “HERE1” not getting printed?

```go
package main

import "fmt"

func main() {
    cnp := make(chan func(), 10)
    for i := 0; i < 4; i++ {
        go func() {
            for f := range cnp {
                f()
            }
        }()
    }
    cnp <- func() {
        fmt.Println("HERE1")
    }
    fmt.Println("Hello")
}
```

Create a video demonstration of the actual execution of the eBPF code in terminal mode. Upload the video in a drive and share the drive link (Enable the access for everyone to view the video)

Once you are ready/partially ready, reply to this email, and will schedule a call.

If you have any questions, please feel free to reply.