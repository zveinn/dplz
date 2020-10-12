package main

import (
	"io/ioutil"
	"log"
	"time"

	"golang.org/x/crypto/ssh"
)

func NewSSHConfig(user, key string, timeout int, ignoreInsecure bool) (cfg *ssh.ClientConfig) {
	cfg = new(ssh.ClientConfig)
	cfg.User = user
	cfg.Auth = []ssh.AuthMethod{
		PrivateKey(key),
	}
	if ignoreInsecure {
		cfg.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}
	cfg.Timeout = time.Duration(time.Duration(timeout) * time.Second)
	return
}
func PrivateKey(path string) ssh.AuthMethod {
	key, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		panic(err)
	}
	return ssh.PublicKeys(signer)
}

func NewSessionForCommand(cmd *CMD, conn *ssh.Client) {
	cmd.StdOut.Buffer = make(chan []byte, 1000000)
	cmd.StdErr.Buffer = make(chan []byte, 1000000)
	session, err := conn.NewSession()
	if err != nil {
		log.Println("Session error:", err)
		return
	}
	cmd.Session = session
	session.Stdout = &cmd.StdOut
	session.Stderr = &cmd.StdErr
	newSTDin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		log.Println("STDOUT:", err)
		return
	}

	cmd.StdIn = newSTDin

}
