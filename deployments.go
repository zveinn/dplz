package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/karrick/godirwalk"
	"github.com/opensourcez/logger"
	"golang.org/x/crypto/ssh"
)

func (c *CMD) Execute() {
	c.ID = uuid.New()
	id := ReplaceInUUID(c.ID.String())
	c.SetBuffersAndOpenShell()
	// keeping this one just in case ...
	// log.Println("executing..")
	// err := c.Session.Run(c.Run)
	// if err != nil {
	// 	log.Println(err, err.Error())
	// 	fmt.Fprint(c.StdIn, err.Error(), CloseTag)
	// }

	fmt.Fprint(c.StdIn, c.Run+" 2> /tmp/"+id+".err 1> /tmp/"+id+".out;if [ -s /tmp/"+id+".err ];then cat /tmp/"+id+".err <(echo "+CloseTag+") >&2;else cat /tmp/"+id+".out <(echo "+CloseTag+");fi;rm /tmp/"+id+".*;\n")
	// fmt.Fprint(c.StdIn, c.Run+" &> /tmp/"+id+".out; cat /tmp/"+id+".out <(echo "+CloseTag+");rm /tmp/"+id+".out;\n")
}

func (c *CMD) CopyTemplate(server *Server) {
	c.ID = uuid.New()
	c.Run = " TEMPLATE > " + c.Template.Local
	c.SetBuffers()
	err := os.WriteFile("/tmp/"+c.ID.String(), c.Template.Data, 0o644)
	if err != nil {
		fmt.Fprintf(&c.StdErr, err.Error()+" "+CloseTag+"\n")
		return
	}

	c.MoveFile("/tmp/"+c.ID.String(), c.Template.Remote, c.Template.Mode)

	err = os.Remove("/tmp/" + c.ID.String())
	if err != nil {
		fmt.Fprintf(&c.StdErr, err.Error()+" "+CloseTag+"\n")
		return
	}

	fmt.Fprintf(&c.StdOut, c.Template.Local+" > "+c.Template.Remote+"\n"+CloseTag+"\n")
}

func (c *CMD) CopyDirectory(server *Server) {
	c.ID = uuid.New()
	c.Run = " DIRECTORY > " + c.Directory.Local
	if _, err := os.Stat(c.Directory.Local); os.IsNotExist(err) {
		c.WriteCopyDirectoryError(err.Error())
		return
	}

	out, err := c.Session.CombinedOutput("stat " + c.Directory.Remote)
	c.Session.Close()
	c.NewSessionForCommand(c.Conn)
	c.SetBuffers()
	if err != nil {
		fmt.Fprint(&c.StdErr, string(out), CloseTag)
		return
	}

	var preLevel int = 0
	var level int = 0
	var pathSplit []string
	if err := c.Session.Start("/bin/scp -rt " + c.Directory.Remote); err != nil {
		c.WriteCopyDirectoryError(err.Error())
		return
	}
	c.ExpectedSuccessCount++
	_ = godirwalk.Walk(c.Directory.Local, &godirwalk.Options{
		Callback: func(osPathname string, info *godirwalk.Dirent) error {
			// skip the root directory.
			if osPathname == c.Directory.Local {
				return nil
			}
			pathSplit = strings.Split(osPathname, "/")
			level = len(pathSplit)
			isDir := info.IsDir()
			if !isDir {
				level--
			}
			levelJump := preLevel - level
			// log.Println(osPathname, info.IsDir(), info, info.Name(), level, preLevel, levelJump)
			for i := 0; i < levelJump; i++ {
				// log.Println("Leaving dir:", osPathname)
				fmt.Fprintln(c.StdIn, "E")
				c.ExpectedSuccessCount++
			}
			if info.IsDir() {
				if level != preLevel {
					if level > preLevel {
						// log.Println("Creating dir:", info.Name())
						fmt.Fprintln(c.StdIn, "D"+c.Directory.Mode, 0, info.Name())
						c.ExpectedSuccessCount++
					}
				}
			} else {

				file, err := os.OpenFile(osPathname, os.O_RDWR, 0o644)
				if err != nil {
					c.WriteCopyDirectoryError(err.Error())
					return err
				}
				defer file.Close()
				s, err := file.Stat()
				if err != nil {
					c.WriteCopyDirectoryError(err.Error())
					return err
				}

				// log.Println("Creating file:", osPathname)
				fmt.Fprintln(c.StdIn, "C"+c.Directory.Mode, s.Size(), s.Name())
				c.ExpectedSuccessCount++
				n, err := io.Copy(c.StdIn, file)
				if err != nil {
					c.WriteCopyDirectoryError(err.Error())
					return err
				}
				if n != s.Size() {
					c.WriteCopyDirectoryError("Failed sending file, only sent " + strconv.Itoa(int(n)) + " bytes of " + strconv.Itoa(int(s.Size())))
					return errors.New("...")
				}
				fmt.Fprint(c.StdIn, "\x00")
				c.ExpectedSuccessCount++
			}
			preLevel = level
			return nil
		},
		Unsorted: true,
	})
}

func (c *CMD) CopyFile(server *Server) {
	c.ID = uuid.New()
	c.Run = " FILE > " + c.File.Local
	c.SetBuffers()
	c.MoveFile(c.File.Local, c.File.Remote, c.File.Mode)
}

func (c *CMD) MoveFile(local, remote, mode string) {
	pathSplit := strings.Split(remote, "/")
	pathSplitLenth := len(pathSplit)
	var fileName string
	var remotePath string
	if pathSplitLenth == 1 {
		fileName = pathSplit[0]
		remotePath = "./"
	} else {
		fileName = pathSplit[len(pathSplit)-1]
		remotePath = strings.Join(pathSplit[0:len(pathSplit)-1], "/")
	}
	// log.Println(remotePath+"/", fileName)
	file, err := os.OpenFile(local, os.O_RDWR, 0o644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	s, err := file.Stat()
	if err != nil {
		panic(err)
	}
	if err := c.Session.Start("/bin/scp -rt " + remotePath + "/"); err != nil {
		panic("Failed to run: " + err.Error())
	}
	c.ExpectedSuccessCount++
	fmt.Fprintln(c.StdIn, "C"+mode, s.Size(), fileName)
	c.ExpectedSuccessCount++
	n, err := io.Copy(c.StdIn, file)
	if err != nil {
		n, err := io.Copy(c.StdIn, file)
		if err != nil {
			panic(fmt.Sprintf("Did not copy full file, copied %d bytes but file is %d bytes. Error: %s", n, s.Size(), err.Error()))
		}
	}
	if n != s.Size() {
		panic(fmt.Sprintf("Did not copy full file, copied %d bytes but file is %d bytes", n, s.Size()))
	}
	fmt.Fprint(c.StdIn, "\x00")
	c.ExpectedSuccessCount++
}

func (c *CMD) WriteCopyDirectoryError(msg string) {
	fmt.Fprintf(&c.StdErr, c.Directory.Local+" > "+c.Directory.Remote+":"+msg+"\n"+CloseTag+"\n")
}

func OpenSessionsAndRunCommands(server *Server) {
	defer func() {
		if r := recover(); r != nil {
			logger.GenericError(logger.TypeCastRecoverInterface(r)).Log()
		}
	}()

	conn, err := ssh.Dial("tcp", server.IP+":"+server.Port, NewSSHConfig(server.User, server.Key, 10, true))
	if err != nil {
		newErr := logger.GenericErrorWithMessage(err, "@ "+server.IP+">"+server.Hostname)
		newErr.Log()
		return
	}
	server.Client = conn

	for i := range server.Pre {
		server.Pre[i].Hostname = server.Hostname
		server.Pre[i].NewSessionForCommand(server.Client)
		ParseWaitGroup.Add(1)
		server.Pre[i].Execute()
		CommandOutputHandler(&server.Pre[i], &ParseWaitGroup)
	}

	for i, s := range server.Scripts {
		if ScriptFilter != "" {
			if s.Filter != ScriptFilter && ScriptFilter != "*" {
				// fmt.Println("SKIPPING: ", s.Filter, " .. ", ScriptFilter)
				continue
			}
		}

		for ii, iv := range s.CMD {
			// log.Println("RUNNING:", ii, iv, CMDFilter)
			if CMDFilter != "" {
				if CMDFilter != iv.Filter && CMDFilter != "*" {
					continue
				}
			}

			server.Scripts[i].CMD[ii].Hostname = server.Hostname
			server.Scripts[i].CMD[ii].NewSessionForCommand(server.Client)
			ParseWaitGroup.Add(1)
			if iv.Template != nil {
				if iv.Async {
					go server.Scripts[i].CMD[ii].CopyTemplate(server)
				} else {
					server.Scripts[i].CMD[ii].CopyTemplate(server)
				}
			} else if iv.File != nil {
				if iv.Async {
					go server.Scripts[i].CMD[ii].CopyFile(server)
				} else {
					server.Scripts[i].CMD[ii].CopyFile(server)
				}
			} else if iv.Directory != nil {
				if iv.Async {
					go server.Scripts[i].CMD[ii].CopyDirectory(server)
				} else {
					server.Scripts[i].CMD[ii].CopyDirectory(server)
				}
			} else {
				if iv.Async {
					go server.Scripts[i].CMD[ii].Execute()
				} else {
					server.Scripts[i].CMD[ii].Execute()
				}
			}
			if iv.Async {
				go CommandOutputHandler(&server.Scripts[i].CMD[ii], &ParseWaitGroup)
			} else {
				CommandOutputHandler(&server.Scripts[i].CMD[ii], &ParseWaitGroup)
			}
		}
	}
	for i := range server.Post {
		server.Post[i].Hostname = server.Hostname
		server.Post[i].NewSessionForCommand(server.Client)
		ParseWaitGroup.Add(1)
		server.Post[i].Execute()
		CommandOutputHandler(&server.Post[i], &ParseWaitGroup)
	}
}

func CommandOutputHandler(cmd *CMD, wait *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			logger.GenericError(logger.TypeCastRecoverInterface(r)).Log()
		}
		cmd.Session.Close()
		close(cmd.StdOut.Buffer)
		close(cmd.StdErr.Buffer)
		wait.Done()
		// color.Magenta("Closing: " + cmd.ID.String())
	}()

	var closing bool
	for {
		time.Sleep(100 * time.Millisecond)
		// log.Println("LEN:", cmd.ID, len(cmd.StdErr.Buffer), len(cmd.StdOut.Buffer))
		select {
		case msg, ok := <-cmd.StdOut.Buffer:
			if !ok {
				continue
			}
			if bytes.Contains(msg, []byte(CloseTag)) {
				closing = true
				msg = bytes.Replace(msg, []byte(CloseTag+"\n"), []byte(""), -1)
			}
			if len(msg) > 0 {
				// log.Println(string(msg), msg)
				if cmd.File != nil || cmd.Directory != nil && msg[0] == 0 {
					// 0 means scp success
					cmd.TotalSuccessCount = cmd.TotalSuccessCount + len(msg)
					// log.Println(cmd.TotalSuccessCount, cmd.ExpectedSuccessCount)
					if cmd.TotalSuccessCount >= cmd.ExpectedSuccessCount {
						color.Green("(" + cmd.Hostname + "): " + cmd.Run)
						fmt.Println(string(msg))
						closing = true
					}
				} else if cmd.File != nil && msg[0] == 1 {
					// 1 means scp error
					color.Red("(" + cmd.Hostname + "): " + cmd.Run)
					fmt.Println(string(msg))
					closing = true
				} else {
					color.Green("(" + cmd.Hostname + "): " + cmd.Run)
					fmt.Println(string(msg))
				}
			}
			if closing {
				cmd.Done = true
				cmd.Success = true
				return
			}

		case msg, ok := <-cmd.StdErr.Buffer:
			if !ok {
				continue
			}
			if bytes.Contains(msg, []byte(CloseTag)) {
				closing = true
				msg = bytes.Replace(msg, []byte(CloseTag+"\n"), []byte(""), -1)
			}
			if len(msg) > 0 {
				color.Red("(" + cmd.Hostname + "): " + cmd.Run)
				fmt.Println(string(msg))
			}
			if closing {
				cmd.Done = true
				return
			}

		case <-time.After(360 * time.Second):
			log.Println("TIMEOUT:", cmd.Hostname, cmd.Run)
			return
		default:
			// log.Println(cmd.ID)
		}
	}
}

func InjectVariables() {
	for i, v := range Servers {
		for ii, iv := range v.Pre {
			Servers[i].Pre[ii].Run = ReplaceVariables(iv.Run, v, nil)
		}
		for ii, iv := range v.Scripts {
			for iii, iiv := range iv.CMD {
				if iiv.Template != nil {
					Servers[i].Scripts[ii].CMD[iii].Template.Data = []byte(ReplaceVariables(string(iiv.Template.Data), v, &iv))
				} else {
					Servers[i].Scripts[ii].CMD[iii].Run = ReplaceVariables(iiv.Run, v, &iv)
				}
			}
		}
		for ii, iv := range v.Post {
			Servers[i].Post[ii].Run = ReplaceVariables(iv.Run, v, nil)
		}
		// for ii, iv := range v.Templates {
		// 	Servers[i].Templates[ii] = []byte(ReplaceVariables(string(iv), v, nil))
		// 	// log.Println("DATA:", i, string(Servers[i].Templates[ii]))
		// }
	}
}

func ReplaceVariables(in string, server *Server, script *Script) (out string) {
	out = in
	for i, v := range Variables {
		out = strings.Replace(out, "{["+i+"]}", v, -1)
	}
	out = strings.Replace(out, "{[server.ip]}", server.IP, -1)
	out = strings.Replace(out, "{[server.key]}", server.Key, -1)
	out = strings.Replace(out, "{[server.user]}", server.User, -1)
	out = strings.Replace(out, "{[server.port]}", server.Port, -1)
	out = strings.Replace(out, "{[server.hostname]}", server.Hostname, -1)

	for i, v := range server.Variables {
		out = strings.Replace(out, "{[server.variables."+i+"]}", v, -1)
	}

	for i, v := range Deployment.Variables {
		out = strings.Replace(out, "{[deployment.variables."+i+"]}", v, -1)
	}

	out = strings.Replace(out, "{[deployment.varFile]}", Deployment.Vars, -1)
	out = strings.Replace(out, "{[deployment.project]}", Deployment.Project, -1)
	out = strings.Replace(out, "{[deployment.servers]}", Deployment.Servers, -1)

	if script != nil {
		for i, v := range script.Variables {
			out = strings.Replace(out, "{[script.variables."+i+"]}", v, -1)
		}
		out = strings.Replace(out, "{[script.tag]}", script.Name, -1)
	}
	return
}

func (c *ChannelWriter) Write(buf []byte) (n int, err error) {
	// log.Println("WRITING:", len(buf), string(buf))
	b := make([]byte, len(buf))
	copy(b, buf)
	select {
	case c.Buffer <- b:
	default:
		err = errors.New("COULD NOT WRITE ON CHANNEL")
	}
	return len(b), err
}
