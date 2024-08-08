package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime/debug"
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

func readFromBuffers(c *CMD) {
	for {
		select {
		case msg, ok := <-c.StdOut.Buffer:
			if !ok {
				continue
			}

			fmt.Print("\n" + color.CyanString("OUT ("+c.Hostname+"): "+c.Run) + "\n" + string(msg) + "\n")
		case msg, ok := <-c.StdErr.Buffer:
			if !ok {
				continue
			}
			fmt.Print(color.RedString("ERR ("+c.Hostname+"): "+c.Run) + "\n" + string(msg) + "\n")
			// color.Red("ERR (" + c.Hostname + "): " + c.Run)
			// fmt.Println(string(msg))
		default:
			// color.Green("FIN (" + c.Hostname + "): " + c.Run)
			fmt.Print(color.GreenString("FIN ("+c.Hostname+"): "+c.Run) + "\n")
			return
		}
	}
}

func (c *CMD) Execute() {
	defer func() {
		r := recover()
		if r != nil {
			log.Println(r, string(debug.Stack()))
		}
		ParseWaitGroup.Done()
	}()

	c.ID = uuid.New()
	err := c.Session.Run(c.Run)
	if err != nil {
	}
	readFromBuffers(c)
}

func (c *CMD) ExecuteLocal() {
	defer func() {
		r := recover()
		if r != nil {
			log.Println(r, string(debug.Stack()))
		}
		ParseWaitGroup.Done()
	}()

	cmd := exec.Command("sh", "-c", c.Run)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error executing command:", err)
		return
	}

	fmt.Println(string(output))
}

func (c *CMD) CopyTemplate(server *Server) {
	defer func() {
		r := recover()
		if r != nil {
			log.Println(r, string(debug.Stack()))
		}
		ParseWaitGroup.Done()
	}()

	c.ID = uuid.New()
	c.Run = "TEMPLATE > " + c.Template.Local
	err := os.WriteFile("/tmp/"+c.ID.String(), c.Template.Data, 0o644)
	if err != nil {
		fmt.Fprintf(&c.StdErr, err.Error()+"\n")
		return
	}

	c.MoveFile("/tmp/"+c.ID.String(), c.Template.Remote, c.Template.Mode)

	err = os.Remove("/tmp/" + c.ID.String())
	if err != nil {
		fmt.Fprintf(&c.StdErr, err.Error()+"\n")
		return
	}

	fmt.Fprintf(&c.StdOut, c.Template.Local+" > "+c.Template.Remote+"\n")

	readFromBuffers(c)
}

func (c *CMD) CopyDirectory(server *Server) {
	defer func() {
		r := recover()
		if r != nil {
			log.Println(r, string(debug.Stack()))
		}
		ParseWaitGroup.Done()
	}()

	c.ID = uuid.New()
	c.Run = " DIRECTORY > " + c.Directory.Src
	if _, err := os.Stat(c.Directory.Src); os.IsNotExist(err) {
		c.WriteCopyDirectoryError(err.Error())
		return
	}

	// out, err := c.Session.CombinedOutput("stat " + c.Directory.Dst)
	// c.Session.Close()
	// c.NewSessionForCommand(c.Conn)
	// c.SetBuffers()
	// if err != nil {
	// 	fmt.Fprint(&c.StdErr, string(out), CloseTag)
	// 	return
	// }

	var preLevel int = 0
	var level int = 0
	var pathSplit []string
	if err := c.Session.Start("/bin/scp -rt " + c.Directory.Dst); err != nil {
		c.WriteCopyDirectoryError(err.Error())
		return
	}
	c.ExpectedSuccessCount++
	_ = godirwalk.Walk(c.Directory.Src, &godirwalk.Options{
		Callback: func(osPathname string, info *godirwalk.Dirent) error {
			// skip the root directory.
			if osPathname == c.Directory.Src {
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
	defer func() {
		r := recover()
		if r != nil {
			log.Println(r, string(debug.Stack()))
		}
		ParseWaitGroup.Done()
	}()

	c.ID = uuid.New()
	c.Run = "FILE > " + c.File.Local
	c.MoveFile(c.File.Local, c.File.Remote, c.File.Mode)
	readFromBuffers(c)
}

func (c *CMD) MoveFile(src, dst, mode string) {
	pathSplit := strings.Split(dst, "/")
	pathSplitLenth := len(pathSplit)
	var fileName string
	var dstPath string
	if pathSplitLenth == 1 {
		fileName = pathSplit[0]
		dstPath = "./"
	} else {
		fileName = pathSplit[len(pathSplit)-1]
		dstPath = strings.Join(pathSplit[0:len(pathSplit)-1], "/")
	}
	// log.Println(dstPath+"/", fileName)
	file, err := os.OpenFile(src, os.O_RDWR, 0o644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	s, err := file.Stat()
	if err != nil {
		panic(err)
	}
	if err := c.Session.Start("/bin/scp -rt " + dstPath + "/"); err != nil {
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
	fmt.Fprintf(&c.StdErr, c.Directory.Src+" > "+c.Directory.Dst+":"+msg+"\n"+CloseTag+"\n")
}

func OpenSessionsAndRunCommands(server *Server) {
	defer func() {
		if r := recover(); r != nil {
			logger.GenericError(logger.TypeCastRecoverInterface(r)).Log()
		}
	}()

	conn, err := ssh.Dial("tcp", server.IP+":"+server.Port, NewSSHConfig(server.User, server.Key, server.Password, 10, true))
	if err != nil {
		newErr := logger.GenericErrorWithMessage(err, "@ "+server.IP+">"+server.Hostname)
		newErr.Log()
		return
	}
	server.Client = conn

	// for i := range server.Pre {
	// 	server.Pre[i].Hostname = server.Hostname
	// 	server.Pre[i].NewSessionForCommand(server.Client)
	// 	ParseWaitGroup.Add(1)
	// 	server.Pre[i].Execute()
	// 	CommandOutputHandler(&server.Pre[i], &ParseWaitGroup)
	// }

	for i, s := range server.Scripts {
		if ScriptFilter != "" {
			if s.Filter != ScriptFilter && ScriptFilter != "*" {
				fmt.Println("SKIPPING: ", s.Filter, " .. ", ScriptFilter)
				continue
			}
		}

		for ii, iv := range s.CMD {
			if CMDFilter != "" {
				if CMDFilter != iv.Filter && CMDFilter != "*" {
					continue
				}
			}

			doCMD := func(c *CMD) {
				// log.Println("RUNNING:", c.File, c.Template, c.Filter)
				c.Hostname = server.Hostname
				err := c.NewSessionForCommand(server.Client)
				if err != nil {
					fmt.Println(err)
					return
				}
				if c.Template != nil {
					c.CopyTemplate(server)
				} else if c.File != nil {
					c.CopyFile(server)
				} else if c.Directory != nil {
					c.CopyDirectory(server)
				} else if c.Local {
					c.ExecuteLocal()
				} else {
					c.Execute()
				}
			}

			ParseWaitGroup.Add(1)
			if iv.Async {
				go doCMD(&server.Scripts[i].CMD[ii])
			} else {
				doCMD(&server.Scripts[i].CMD[ii])
			}

			// if iv.Async {
			// 	go CommandOutputHandler(&server.Scripts[i].CMD[ii], &ParseWaitGroup)
			// } else {
			// CommandOutputHandler(&server.Scripts[i].CMD[ii], &ParseWaitGroup)
			// }
		}
	}
	// for i := range server.Post {
	// 	server.Post[i].Hostname = server.Hostname
	// 	server.Post[i].NewSessionForCommand(server.Client)
	// 	ParseWaitGroup.Add(1)
	// 	server.Post[i].Execute()
	// 	CommandOutputHandler(&server.Post[i], &ParseWaitGroup)
	// }
}

func CommandOutputHandler(cmd *CMD, wait *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			logger.GenericError(logger.TypeCastRecoverInterface(r)).Log()
		}
		cmd.Session.Close()
		// close(cmd.StdOut.Buffer)
		// close(cmd.StdErr.Buffer)
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
		// for ii, iv := range v.Pre {
		// 	Servers[i].Pre[ii].Run = ReplaceVariables(iv.Run, v, nil)
		// }
		for ii, iv := range v.Scripts {
			for iii, iiv := range iv.CMD {
				if iiv.Template != nil {
					Servers[i].Scripts[ii].CMD[iii].Template.Data = []byte(ReplaceVariables(string(iiv.Template.Data), v, &iv))
					Servers[i].Scripts[ii].CMD[iii].Template.Local = ReplaceVariables(iiv.Template.Local, v, &iv)
					Servers[i].Scripts[ii].CMD[iii].Template.Remote = ReplaceVariables(iiv.Template.Remote, v, &iv)
				}
				if iiv.Run != "" {
					Servers[i].Scripts[ii].CMD[iii].Run = ReplaceVariables(iiv.Run, v, &iv)
				}
				if iiv.File != nil {
					Servers[i].Scripts[ii].CMD[iii].File.Local = ReplaceVariables(iiv.File.Local, v, &iv)
					Servers[i].Scripts[ii].CMD[iii].File.Remote = ReplaceVariables(iiv.File.Remote, v, &iv)
				}
			}
		}
		// for ii, iv := range v.Post {
		// 	Servers[i].Post[ii].Run = ReplaceVariables(iv.Run, v, nil)
		// }
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

	// for i, v := range Deployment.Variables {
	// 	out = strings.Replace(out, "{[deployment.variables."+i+"]}", v, -1)
	// }

	out = strings.Replace(out, "{[deployment.varFile]}", Deployment.Vars, -1)
	out = strings.Replace(out, "{[deployment.project]}", Deployment.Script, -1)
	out = strings.Replace(out, "{[deployment.servers]}", Deployment.Server, -1)

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
