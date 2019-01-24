// Package cmn provides common low-level types and utilities for all aistore projects
/*
 * Copyright (c) 2018, NVIDIA CORPORATION. All rights reserved.
 */
package cmn

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
	"syscall"
)

// as of 1.9 net/http does not appear to provide any better way..
func IsErrConnectionRefused(err error) (yes bool) {
	if uerr, ok := err.(*url.Error); ok {
		if noerr, ok := uerr.Err.(*net.OpError); ok {
			if scerr, ok := noerr.Err.(*os.SyscallError); ok {
				if scerr.Err == syscall.ECONNREFUSED {
					yes = true
				}
			}
		}
	}
	return
}

// Checks if the error is generated by any IO operation and if the error
// is severe enough to run the FSHC for mountpath testing
//
// for mountpath definition, see fs/mountfs.go
func IsIOError(err error) bool {
	if err == nil {
		return false
	}
	if err == io.ErrShortWrite {
		return true
	}

	isIO := func(e error) bool {
		return e == syscall.EIO || // I/O error
			e == syscall.ENOTDIR || // mountpath is missing
			e == syscall.EBUSY || // device or resource is busy
			e == syscall.ENXIO || // No such device
			e == syscall.EBADF || // Bad file number
			e == syscall.ENODEV || // No such device
			e == syscall.EUCLEAN || // (mkdir)structure needs cleaning = broken filesystem
			e == syscall.EROFS || // readonly filesystem
			e == syscall.EDQUOT || // quota exceeded
			e == syscall.ESTALE || // stale file handle
			e == syscall.ENOSPC // no space left
	}

	switch e := err.(type) {
	case *os.PathError:
		return isIO(e.Err)
	case *os.SyscallError:
		return isIO(e.Err)
	default:
		return false
	}
}

// Check if a given error is a broken-pipe one
// The code is partially borrowed from go-<version>/src/os/pipe_test.go and
// added conversion from net.OpError
func IsErrBrokenPipe(err error) bool {
	if uerr, ok := err.(*url.Error); ok {
		err = uerr
	}
	if nerr, ok := err.(*net.OpError); ok {
		err = nerr.Err
	}
	if serr, ok := err.(*os.SyscallError); ok {
		err = serr.Err
	}

	isBrokenPipe := err == syscall.EPIPE
	if !isBrokenPipe {
		isBrokenPipe = strings.Contains(err.Error(), "broken pipe")
	}

	return isBrokenPipe
}

//===========================================================================
//
// Common errors reusable in API or client
//
//===========================================================================

type InvalidCksumError struct {
	ExpectedHash string
	ActualHash   string
}

func (e InvalidCksumError) Error() string {
	return fmt.Sprintf("Expected Hash: [%s] Actual Hash: [%s]", e.ExpectedHash, e.ActualHash)
}

func NewInvalidCksumError(eHash string, aHash string) InvalidCksumError {
	return InvalidCksumError{
		ActualHash:   aHash,
		ExpectedHash: eHash,
	}
}
