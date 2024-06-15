package main

import (
	"io"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

var Servers []*Server
var Variables = make(map[string]string)
var Deployment *D
var ParseWaitGroup = sync.WaitGroup{}
var CloseTag = "lkajbdflkajbdslkfbalkdfdeploy"
var CMDFilter string
var ScriptFilter string

type D struct {
	Servers   string
	Project   string
	Vars      string
	Variables map[string]string
}

// Server ..
type Server struct {
	Key       string
	Password  string
	User      string
	Hostname  string
	IP        string
	Port      string
	Scripts   []Script // list of service tags
	Pre       []CMD    `json:"pre"`
	Post      []CMD    `json:"post"`
	Variables map[string]string

	Client *ssh.Client
}
type ChannelWriter struct {
	Buffer chan []byte
}

// Service ..
// This represents a base service you wish to use.
type Script struct {
	Name      string
	CMD       []CMD `json:"cmd"`
	Variables map[string]string
	Filter    string
	// Templates map[string][]byte
}

type CMD struct {
	ID   uuid.UUID
	Done bool
	Run  string

	Out      string
	Success  bool
	Hostname string
	Filter   string
	Async    bool

	// Experimental
	Conn *ssh.Client

	// File and Directory specific
	File                 *File
	Template             *Template
	Directory            *Direcotry
	ExpectedSuccessCount int
	TotalSuccessCount    int

	// Session
	StdIn   io.WriteCloser
	StdOut  ChannelWriter
	StdErr  ChannelWriter
	Session *ssh.Session
}
type File struct {
	Local  string
	Remote string
	Mode   string
}
type Template struct {
	Local  string
	Remote string
	Data   []byte
	Mode   string
}
type Direcotry struct {
	Local  string
	Remote string
	Mode   string
}
