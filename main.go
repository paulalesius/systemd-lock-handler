package main

import (
	"fmt"
	"log"
	"os/user"

	"github.com/godbus/dbus/v5"
)

func StartSystemdUserUnit(unitName string) error {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	obj := conn.Object("org.freedesktop.systemd1", "/org/freedesktop/systemd1")

	call := obj.Call("org.freedesktop.systemd1.Manager.StartUnit", 0, unitName, "replace")
	if call.Err != nil {
		return fmt.Errorf("failed to start unit: %v", call.Err)
	}

	log.Println("Started unit: ", unitName)
	return nil
}

func ListenForSleep() {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		log.Fatalln("Could not connect to the system D-Bus", err)
	}
	defer conn.Close()

	err = conn.AddMatchSignal(
		dbus.WithMatchObjectPath("/org/freedesktop/login1"),
		dbus.WithMatchInterface("org.freedesktop.login1.Manager"),
		dbus.WithMatchMember("PrepareForSleep"),
	)
	if err != nil {
		log.Fatalln("Failed to listen for sleep signals", err)
	}

	c := make(chan *dbus.Signal, 10)
	conn.Signal(c)
	for {
		<-c
		log.Println("The system is going to sleep")

		err = StartSystemdUserUnit("sleep.target")
		if err != nil {
			log.Println("Error starting sleep.target:", err)
		}
	}
}

func ListenForLock(user *user.User) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		log.Fatalln("Could not connect to the system D-Bus", err)
	}
	defer conn.Close()

	err = conn.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.login1.Session"),
		dbus.WithMatchSender("org.freedesktop.login1"),
		dbus.WithMatchMember("Lock"),
	)
	if err != nil {
		log.Fatalln("Failed to listen for lock signals", err)
	}

	c := make(chan *dbus.Signal, 10)
	conn.Signal(c)
	for {
		v := <-c

		// Get the locked session...
		obj := conn.Object("org.freedesktop.login1", v.Path)

		name, err := obj.GetProperty("org.freedesktop.login1.Session.Name")
		if err != nil {
			log.Println("WARNING: Could not obtain details for locked session:", err)
			continue
		}

		// ... And check that it belongs to the current user:
		if name.Value() != user.Username {
			continue
		}
		log.Println("Session locked for current user.")

		err = StartSystemdUserUnit("lock.target")
		if err != nil {
			log.Println("Error starting lock.target:", err)
		}
	}
}

func main() {
	log.SetFlags(log.Lshortfile)

	user, err := user.Current()
	if err != nil {
		log.Fatalln("Failed to get username:", err)
	}
	log.Println("Running for user:", user.Username)

	go ListenForSleep()
	go ListenForLock(user)

	select {}
}
