package daemongo

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

func Start(child bool, logfile string) error {
	if child {
		return childMain(logfile)
	}
	if err := parentMain(); err != nil {
		log.Fatalf("Error occurred [%v]", err)
		return err
	} else {
		return errors.New("Successfully finished parentMain process.")
	}
}

func parentMain() (err error) {
	args := []string{"--child"}
	args = append(args, os.Args[1:]...)

	// 子プロセスとのパイプを作っておく
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

	// パイプから子プロセスの起動状態を取得する
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

	// 子プロセスの起動を30秒待つ
	i := 0
	for i < 60 {
		if status != DaemonStart {
			fmt.Println("DAEMON:", status)
			break
		}
		time.Sleep(500 * time.Millisecond)
		i++
	}

	// 親プロセス終了
	if status == DaemonSuccess {
		return nil
	} else {
		return fmt.Errorf("Child failed to start")
	}
}

func childMain(logfile string) error {
	var err error

	// 子プロセスの起動状態を親プロセスに通知する
	pipe := os.NewFile(uintptr(3), "pipe")
	if pipe != nil {
		defer pipe.Close()
		if err == nil {
			pipe.Write([]byte{DaemonSuccess})
		} else {
			pipe.Write([]byte{DaemonFailure})
		}
	}

	// logのファイルパスを絶対パスにする
	logfilepath, err := filepath.Abs(logfile)
	if err != nil {
		return err
	}

	// ログの出力先をファイルに。
	f, err := os.OpenFile(logfilepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	log.SetOutput(f)

	// SIGCHILDを無視する
	signal.Ignore(syscall.SIGCHLD)

	// STDOUT, STDIN, STDERRをクローズ
	syscall.Close(0)
	syscall.Close(1)
	syscall.Close(2)

	// プロセスグループリーダーになる
	syscall.Setsid()

	// Umaskをクリア
	syscall.Umask(022)

	// / にchdirする
	// syscall.Chdir("/")

	return nil
}
