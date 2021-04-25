package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"syscall"

	"github.com/ghetzel/shmtool/shm"
)

var (
	flagIn      = flag.String("input", "", "input directory")
	flagOut     = flag.String("output", "", "output directory")
	flagDebug   = flag.Bool("debug", false, "debug option, default false")
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

	segment, err = shm.Create(MAP_SIZE)
	if err != nil {
		log.Fatalf("failed to setup shm: %v\n", err)
	}
	defer segment.Destroy()
	segment_addr, err := segment.Attach()
	if err != nil {
		panic(err)
	}
	defer segment.Detach(segment_addr)
	os.Setenv(SHM_ENV_VAR, fmt.Sprint(segment.Id))

	input_segment, err = shm.Create(INPUT_MAP_SIZE)
	if err != nil {
		log.Fatalf("failed to setup shm: %v\n", err)
	}
	defer input_segment.Destroy()
	input_segment_addr, err := input_segment.Attach()
	if err != nil {
		panic(err)
	}
	defer input_segment.Detach(input_segment_addr)
	os.Setenv(INPUT_SHM_ENV_VAR, fmt.Sprint(input_segment.Id))

	for {
		for _, q := range queue {
			file, err := os.Open(q)
			if err != nil {
				panic(err)
			}
			_, err = io.Copy(input_segment, file)
			if err != nil {
				panic(err)
			}
			b := make([]byte, input_segment.Size)
			input_segment.Read(b)
			fuzz_one(run_args, b)
			os.Exit(1)
		}
	}
}

func read_testcases() {
	files, err := ioutil.ReadDir(*flagIn)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if !f.IsDir() {
			queue = append(queue, filepath.Join(*flagIn, f.Name()))
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
	log.Printf("exit code=%d\n", ws.ExitStatus())
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
	b := make([]byte, segment.Size)
	segment.Read(b)
	log.Printf("trace_bits=%v\n", b[:100])
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

func setup_dir_fds() {
	file, err := os.Open("/dev/null")
	if err != nil {
		log.Fatalf("failed to open /dev/null: %v\n", err)
	}
	dev_null_fd := file.Fd()
	defer syscall.Close(int(dev_null_fd))
}
