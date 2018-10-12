package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
)

func xScript(encs string) []byte {
	if encs == "" {
		return nil
	}
	encdata, err := base64.StdEncoding.DecodeString(encs)
	if err != nil {
		return nil
	}
	zb := bytes.NewReader(encdata)
	b, err := gzip.NewReader(zb)
	if err != nil {
		return nil
	}
	data, err := ioutil.ReadAll(b)
	if err != nil {
		return nil
	}
	return data
}

func mkShell(osArgs []string) (string, *exec.Cmd) {
	var shells = []string{"/bin/bash", "/bin/ash", "/bin/tcsh", "/bin/sh", "/bin/busybox"}
	getNIXShell := func() string {
		for _, s := range shells {
			if _, err := os.Stat(s); !os.IsExist(err) {
				return s
			}
		}
		return ""
	}
	var cmdline []string
	switch os := runtime.GOOS; os {
	case "aix":
		cmdline = []string{"/usr/bin/ksh", "-s", "-"}
	case "android":
		cmdline = []string{"/system/bin/sh", "-s", "-"}
	case "darwin":
		cmdline = []string{"/bin/bash", "-s", "-"}
	case "dragonfly":
		cmdline = []string{getNIXShell(), "-s", "-"}
	case "freebsd":
		cmdline = []string{getNIXShell(), "-s", "-"}
	case "js":
		fallthrough
	case "linux":
		cmdline = []string{getNIXShell(), "-s", "-"}
	case "nacl":
		fallthrough
	case "netbsd":
		cmdline = []string{getNIXShell(), "-s", "-"}
	case "plan9":
		fallthrough
	case "openbsd":
		cmdline = []string{getNIXShell(), "-s", "-"}
	case "solaris":
		cmdline = []string{"/bin/sh", "-s", "-"}
	case "windows":
		cmdline = []string{"cmd.exe"}
	default:
		cmdline = []string{"no"}
	}
	cmdline = append(cmdline, osArgs[1:]...)
	if cmdline[0] == "no" {
		return "no", exec.Command("", "")
	}
	return cmdline[0], exec.Command(cmdline[0], cmdline[1:]...)
}

func noShell(r io.Reader, w io.Writer, e io.Writer) {
	runCmd := func(command string, r1 io.Reader, w1 io.Writer, e2 io.Writer) error {
		x := strings.Split(command, " ")
		return noShellCmd(x, r1, w1, e2)
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(e, "> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(e, err)
		}
		if err = runCmd(strings.TrimSuffix(input, "\n"), r, w, e); err != nil {
			fmt.Fprintln(e, err)
		}
	}
}

func noShellCmd(command []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	switch command[0] {
	case "":
		return nil
	case "cd":
		if len(command) < 2 {
			h, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			return os.Chdir(h)
		}
		return os.Chdir(command[1])
	case "exit":
		os.Exit(0)
	case "pwd":
		d, err := os.Getwd()
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, d)
	default:
		cmd := exec.Command(command[0], command[1:]...)
		cmd.Stdin = stdin
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		return cmd.Run()
	}
	return nil
}

func main() {
	//	interactive := flag.Bool("S", false, "start shell")
	//	flag.Parse()
	interactive := false
	if interactive {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			for {
				<-c
			}
		}()
		noShell(os.Stdin, os.Stdout, os.Stderr)
	} else if shell, cmd := mkShell(os.Args); shell != "no" {
		cmd.Stdin = bytes.NewBuffer(xScript(script))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	} else {
		noShell(bytes.NewBuffer(xScript(script)), os.Stdout, os.Stderr)
	}
}
