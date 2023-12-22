package main

import (
	"strings"
)

type File struct {
	Name     string
	Contents string
	Size     int

	Owner string
	Group string
}

type Directory struct {
	Name string
	Size int

	Owner string
	Group string

	Files       []File
	FileNames   map[string]int
	Directories map[string]*Directory
}

type FakeOS struct {
	Directories map[string]*Directory
	DefaultPath string

	CurrentUser *FakeUser
}

var commandsList []string

func InitializeOS() FakeOS {
	var os FakeOS

	os.DefaultPath = "/"

	osMainDir := new(Directory)
	osMainDir.Name = "/"

	bin := new(Directory)
	bin.Name = "bin"

	boot := new(Directory)
	boot.Name = "boot"

	dev := new(Directory)
	dev.Name = "dev"

	////
	home := new(Directory)
	home.Name = "home"

	home_admin := new(Directory)
	home_admin.Name = "admin"
	home_admin.Owner = "admin"
	home_admin.Group = "admin"
	home_admin.Files = append(home_admin.Files, File{Name: "database_creds", Contents: "user: sf - password: iloveyoualot41L"})
	home_admin.FileNames = map[string]int{}
	home_admin.FileNames["database_creds"] = 0

	home_admin.Directories = map[string]*Directory{}

	mineDir := new(Directory)
	mineDir.Name = "mine"
	home_admin.FileNames["mine"] = -1

	home_admin.Directories["/mine"] = mineDir
	bobDir := new(Directory)
	bobDir.Name = "bob"
	home_admin.Directories["/bob"] = bobDir
	home_admin.FileNames["bob"] = -1
	//
	home.Directories = map[string]*Directory{}
	home.Directories["/admin"] = home_admin
	///
	lib := new(Directory)
	lib.Name = "lib"

	lib32 := new(Directory)
	lib32.Name = "lib32"

	lib64 := new(Directory)
	lib64.Name = "lib64"

	mount := new(Directory)
	mount.Name = "mnt"

	media := new(Directory)
	media.Name = "media"

	proc := new(Directory)
	proc.Name = "proc"

	root := new(Directory)
	root.Name = "root"

	tmp := new(Directory)
	tmp.Name = "tmp"

	// var mainDirs []Directory
	// mainDirs = append(mainDirs, bin, boot, dev, home, lib, lib32, lib64, mount, media, proc, root, tmp)
	os.Directories = map[string]*Directory{}
	os.Directories["/bin"] = bin
	os.Directories["/boot"] = boot
	os.Directories["/dev"] = dev
	os.Directories["/home"] = home
	os.Directories["/lib"] = lib
	os.Directories["/lib32"] = lib32
	os.Directories["/mnt"] = mount
	os.Directories["/media"] = media
	os.Directories["/proc"] = proc
	os.Directories["/root"] = root
	os.Directories["/tmp"] = tmp

	for k, v := range os.Directories {
		dir := v
		dir.Owner = "root"
		dir.Group = "root"
		v = dir

		os.Directories[k] = v
	}

	return os
}

func (os *FakeOS) GetCurrentDirectory(path string) *Directory {
	iteratePath := strings.Split(path, "/")

	var newIteratePath []string

	for i := 0; i < len(iteratePath); i++ {
		if iteratePath[i] != "" {
			newIteratePath = append(newIteratePath, "/"+iteratePath[i])
		}
	}

	var currentDirectory *Directory
	currentDirectoryFound := false
	var ok bool = false
	for i := 0; i < len(newIteratePath); i++ {
		if !currentDirectoryFound {
			currentDirectory, ok = os.Directories[newIteratePath[i]]
			if !ok {
				panic(ok)
			}
			currentDirectoryFound = true
			continue
		}

		currentDirectory, ok = currentDirectory.Directories[newIteratePath[i]]

		if !ok {
			panic(ok)
		}
	}

	return currentDirectory
}

func DoesCommandExist(command string) bool {
	unknown := true
	for _, c := range commandsList {
		if c == command {
			unknown = false
			break
		}
	}

	return unknown
}
