package main

import (
	"fmt"
	"log"

	"github.com/cilium/ebpf"
)

func main() {
	spec, err := ebpf.LoadCollectionSpec("filter.o")
	if err != nil {
		log.Fatalf("read BPF object failed: %v", err)
	}
	fmt.Println("programs found:", len(spec.Programs))

	// find the program with c func. named filter
	programSpec, ok := spec.Programs["filter"]
	if !ok {
		log.Fatal("BPF object does not contain program named filter")
	}

	fmt.Println("program type:", programSpec.Type)

	// load the objects programs into kernel
	// kernel verifier works here during this process
	collection, err := ebpf.NewCollection(spec)
	if err != nil {
		log.Fatalf("load BPF collection: %v", err)
	}

	defer collection.Close()

	fmt.Println("bpf collection load success!")

}
