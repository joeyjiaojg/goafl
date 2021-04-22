package main

import (
	"bytes"
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
	_, err = segment.Read(inbuf)
	if err != nil {
		panic(err)
	}

	if *flagDebug {
		log.Printf("executing: %v", os.Args)
	}
	target := exec.Command(binary, os.Args...)
	buffer := bytes.Buffer{}
	buffer.Write(inbuf)
	target.Stdin = &buffer
	target.Stdout = os.Stdout
	target.Stderr = os.Stderr
	err = target.Run()
	if err != nil {
		log.Println("exec failure")
		var shm_id int
		shm_str := os.Getenv(SHM_ENV_VAR)
		if shm_str != "" {
			shm_id, err = strconv.Atoi(shm_str)
			if err != nil {
				panic(err)
			}
		} else {
			panic(fmt.Errorf("env __AFL_SHM_ID is not set"))
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
		b := []byte{0xad, 0xde, 0xe1, 0xfe}
		_, err = segment.Write(b)
		panic(err)
	}

	os.Exit(0)
}
