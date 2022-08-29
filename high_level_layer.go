/*
Unless explicitly stated otherwise all files in this repository are licensed
under the MIT License.
This product includes software developed at Datadog (https://www.datadoghq.com/).
Copyright 2018 Datadog, Inc.
*/

package python3

/*
#include "Python.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"strings"
	"unsafe"
)

//Py_Main : https://docs.python.org/3/c-api/veryhigh.html?highlight=pycompilerflags#c.Py_Main
// "error" will be set if we fail to call "Py_DecodeLocale" on every "args".
func Py_Main(args []string) (int, error) {
	argc := C.int(len(args))
	argv := make([]*C.wchar_t, argc, argc)
	for i, arg := range args {
		carg := C.CString(arg)
		defer C.free(unsafe.Pointer(carg))

		warg := C.Py_DecodeLocale(carg, nil)
		if warg == nil {
			return -1, fmt.Errorf("fail to call Py_DecodeLocale on '%s'", arg)
		}
		// Py_DecodeLocale requires a call to PyMem_RawFree to free the memory
		defer C.PyMem_RawFree(unsafe.Pointer(warg))
		argv[i] = warg
	}

	return int(C.Py_Main(argc, (**C.wchar_t)(unsafe.Pointer(&argv[0])))), nil
}

//PyRun_AnyFile : https://docs.python.org/3/c-api/veryhigh.html?highlight=pycompilerflags#c.PyRun_AnyFile
// "error" will be set if we fail to open "filename".
func PyRun_AnyFile(filename string) (int, error) {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	mode := C.CString("r")
	defer C.free(unsafe.Pointer(mode))

	cfile, err := C.fopen(cfilename, mode)
	if err != nil {
		return -1, fmt.Errorf("fail to open '%s': %s", filename, err)
	}
	defer C.fclose(cfile)

	// C.PyRun_AnyFile is a macro, using C.PyRun_AnyFileFlags instead
	return int(C.PyRun_AnyFileFlags(cfile, cfilename, nil)), nil
}

//PyRun_SimpleFile : https://docs.python.org/3.9/c-api/veryhigh.html?highlight=pycompilerflags#c.PyRun_SimpleFile
// "error" will be set if we fail to open "filename".
func PyRun_SimpleFile(filename string) (int, error) {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	mode := C.CString("rb")
	defer C.free(unsafe.Pointer(mode))

	cfile, err := C.fopen(cfilename, mode)
	if err != nil {
		return -1, fmt.Errorf("fail to open '%s': %v", filename, err)
	}
	defer C.fclose(cfile)

	// For now we have an issue (seg fault) with this function,
	// so let's read the input file ourselves and execute as
	// a string command instead.
	//ret := C.PyRun_SimpleFileExFlags(cfile, cfilename, 1, nil)


	C.fseek(cfile , 0 , C.SEEK_END);
	lSize := C.ftell(cfile);
	C.rewind(cfile);

	cbuffer := C.CString(strings.Repeat("0", int(lSize)))
	if cbuffer == nil {
		return 1, errors.New("memory alloc fails")
	};
	defer C.free(unsafe.Pointer(cbuffer))

	if 1!=C.fread( unsafe.Pointer(cbuffer) , C.ulonglong(lSize), 1 , cfile) {
		return 1, errors.New("entire read fails")
	}

	ret := C.PyRun_SimpleStringFlags(cbuffer, nil)

	return int(ret), nil
}

//PyRun_SimpleString : https://docs.python.org/3/c-api/veryhigh.html?highlight=pycompilerflags#c.PyRun_SimpleString
func PyRun_SimpleString(command string) int {
	ccommand := C.CString(command)
	defer C.free(unsafe.Pointer(ccommand))

	// C.PyRun_SimpleString is a macro, using C.PyRun_SimpleStringFlags instead
	return int(C.PyRun_SimpleStringFlags(ccommand, nil))
}
