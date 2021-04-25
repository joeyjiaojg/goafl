package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
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

	log.Printf("executing: %v", os.Args)
	buffer := bytes.Buffer{}
	buffer.Write(inbuf)
	target := exec.Command(binary, os.Args...)
	target.Stdin = &buffer
	target.Stdout = os.Stdout
	target.Stderr = os.Stderr
	target.Env = os.Environ()
	err = target.Run()
	if err != nil {
		log.Println("traget crash")
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGABRT)
		s := <-c
		fmt.Println("got signal:", s)
		// var shm_id int
		// shm_str := os.Getenv(SHM_ENV_VAR)
		// if shm_str != "" {
		// 	shm_id, err = strconv.Atoi(shm_str)
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// } else {
		// 	panic(fmt.Errorf("env __AFL_SHM_ID is not set"))
		// }
		// segment, err := shm.Open(int(shm_id))
		// if err != nil {
		// 	panic(err)
		// }
		// segment_addr, err := segment.Attach()
		// if err != nil {
		// 	panic(err)
		// }
		// defer segment.Detach(segment_addr)
		// b := []byte{0xad, 0xde, 0xe1, 0xfe}
		// _, err = segment.Write(b)
		log.Printf("exit=%d\n", target.ProcessState.ExitCode())
		// os.Exit(target.ProcessState.ExitCode())
	}
}
