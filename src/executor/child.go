package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/ghetzel/shmtool/shm"
)

func exec_child() {
	debug := os.Getenv(AFL_DEBUG_CHILD) != ""
	if !debug {
		syscall.Dup2(dev_null_fd, 1)
		syscall.Dup2(dev_null_fd, 2)
	}

	syscall.Close(dev_null_fd)

	binary, err := exec.LookPath(os.Args[0])
	if err != nil {
		panic(err)
	}
	var shm_id int
	shm_str := os.Getenv(INPUT_SHM_ENV_VAR)
	if shm_str != "" {
		shm_id, err = strconv.Atoi(shm_str)
		if err != nil {
			panic(err)
		}
	} else {
		panic(fmt.Errorf("env __AFL_SHM_ID_INPUT is not set"))
	}
	segment, err := shm.Open(int(shm_id))
	if err != nil {
		panic(err)
	}
	segment_addr, err := segment.Attach()
	if err != nil {
		panic(err)
	}
	defer segment.Detach(segment_addr)
	inbuf := make([]byte, segment.Size)
	n, err := segment.Read(inbuf)
	if err != nil {
		panic(err)
	}
	log.Printf("n=%d, input=%v\n", n, inbuf[:10])
	inReader, inWriter, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	os.Stdin = inReader
	_, err = inWriter.Write(inbuf)
	if err != nil {
		inWriter.Close()
		panic(err)
	}

	if *flagDebug {
		log.Printf("executing: %v", os.Args)
	}
	if err := syscall.Exec(binary, os.Args, os.Environ()); err != nil {
		panic(err)
	}

	log.Println("execv failure")
	trace_bits[0] = 0xad
	trace_bits[1] = 0xde
	trace_bits[2] = 0xe1
	trace_bits[3] = 0xfe
	os.Exit(0)
}
