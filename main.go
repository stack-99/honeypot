package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/spf13/viper"
	"github.com/stack-99/honeypot/writers"
	"github.com/stack-99/honeypot/writers/colors"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

const appName = "go0r"

func authenticatePassword(ctx ssh.Context, password string) bool {
	//logParameters(conn)

	fmt.Println(fmt.Sprintf("Request with password: %s ", password))

	if ctx.User() == "admin" && string(password) == "admin" {
		fmt.Println("He logged in :)")

		return true
	}

	return false
}

func init() {
	usr, err := user.Current()
	if err != nil {
		fmt.Println(err)
	}
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(err)
	}

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/" + appName)
	viper.AddConfigPath(usr.HomeDir + "/" + appName)
	viper.AddConfigPath(dir + "/configs/")

	viper.BindEnv("port", "GOOR_PORT")
	viper.SetDefault("port", "1400")

	viper.BindEnv("host_key", "GOOR_HOST_KEY")
	viper.SetDefault("host_key", "./host_key")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Can not load config file. defaults are loaded!")
	}
}

func GetIPPort(remoteAddress net.Addr) (string, uint16) {
	switch addr := remoteAddress.(type) {
	case *net.UDPAddr:
		return addr.IP.String(), uint16(addr.Port)
	case *net.TCPAddr:
		return addr.IP.String(), uint16(addr.Port)
	}

	return "", 0
}

// FakeShell create a fake shell to waste attacker's time
// Read command, and "execute" them
func FakeShell(s ssh.Session) {
	var cmdsFilePath string

	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// if this env exists, honeypot is running on docker, and path is "/app/config/cmds.txt"
	if os.Getenv("HONEYPOT_CMDFILE") == "" {
		cmdsFilePath = path + "/config/cmds.txt"
	} else {
		cmdsFilePath = os.Getenv("HONEYPOT_CMDFILE")
	}

	bytes, err := os.ReadFile(cmdsFilePath)
	if err != nil {
		panic(err)
	}

	commandsList = strings.Split(string(bytes), "\n")

	fakeOS := InitializeOS()
	fakeOS.CurrentUser = InitializeUser(s.User())
	fakeOS.CurrentUser.CurrentPath = fakeOS.CurrentUser.DefaultPath

	term := term.NewTerminal(s, fmt.Sprintf(
		"%s%s%s@%s%s%s>$%s ",
		colors.Yellow,
		s.User(),
		colors.Green,
		colors.Blue,
		s.LocalAddr(),
		colors.Green,
		colors.Reset,
	))

	ip, _ := GetIPPort(s.RemoteAddr())
	writers.ColorWriteFast(term, `Welcome to Ubuntu 20.04.2 LTS (GNU/Linux 5.4.0-73-generic x86_64)

	* Documentation:  https://help.ubuntu.com
	* Management:     https://landscape.canonical.com
	* Support:        https://ubuntu.com/advantage
	`, colors.White)

	writers.ColorWriteFast(term, fmt.Sprintf(`
Welcome!
   
This server is hosted by Contabo. If you have any questions or need help,
please don't hesitate to contact us at support@contabo.com.
   
Last login: Thu Apr 14 16:59:26 2022 from %s`, ip), colors.White)

	writers.PrintEnd(term, 1)

	for {

		ln, err := term.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				panic(err)
			}
		}
		if ln == "exit" {
			break
		}

		commandAndArgs := strings.Split(ln, " ")
		command := commandAndArgs[0]
		args := commandAndArgs[1:]

		if DoesCommandExist(command) {
			writers.ColorWriteFast(term, fmt.Sprintf("%s: command not found", command), colors.Red)
			writers.PrintEnd(term, 1)
			continue
		}

		if command == "whoami" {
			writers.ColorWriteFast(term, fakeOS.CurrentUser.Name, colors.White)
			writers.PrintEnd(term, 1)
		} else if command == "pwd" {
			writers.ColorWriteFast(term, fakeOS.CurrentUser.CurrentPath, colors.White)
			writers.PrintEnd(term, 1)
		} else if command == "cat" {
			var currentDirectory *Directory = fakeOS.GetCurrentDirectory(fakeOS.CurrentUser.CurrentPath)

			idx, ok := currentDirectory.FileNames[args[0]]
			if !ok {
				writers.ColorWriteFast(term, `cat: `+args[0]+`: No such file or directory`, colors.White)
				writers.PrintEnd(term, 1)
				continue
			}

			if idx == -1 {
				writers.ColorWriteFast(term, `cat: `+args[0]+`: Is a directory`, colors.White)
				writers.PrintEnd(term, 1)
				continue
			}

			file := currentDirectory.Files[idx]

			writers.ColorWriteFast(term, file.Contents, colors.White)
			writers.PrintEnd(term, 1)
		} else if command == "ls" {
			var iteratePath string = ""
			if len(args) >= 1 {
				// 				if len(args) == 3 && args[2] == "/" {
				// 					writers.ColorWriteFast(term, rootDir, colors.White)
				// 					writers.PrintEnd(term, 1)
				// 				} else {
				// 					writers.ColorWriteFast(term, `drwxr-xr-x 9 admin admin    4096 Apr 12 19:43 .
				// drwxr-xr-x 7 root root    4096 Dec 27 13:24 ..`, colors.White)
				// 					writers.PrintEnd(term, 1)
				// 				}
			} else {
				iteratePath = fakeOS.CurrentUser.CurrentPath
			}

			var currentDirectory *Directory = fakeOS.GetCurrentDirectory(iteratePath)

			list := ""
			for _, v := range currentDirectory.Directories {
				list = list + v.Name + " "
			}

			for i := 0; i < len(currentDirectory.Files); i++ {
				list = list + currentDirectory.Files[i].Name + " "
			}

			writers.ColorWriteFast(term, list, colors.White)
			writers.PrintEnd(term, 1)

			// } else if command == "cd" {
			// 	if len(commandAndArgs) == 1 || len(commandAndArgs) == 2 {
			// 		currentPath = homePath
			// 	} else if len(commandAndArgs) == 4 && commandAndArgs[3] == "/" {
			// 		logger.Info("Switched to root dir")
			// 		currentPath = commandAndArgs[2]
			// 	} else if len(commandAndArgs) == 4 {
			// 		logger.Infof("Switched to %s", commandAndArgs[2])
			// 		currentPath = commandAndArgs[2]
			// 	}
		} else if command == "touch" {
			if len(args) == 0 {
				writers.ColorWriteFast(term, `touch: missing file operand
Try 'touch --help' for more information.`, colors.White)
				writers.PrintEnd(term, 1)

				continue
			}

			var currentDirectory *Directory = fakeOS.GetCurrentDirectory(fakeOS.CurrentUser.CurrentPath)

			if _, ok := currentDirectory.FileNames[args[0]]; ok {
				continue
			}

			currentDirectory.FileNames[args[0]] = len(currentDirectory.Files)
			currentDirectory.Files = append(currentDirectory.Files, File{Name: args[0]})
		} else if command == "mkdir" {
			if len(args) == 0 {
				writers.ColorWriteFast(term, `mkdir: missing operand
Try 'mkdir --help' for more information.`, colors.White)
				writers.PrintEnd(term, 1)

				continue
			}

			var currentDirectory *Directory = fakeOS.GetCurrentDirectory(fakeOS.CurrentUser.CurrentPath)

			if _, ok := currentDirectory.FileNames[args[0]]; ok {
				writers.ColorWriteFast(term, `mkdir: cannot create directory ‘`+args[0]+`’: File exists`, colors.White)
				writers.PrintEnd(term, 1)
				continue
			}

			new_dir := new(Directory)
			new_dir.Name = "admin"
			new_dir.Owner = fakeOS.CurrentUser.Name
			new_dir.Group = fakeOS.CurrentUser.Name
			new_dir.FileNames = map[string]int{}

			currentDirectory.FileNames[args[0]] = -1
			currentDirectory.Directories["/"+args[0]] = new_dir
		} else if command == "date" {
			currentTime := time.Now()
			layout := "Mon 02 Jan 2006 03:04:05 PM"

			writers.ColorWriteFast(term, currentTime.Format(layout)+" CEST", colors.White)
			writers.PrintEnd(term, 1)
		}
	}

	_, err = term.Write([]byte(colors.Reset))
	if err != nil {
		if err == io.EOF {
			s.Close()
			return
		} else {
			panic(err)
		}
	}
	s.Close()
}

// sessionHandler is called after authentication
func sessionHandler(s ssh.Session) {
	FakeShell(s)
}

// ReadHostKeyFile read the given hostkeyfile and return a gossh.Signer which contains the key
func ReadHostKeyFile(filepath string) (gossh.Signer, error) {

	keyBytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	key, err := gossh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func main() {
	keyPath := viper.GetString("host_key")
	key, err := ReadHostKeyFile(keyPath)

	if err != nil {
		panic(err)
	}

	s := &ssh.Server{
		Addr:            fmt.Sprintf("0.0.0.0:%s", viper.GetString("port")),
		Handler:         sessionHandler,
		PasswordHandler: authenticatePassword,
		IdleTimeout:     5 * time.Minute,
	}

	s.AddHostKey(key)

	panic(s.ListenAndServe().Error())
}
