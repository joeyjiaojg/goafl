package main

import (
	"log"
	"os"
	"os/exec"
	"runtime/debug"
	"syscall"
)

const (
	MAP_SIZE    = 64 * 1024
	CHILD_ARG   = "child"
	SHM_ENV_VAR = "__AFL_SHM_ID"
)

// TODO: ctrl pipe

func main() {
	debug.SetGCPercent(50)

	if len(os.Args) >= 2 && os.Args[1] == CHILD_ARG {
		var args []string
		if os.Args[2] == "--" {
			args = os.Args[3:]
		} else {
			args = os.Args[2:]
		}
		if len(args) == 0 {
			log.Fatal("no target fuzzing binary provided")
		}
		binary, err := exec.LookPath(args[0])
		if err != nil {
			panic(err)
		}
		log.Printf("child shm_id=%v\n", os.Getenv(SHM_ENV_VAR))

		log.Printf("executing: %v", args)
		if err := syscall.Exec(binary, args, os.Environ()); err != nil {
			panic(err)
		}
	}

	trace_bits, err := setup_shm()
	if err != nil {
		log.Fatalf("failed to setup shm: %v\n", err)
	}
	log.Printf("trace_bits=%v\n", len(trace_bits))

	var args []string
	args = append(args, os.Args[0], CHILD_ARG)
	args = append(args, os.Args[1:]...)
	pid, err := syscall.ForkExec(os.Args[0], args, &syscall.ProcAttr{
		Files: []uintptr{0, 1, 2},
		Env:   os.Environ(),
	})
	if err != nil {
		panic(err.Error())
	}
	log.Printf("fork child pid = %v\n", pid)
	proc, err := os.FindProcess(pid)
	if err != nil {
		panic(err.Error())
	}
	state, err := proc.Wait()
	if err != nil {
		panic(err.Error())
	}
	status := state.Sys().(syscall.WaitStatus)
	if status.Signaled() {
		log.Printf("child process exit signal=%d\n", status.Signal())
	} else {
		log.Printf("child process exit status=%d\n", status.ExitStatus())
	}
	log.Println("Goodbye.")
}
