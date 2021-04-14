package main

const (
	FAULT_NONE   = 0
	FAULT_TMOUT  = 1
	FAULT_CRASH  = 2
	FAULT_ERROR  = 3
	FAULT_NOINST = 4
	FAULT_NOBITS = 5
)

const (
	CHILD_ARG       = "afl_fork_child"
	SHM_ENV_VAR     = "__AFL_SHM_ID"
	AFL_DEBUG_CHILD = "AFL_DEBUG_CHILD"
)

const (
	MAP_SIZE   = 64 * 1024
	FORKSRV_FD = 198
)

var (
	dev_null_fd     int
	fsrv_ctl_fd     int
	fsrv_st_fd      int
	trace_bits      []byte
	total_execs     uint64
	child_pid       int
	child_timed_out bool
)
