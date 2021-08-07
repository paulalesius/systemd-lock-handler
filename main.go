package main

import (
	"context"
	"fmt"
	"log"
	"os/user"

	"github.com/coreos/go-systemd/v22/daemon"
	systemd "github.com/coreos/go-systemd/v22/dbus"
	"github.com/godbus/dbus/v5"
)

func StartSystemdUserUnit(unitName string) error {
	conn, err := systemd.NewUserConnectionContext(context.Background())
	if err != nil {
		return fmt.Errorf("failed to connect to systemd user session: %v", err)
	}

	_, err = conn.StartUnitContext(context.Background(), unitName, "replace", nil)
	if err != nil {
		return fmt.Errorf("failed to start unit: %v", err)
	}

	log.Println("Started systemd unit:", unitName)
	return nil
}

func ListenForSleep() {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		log.Fatalln("Could not connect to the system D-Bus", err)
	}

	err = conn.AddMatchSignal(
		dbus.WithMatchObjectPath("/org/freedesktop/login1"),
		dbus.WithMatchInterface("org.freedesktop.login1.Manager"),
		dbus.WithMatchMember("PrepareForSleep"),
	)
	if err != nil {
		log.Fatalln("Failed to listen for sleep signals", err)
	}

	c := make(chan *dbus.Signal, 10)

	go func() {
		for {
			<-c
			log.Println("The system is going to sleep")

			err = StartSystemdUserUnit("sleep.target")
			if err != nil {
				log.Println("Error starting sleep.target:", err)
			}
		}
	}()

	conn.Signal(c)
	log.Println("Listening for sleep events...")
}

func ListenForLock(user *user.User) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		log.Fatalln("Could not connect to the system D-Bus", err)
	}

	err = conn.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.login1.Session"),
		dbus.WithMatchSender("org.freedesktop.login1"),
		dbus.WithMatchMember("Lock"),
	)
	if err != nil {
		log.Fatalln("Failed to listen for lock signals", err)
	}

	c := make(chan *dbus.Signal, 10)
	go func() {
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
	}()

	conn.Signal(c)
	log.Println("Listening for lock events...")
}

func main() {
	log.SetFlags(log.Lshortfile)

	user, err := user.Current()
	if err != nil {
		log.Fatalln("Failed to get username:", err)
	}
	log.Println("Running for user:", user.Username)

	ListenForSleep()
	ListenForLock(user)

	log.Println("Initialization complete.")

	sent, err := daemon.SdNotify(true, daemon.SdNotifyReady)
	if !sent {
		log.Println("Couldn't call sd_notify. Not running via systemd?")
	}
	if err != nil {
		log.Println("Call to sd_notify failed:", err)
	}

	select {}
}
