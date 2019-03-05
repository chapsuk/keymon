package keymon

import (
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

var (
	lock = sync.Mutex{}

	patches = make(map[reflect.Value]patch)
)

type (
	patch struct {
		originalBytes []byte
		replacement   *reflect.Value
	}

	value struct {
		_   uintptr
		ptr unsafe.Pointer
	}
)

func getPtr(v reflect.Value) unsafe.Pointer {
	return (*value)(unsafe.Pointer(&v)).ptr
}

func Patch(target, replacement interface{}) {
	t := reflect.ValueOf(target)
	r := reflect.ValueOf(replacement)
	patchValue(t, r)
}

func patchValue(target, replacement reflect.Value) {
	lock.Lock()
	defer lock.Unlock()

	if replacement.Kind() != reflect.Func {
		panic("path should be a function")
	}

	if target.Kind() != reflect.Func {
		panic("target should be a function")
	}

	if target.Type() != replacement.Type() {
		panic(fmt.Sprintf("same type required: %s != %s", target.Type(), replacement.Type()))
	}

	if patch, ok := patches[target]; ok {
		unpatch(target, patch)
	}

	bytes := replaceFunction(*(*uintptr)(getPtr(target)), uintptr(getPtr(replacement)))
	patches[target] = patch{bytes, &replacement}
}

func Unpatch(target interface{}) bool {
	return unpatchValue(reflect.ValueOf(target))
}

func unpatchValue(target reflect.Value) bool {
	lock.Lock()
	defer lock.Unlock()
	patch, ok := patches[target]
	if !ok {
		return false
	}
	unpatch(target, patch)
	delete(patches, target)
	return true
}

func unpatch(target reflect.Value, p patch) {
	copyToLocation(*(*uintptr)(getPtr(target)), p.originalBytes)
}
