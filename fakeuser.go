package main

type FakeUser struct {
	Name        string
	DefaultPath string
	CurrentPath string
}

func InitializeUser(name string) *FakeUser {
	us := new(FakeUser)
	us.DefaultPath = "/home/" + name

	us.Name = name
	return us
}
