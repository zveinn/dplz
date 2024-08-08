package main

import (
	"log"
	"os"
	"runtime/debug"
	"time"

	"github.com/opensourcez/logger"
	"golang.org/x/crypto/ssh"
)

func NewSSHConfig(user, key string, password string, timeout int, ignoreInsecure bool) (cfg *ssh.ClientConfig) {
	cfg = new(ssh.ClientConfig)
	cfg.User = user

	if password != "" {
		cfg.Auth = []ssh.AuthMethod{
			ssh.Password(password),
		}
	}

	if key != "" {
		cfg.Auth = []ssh.AuthMethod{
			PrivateKey(key),
		}
	}

	if ignoreInsecure {
		cfg.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}
	cfg.Timeout = time.Duration(time.Duration(timeout) * time.Second)
	return
}

func PrivateKey(path string) ssh.AuthMethod {
	key, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		panic(err)
	}
	return ssh.PublicKeys(signer)
}

func (c *CMD) SetBuffersAndOpenShell() {
	// THE SHELL NEEDS TO BE LAST!
	err := c.Session.Shell()
	if err != nil {
		log.Println(err, string(debug.Stack()))
	}
}

func (c *CMD) NewSessionForCommand(conn *ssh.Client) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			logger.GenericError(logger.TypeCastRecoverInterface(r)).Log()
		}
	}()
	session, err := conn.NewSession()
	if err != nil {
		return err
	}
	c.Session = session
	c.Conn = conn
	c.StdOut.Buffer = make(chan []byte, 10000000)
	c.StdErr.Buffer = make(chan []byte, 10000000)
	c.Session.Stdout = &c.StdOut
	c.Session.Stderr = &c.StdErr
	newSTDin, err := c.Session.StdinPipe()
	if err != nil {
		c.Session.Close()
		return err
	}
	c.StdIn = newSTDin
	return nil
}
