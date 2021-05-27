package main

import (
	"fmt"
	"os"
	"os/user"

	"github.com/godbus/dbus/v5"
)

func StartSystemdUserUnit(unitName string) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	obj := conn.Object("org.freedesktop.systemd1", "/org/freedesktop/systemd1")
	call := obj.Call("org.freedesktop.systemd1.Manager.StartUnit", 0, unitName, "replace")

	if call.Err != nil {
		// FIXME: Don't panic!
		panic(call.Err)
	}

	fmt.Printf(">> Started %v.\n", unitName)
}

func ListenForSleep() {
	// Listen for going-to-sleep signal.

	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	err = conn.AddMatchSignal(
		dbus.WithMatchObjectPath("/org/freedesktop/login1"),
		dbus.WithMatchInterface("org.freedesktop.login1.Manager"),
		dbus.WithMatchMember("PrepareForSleep"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		panic(err)
	}

	c := make(chan *dbus.Signal, 10)
	conn.Signal(c)
	for range c {
		fmt.Printf("> System is going to sleep.\n")
		StartSystemdUserUnit("sleep.target")
	}

	// This never returns.
	panic("Unreachable")
}

func ListenForLock(user *user.User) {
	// Listen for lock signals.

	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	err = conn.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.login1.Session"),
		dbus.WithMatchSender("org.freedesktop.login1"),
		dbus.WithMatchMember("Lock"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		panic(err)
	}

	c := make(chan *dbus.Signal, 10)
	conn.Signal(c)
	for v := range c {
		// Check that the locked session is for the current user.

		obj := conn.Object("org.freedesktop.login1", v.Path)

		name, err := obj.GetProperty("org.freedesktop.login1.Session.Name")
		if err != nil {
			panic(err)
		}

		if name.Value() == user.Username {
			fmt.Printf("> Session locked for current user.\n")
			StartSystemdUserUnit("lock.target")
		}
	}

	// This never returns.
	panic("Unreachable")
}

func main() {
	user, err := user.Current()
	if err != nil {
		fmt.Printf("Failed to get username: %s\n", err)
		os.Exit(-1)
	}
	fmt.Printf("Running for user %v.\n", user.Username)

	go ListenForSleep()
	go ListenForLock(user)

	select {}
}
