package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

func main() {
	if err := run(); err != nil {
		log.Printf("error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(shutdown)

	iface, err := net.InterfaceByName("ak-host")
	if err != nil {
		return fmt.Errorf("find test interface: %w", err)
	}

	fmt.Println("interface:", iface.Name, "index:", iface.Index)

	spec, err := ebpf.LoadCollectionSpec("filter.o")
	if err != nil {
		return fmt.Errorf("read BPF object failed: %w", err)
	}

	fmt.Println("programs found:", len(spec.Programs))
	programSpec, ok := spec.Programs["filter"]
	if !ok {
		return fmt.Errorf("BPF object does not contain program named filter")
	}

	fmt.Println("program type:", programSpec.Type)

	collection, err := ebpf.NewCollection(spec)
	if err != nil {
		return fmt.Errorf("load BPF collection: %w", err)
	}

	defer collection.Close()

	fmt.Println("bpf collection load success!")

	// Attach the loaded program to test interface's ingress path.
	attached, err := link.AttachXDP(link.XDPOptions{
		Program:   collection.Programs["filter"],
		Interface: iface.Index,
		Flags:     link.XDPGenericMode,
	})

	if err != nil {
		return fmt.Errorf("attach XDP: %w", err)
	}
	defer attached.Close()

	fmt.Println("XDP attached to", iface.Name)

	fmt.Println("press Ctrl+C to exit")

	<-shutdown

	fmt.Println("shutting down")
	return nil
}
