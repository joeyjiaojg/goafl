package main

func WIFSTOPPED(__status int) bool {
	return WTERMSIG(__status) == 0x7f
}

func WIFSIGNALED(__status int) bool {
	return WTERMSIG(__status+1) >= 2
}

func WTERMSIG(__status int) int {
	return __status & 0x7f
}
