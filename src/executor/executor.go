package main

import (
	"encoding/binary"
	"log"
	"os"
	"os/exec"
	"runtime/debug"
	"syscall"
)

func main() {
	debug.SetGCPercent(50)

	if len(os.Args) >= 2 && os.Args[1] == CHILD_ARG {
		exec_child()
	}

	// Parent process
	var err error
	setup_dir_fds()
	trace_bits, err = setup_shm()
	if err != nil {
		log.Fatalf("failed to setup shm: %v\n", err)
	}
	log.Printf("trace_bits=%v\n", len(trace_bits))

	init_forkserver()

	for {
		fuzz_one()
	}
}

func fuzz_one() {
	// TODO: mutation things
	common_fuzz_stuff()
}

func common_fuzz_stuff() {
	run_target()
}

func run_target() int {
	tmp := []byte{0, 0, 0, 0}
	len, err := syscall.Write(fsrv_ctl_fd, tmp)
	if err != nil || len != 4 {
		log.Fatal("unable to request new process from for server (OOM?)")
	}
	len, err = syscall.Read(fsrv_st_fd, tmp)
	if err != nil || len != 4 {
		log.Fatal("unable to get child pid.")
	}
	child_pid := int(binary.LittleEndian.Uint32(tmp))
	if child_pid < 0 {
		log.Fatal("fork server is misbehaving.")
	}
	// TODO: timeout
	len, err = syscall.Read(fsrv_st_fd, tmp)
	if err != nil || len != 4 {
		log.Fatal("unable to communicate with fork server.")
	}

	status := int(binary.LittleEndian.Uint32(tmp))
	if !WIFSTOPPED(status) {
		child_pid = 0
	}
	total_execs++

	if WIFSIGNALED(status) {
		// TODO: return FAULT_TMOUT
		// kill_signal = WTERMSIG(status)

		return FAULT_CRASH
	}

	return FAULT_NONE
	// proc, err := os.FindProcess(child_pid)
	// if err != nil {
	// 	panic(err.Error())
	// }
	// state, err := proc.Wait()
	// if err != nil {
	// 	panic(err.Error())
	// }
	// res := state.Sys().(syscall.WaitStatus)
	// res.Stopped()
	// if res.Signaled() {
	// 	log.Printf("child process exit signal=%d\n", res.Signal())
	// } else {
	// 	log.Printf("child process exit status=%d\n", res.ExitStatus())
	// }
}

func init_forkserver() {
	var st_pipe []int
	var ctl_pipe []int
	if err := syscall.Pipe(st_pipe); err != nil {
		log.Fatalf("pipe() failed: %v\n", err)
	}
	if err := syscall.Pipe(ctl_pipe); err != nil {
		log.Fatalf("pipe() failed: %v\n", err)
	}

	pid := fork(ctl_pipe, st_pipe)

	syscall.Close(ctl_pipe[0])
	syscall.Close(st_pipe[1])

	fsrv_ctl_fd = ctl_pipe[1]
	fsrv_st_fd = st_pipe[0]

	var status []byte
	// TODO: timeout for read
	rlen, err := syscall.Read(fsrv_st_fd, status)
	if err != nil {
		log.Fatalf("faield to start fork server: %v\n", err)
	}
	if rlen == 4 {
		log.Println("All right - fork server is up.")
		return
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		panic(err.Error())
	}
	state, err := proc.Wait()
	if err != nil {
		panic(err.Error())
	}
	res := state.Sys().(syscall.WaitStatus)
	if res.Signaled() {
		log.Fatal("the target binary crashed.")
	}

	log.Fatal("fork server handshake failed.")
}

func fork(ctl_pipe, st_pipe []int) int {
	var args []string
	args = append(args, os.Args[0], CHILD_ARG)
	args = append(args, os.Args[1:]...)
	pid, err := syscall.ForkExec(os.Args[0], args, &syscall.ProcAttr{
		Files: []uintptr{0, 1, 2, uintptr(ctl_pipe[0]), uintptr(st_pipe[1])},
		Env:   os.Environ(),
	})
	if err != nil {
		panic(err.Error())
	}
	if pid < 0 {
		panic("fork() failed")
	}
	return pid
}

func exec_child() {
	// Child process
	debug := os.Getenv(AFL_DEBUG_CHILD) != ""
	if !debug {
		syscall.Dup2(dev_null_fd, 1)
		syscall.Dup2(dev_null_fd, 2)
	}
	if err := syscall.Dup2(ctl_pipe[0], FORKSRV_FD); err != nil {
		log.Fatal("dup2() failed")
	}
	if err := syscall.Dup2(st_pipe[1], FORKSRV_FD+1); err != nil {
		log.Fatal("dup2() failed")
	}
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

func setup_dir_fds() {
	file, err := os.Open("/dev/null")
	if err != nil {
		log.Fatalf("failed to open /dev/null: %v\n", err)
	}
	defer file.Close()
	dev_null_fd := file.Fd()
}
