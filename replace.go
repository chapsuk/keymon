package keymon

import (
	"reflect"
	"syscall"
	"unsafe"
)

func rawMemoryAccess(p uintptr, length int) []byte {
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: p,
		Len:  length,
		Cap:  length,
	}))
}

func pageStart(ptr uintptr) uintptr {
	return ptr & ^(uintptr(syscall.Getpagesize() - 1))
}

func replaceFunction(from, to uintptr) (original []byte) {
	jumpData := jmpToFunctionValue(to)
	f := rawMemoryAccess(from, len(jumpData))
	original = make([]byte, len(f))
	copy(original, f)

	copyToLocation(from, jumpData)
	return
}

func copyToLocation(location uintptr, data []byte) {
	f := rawMemoryAccess(location, len(data))
	page := rawMemoryAccess(pageStart(location), syscall.Getpagesize())
	err := syscall.Mprotect(page, syscall.PROT_READ|syscall.PROT_WRITE|syscall.PROT_EXEC)
	if err != nil {
		panic(err)
	}
	copy(f, data[:])
	err = syscall.Mprotect(page, syscall.PROT_READ|syscall.PROT_EXEC)
	if err != nil {
		panic(err)
	}
}
