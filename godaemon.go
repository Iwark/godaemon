package godaemon

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const (
	DaemonStart = 1 + iota
	DaemonSuccess
	DaemonFailure
)

func Start(child bool) error {
	if child {
		return childMain()
	}
	if err := parentMain(); err != nil {
		log.Fatalf("Error occurred [%v]", err)
		return err
	} else {
		return errors.New("Successfully finished parentMain process.")
	}
}

func OutputFile(logfile string) (*os.File, error) {
	logfilepath, err := filepath.Abs(logfile)
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(logfilepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func parentMain() (err error) {
	args := []string{"--child"}
	args = append(args, os.Args[1:]...)

	r, w, err := os.Pipe()
	if err != nil {
		return err
	}

	cmd := exec.Command(os.Args[0], args...)
	cmd.ExtraFiles = []*os.File{w}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err = cmd.Start(); err != nil {
		return err
	}

	var status int = DaemonStart
	go func() {
		buf := make([]byte, 1)
		r.Read(buf)

		if int(buf[0]) == DaemonSuccess {
			status = int(buf[0])
		} else {
			status = DaemonFailure
		}
	}()

	i := 0
	for i < 60 {
		if status != DaemonStart {
			fmt.Println("DAEMON:", status)
			break
		}
		time.Sleep(500 * time.Millisecond)
		i++
	}

	if status == DaemonSuccess {
		return nil
	} else {
		return fmt.Errorf("Child failed to start")
	}
}

func childMain() error {
	var err error

	pipe := os.NewFile(uintptr(3), "pipe")
	if pipe != nil {
		defer pipe.Close()
		if err == nil {
			pipe.Write([]byte{DaemonSuccess})
		} else {
			pipe.Write([]byte{DaemonFailure})
		}
	}

	signal.Ignore(syscall.SIGCHLD)

	syscall.Close(0)
	syscall.Close(1)
	syscall.Close(2)

	syscall.Setsid()
	syscall.Umask(022)

	// syscall.Chdir("/")

	return nil
}
