package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"syscall"
)

var (
	flagIn = flag.String("input", "", "input directory")
	flagOut = flag.String("output", "", "output directory")
	flagDebug = flag.Bool("debug", false, "debug option, default false")
	flagOutFile = flag.String("outfile", "", ".cur_input file")
)

func main() {
	debug.SetGCPercent(50)
	flag.Parse()

	if os.Getenv(CHILD_ARG) == CHILD_ARG {
		// Child process
		exec_child()
	}

	// Parent process
	var run_args []string
	for i := range os.Args {
		if os.Args[i] == "--" {
			run_args = append(run_args, os.Args[i+1:]...)
			break
		}
	}

	var err error
	setup_dir_fds()
	read_testcases()
	setup_stdio_file()
	check_binary()

	perform_dry_run()

	trace_bits, err = setup_shm()
	if err != nil {
		log.Fatalf("failed to setup shm: %v\n", err)
	}
	log.Printf("trace_bits=%v\n", len(trace_bits))

	for {
		// TODO: read in buf
		var buf []byte
		fuzz_one(run_args, buf)
	}
}

func read_testcases() {
	files, err := ioutil.ReadDir(*flagIn)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if !f.IsDir() {
			queue = append(queue, f)
		}
	}
}

func setup_stdio_file() {
	if *flagOutFile == "" {
		*flagOutFile = filepath.Join(*flagOut, ".cur_input")
	}
}

func check_binary() {
	// TODO:
}

func perform_dry_run() {
	// TODO:
}

func fuzz_one(run_args []string, buf []byte) {
	// TODO: mutation things
	common_fuzz_stuff(run_args, buf)
}

func common_fuzz_stuff(run_args []string, buf []byte) bool {
	run_target(run_args, buf)
	return false
}

func run_target(run_args []string, buf []byte) int {
	child_pid = fork(run_args)

	proc, err := os.FindProcess(child_pid)
	if err != nil {
		panic(err.Error())
	}
	state, err := proc.Wait()
	if err != nil {
		panic(err.Error())
	}
	ws := state.Sys().(syscall.WaitStatus)
	if !WIFSTOPPED(ws.ExitStatus()) {
		child_pid = 0
	}
	total_execs++
	// TODO: classify counts

	if ws.Signaled() {
		log.Printf("child process exit signal=%d\n", ws.Signal())
		if child_timed_out && ws.Signal() == syscall.SIGKILL {
			return FAULT_TMOUT
		}
		return FAULT_CRASH
	}
	return FAULT_NONE
}

func fork(run_args []string) int {
	if len(run_args) == 0 {
		log.Fatal("no target fuzzing binary provided")
	}
	os.Setenv(CHILD_ARG, CHILD_ARG)
	pid, err := syscall.ForkExec(os.Args[0], run_args, &syscall.ProcAttr{
		Files: []uintptr{0, 1, 2},
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

	log.Printf("executing: %v", os.Args)
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

func setup_dir_fds() {
	file, err := os.Open("/dev/null")
	if err != nil {
		log.Fatalf("failed to open /dev/null: %v\n", err)
	}
	dev_null_fd := file.Fd()
	defer syscall.Close(int(dev_null_fd))
}
