package main

import "github.com/ghetzel/shmtool/shm"

const (
	FAULT_NONE   = 0
	FAULT_TMOUT  = 1
	FAULT_CRASH  = 2
	FAULT_ERROR  = 3
	FAULT_NOINST = 4
	FAULT_NOBITS = 5
)

const (
	CHILD_ARG         = "afl_fork_child"
	SHM_ENV_VAR       = "__AFL_SHM_ID"
	INPUT_SHM_ENV_VAR = "__AFL_SHM_ID_INPUT"
	AFL_DEBUG_CHILD   = "AFL_DEBUG_CHILD"
)

const (
	MAP_SIZE       = 64 * 1024
	INPUT_MAP_SIZE = 10 * 1024 * 1024
	FORKSRV_FD     = 198
)

var (
	segment         *shm.Segment
	input_segment   *shm.Segment
	dev_null_fd     int
	fsrv_ctl_fd     int
	fsrv_st_fd      int
	trace_bits      []byte
	input_bits      []byte
	total_execs     uint64
	child_pid       int
	child_timed_out bool
)

var (
	queue []string
)
