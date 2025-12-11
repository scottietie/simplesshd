package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	"github.com/pkg/sftp"
)

func main() {
	port := flag.Int("port", 22222, "Port to listen on")
	password := flag.String("password", "secret", "Password for authentication")
	flag.Parse()

	// 1. SFTP Handler (Files)
	sftpHandler := func(s ssh.Session) {
		user := s.User()
		log.Printf("[SFTP] Session started for user: %s", user)

		server, err := sftp.NewServer(
			s,
			sftp.WithMaxTxPacket(32768),
		)
		if err != nil {
			log.Printf("[SFTP] Init error: %v\n", err)
			return
		}

		if err := server.Serve(); err == io.EOF {
			server.Close()
			log.Printf("[SFTP] Session ended for user: %s", user)
		} else if err != nil {
			log.Printf("[SFTP] Server error: %v\n", err)
		}
	}

	// 2. SSH Handler (Shell)
	sshHandler := func(s ssh.Session) {
		user := s.User()
		log.Printf("[SHELL] Session started for user: %s", user)

		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}

		var cmd *exec.Cmd
		if len(s.Command()) > 0 {
			cmd = exec.Command(shell, "-c", s.RawCommand())
		} else {
			cmd = exec.Command(shell)
		}

		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, fmt.Sprintf("SSH_USER=%s", user))

		ptyReq, winCh, isPty := s.Pty()
		if isPty {
			cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
			f, err := pty.Start(cmd)
			if err != nil {
				io.WriteString(s, "Could not start pty.\n")
				return
			}
			defer f.Close()

			go func() {
				for win := range winCh {
					setWinsize(f, win.Width, win.Height)
				}
			}()

			go func() { io.Copy(f, s) }()
			io.Copy(s, f)

			cmd.Wait()
			log.Printf("[SHELL] Session ended for user: %s", user)
		} else {
			io.WriteString(s, "No PTY requested.\n")
			s.Exit(1)
		}
	}

	// 3. Server Configuration
	server := &ssh.Server{
		Addr: fmt.Sprintf(":%d", *port),
		SubsystemHandlers: map[string]ssh.SubsystemHandler{
			"sftp": sftpHandler,
		},
		Handler: sshHandler,
	}

	// 4. Auth Handler
	server.PasswordHandler = func(ctx ssh.Context, pw string) bool {
		if pw == *password {
			log.Printf("[AUTH] User '%s' authenticated.", ctx.User())
			return true
		}
		log.Printf("[AUTH] User '%s' denied.", ctx.User())
		return false
	}

	// Show the password used for authentication at server start
	log.Printf("Password is %s", *password)

	log.Printf("Starting SSH/SFTP server on port %d...", *port)
	log.Fatal(server.ListenAndServe())
}

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}
