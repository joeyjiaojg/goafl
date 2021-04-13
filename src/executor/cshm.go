package main

/*
#include <stdlib.h>
#include <sys/types.h>
#include <sys/ipc.h>
#include <sys/shm.h>

#define MAP_SIZE 64*1024

int go_shmget() {
  return shmget(IPC_PRIVATE, MAP_SIZE, IPC_CREAT | IPC_EXCL | 0600);
}

char* go_shmat(int shm_id) {
  return (char*)shmat(shm_id, NULL, 0);
}

void go_shmclose(int shm_id) {
	 shmctl(shm_id, IPC_RMID, NULL);
}
*/
import "C"
import (
	"fmt"
	"log"
	"os"
	"unsafe"
)

func setup_shm() ([]byte, error) {
	shm_id := int(C.go_shmget())
	if shm_id < 0 {
		return nil, fmt.Errorf("failed to run shmget")
	}
	defer C.go_shmclose(C.int(shm_id))
	os.Setenv(SHM_ENV_VAR, fmt.Sprint(shm_id))
	log.Printf("shm_id=%v\n", shm_id)
	addr, err := C.go_shmat(C.int(shm_id))
	if err != nil {
		return nil, fmt.Errorf("failed to run shmat: %v", err)
	}
	trace_bits := C.GoBytes(unsafe.Pointer(addr), MAP_SIZE)

	return trace_bits, nil
}
