package main

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

var ErrTimerfdCancelled = errors.New("timerfd was cancelled due to discontinuous change")

type Timerfd struct {
	fd *os.File
}

func newTimerfd(clockid int) (*Timerfd, error) {
	fd, err := unix.TimerfdCreate(clockid, unix.TFD_NONBLOCK|unix.TFD_CLOEXEC)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(fd), "timerfd")
	return &Timerfd{fd: f}, nil
}

func (t *Timerfd) Settime(newValue *unix.ItimerSpec, oldValue *unix.ItimerSpec, absolute bool, cancelOnSet bool) error {
	rawConn, err := t.fd.SyscallConn()
	if err != nil {
		return err
	}
	var err2 error
	err = rawConn.Control(func(fd uintptr) {
		var flags int
		if absolute {
			flags |= unix.TFD_TIMER_ABSTIME
		}
		if cancelOnSet {
			flags |= unix.TFD_TIMER_CANCEL_ON_SET
		}
		err2 = unix.TimerfdSettime(int(fd), flags, newValue, oldValue)
	})
	if err != nil {
		return err
	}
	return err2
}

func (t *Timerfd) Wait() (expirations uint64, err error) {
	var buf [8]byte

	n, err := t.fd.Read(buf[:])
	if err != nil {
		if pe, ok := err.(*os.PathError); ok {
			err = pe.Err
		}
		if err == unix.ECANCELED {
			err = ErrTimerfdCancelled
		}
		return 0, err
	}
	if n != 8 {
		panic(fmt.Sprintf("timerfd returned %d bytes (expected 8)", n))
	}
	return NativeEndian.Uint64(buf[:]), nil
}

func (t *Timerfd) Close() error {
	return t.fd.Close()
}

func NewRealtimeTimerfd() (*Timerfd, error) {
	return newTimerfd(unix.CLOCK_REALTIME)
}

func NewBoottimeTimerfd() (*Timerfd, error) {
	return newTimerfd(unix.CLOCK_BOOTTIME)
}
