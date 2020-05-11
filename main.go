package main

import (
	"log"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/unix"
)

func testRealtime() {
	timerfd, err := NewRealtimeTimerfd()
	if err != nil {
		log.Fatal(err)
	}
	defer timerfd.Close()

start:
	err = timerfd.Settime(&unix.ItimerSpec{
		Interval: unix.Timespec{Sec: 60},
		Value:    unix.Timespec{Sec: 60},
	}, nil, true, true)
	if err != nil {
		log.Fatal(err)
	}

	for {
		expirations, err := timerfd.Wait()
		if err != nil {
			if err == ErrTimerfdCancelled {
				log.Printf("CLOCK_REALTIME: cancelled due to discontinious change")
				goto start
			}
			log.Fatal(err)
		}
		log.Printf("CLOCK_REALTIME: expirations=%d", expirations)
	}
}

func testBoottime() {
	timerfd, err := NewBoottimeTimerfd()
	if err != nil {
		log.Fatal(err)
	}
	defer timerfd.Close()

	granularity := time.Minute

	for {
		now := time.Now()
		next := now.Add(granularity).Truncate(granularity)
		dt := next.Sub(now)

		err = timerfd.Settime(&unix.ItimerSpec{
			Value: unix.NsecToTimespec(dt.Nanoseconds()),
		}, nil, false, false)
		if err != nil {
			log.Fatal(err)
		}

		expirations, err := timerfd.Wait()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("CLOCK_BOOTTIME: expirations=%d", expirations)
	}
}

func main() {
	var g errgroup.Group

	g.Go(func() error {
		testRealtime()
		return nil
	})
	g.Go(func() error {
		testBoottime()
		return nil
	})

	g.Wait()
}
