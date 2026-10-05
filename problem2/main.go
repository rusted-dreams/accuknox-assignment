package main

import (
	"flag"
	"fmt"
	"log"
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

	cgroupPath := flag.String("cgroup", "", "path to the isolated test cgroup")
	flag.Parse()

	if *cgroupPath == "" {
		return fmt.Errorf("-cgroup is required")
	}

	spec, err := ebpf.LoadCollectionSpec("policy.o")
	if err != nil {
		return fmt.Errorf("read BPF object: %w", err)
	}

	if _, ok := spec.Programs["allow_connect4"]; !ok {
		return fmt.Errorf("BPF object does not contain program allow_connect4")
	}

	if _, ok := spec.Programs["allow_connect6"]; !ok {
		return fmt.Errorf("BPF object does not contain program allow_connect6")
	}

	collection, err := ebpf.NewCollection(spec)
	if err != nil {
		return fmt.Errorf("load BPF collection: %w", err)
	}
	defer collection.Close()

	// Limit attachment to the explicitly selected test cgroup.
	attached, err := link.AttachCgroup(link.CgroupOptions{
		Path:    *cgroupPath,
		Attach:  ebpf.AttachCGroupInet4Connect,
		Program: collection.Programs["allow_connect4"],
	})
	if err != nil {
		return fmt.Errorf("attach connect4 policy: %w", err)
	}
	defer attached.Close()

	attached6, err := link.AttachCgroup(link.CgroupOptions{
		Path:    *cgroupPath,
		Attach:  ebpf.AttachCGroupInet6Connect,
		Program: collection.Programs["allow_connect6"],
	})
	if err != nil {
		return fmt.Errorf("attach connect6 policy: %w", err)
	}
	defer attached6.Close()

	fmt.Println("IPv4 and IPv6 connection policy attached to", *cgroupPath)
	fmt.Println("press Ctrl+C to exit")
	<-shutdown

	fmt.Println("shutting down")
	return nil
}
