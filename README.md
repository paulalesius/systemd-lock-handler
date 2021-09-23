systemd-lock-handler
====================

`logind` (part of systemd) emits events when the system is locked, unlocked or
goes into sleep.

These events however, are simple D-Bus events, and don't actually run anything.
There are no facilities for users to easily _run_ anything on these events
either (e.g.: a screen locker).

`systemd-lock-handler` is a small, lightweight helper fills this gap.

When the system is either locked, unlocked, or about to go into sleep, this
service will start the `lock.target`, `unlock.target` and `sleep.target`
systemd user targets respectively.

When the system is unlocked, `lock-target` will be stopped.

Any service can be configured to start with any of these targets:

- A screen locker.
- A service that keeps the screen off after 15 seconds of inactivity.
- A service that turns the volume to 0%.
- ...

Note that systemd already has a `sleep.target`, however, that's a system-level
target, and your user-level units can't rely on it. The one included in this
package does not conflict, but rather compliments that one.

Installation
------------

## On ArchLinux

A package is available in the AUR:

    paru -S systemd-lock-handler

## Other platforms

You can manually build and install:

    git@git.sr.ht:~whynothugo/systemd-lock-handler
    cd systemd-lock-handler
    make build
    sudo make install

Usage
-----

The service itself must be enabled for the current user:

    systemctl --user enable --now systemd-lock-handler.service

Additionally, service files must be created and enabled for any service that
should start when the system is locked.

For example, `enabling` this service file would run `slock`:

    [Unit]
    Description=A simple X screen locker
    Requisite=graphical-session.target
    PartOf=lock.target

    [Service]
    ExecStart=/usr/bin/slock
    # Unlock the session when the screen locker exit:
    ExecStopPost=/usr/bin/loginctl unlock-session

    [Install]
    WantedBy=lock.target
    WantedBy=sleep.target

Specifying `PartOf=lock.target` indicates to systemd that this service should
be stopped if `lock.target` is stopped. This is even more important for
services that _aren't_ the screen locker, since this setting means they'll get
stopped when the system is unlocked.

## Locking

Lock your session using `loginctl lock-session`.

This will mark the session as locked, and start `lock.target` along with any
services that are `WantedBy` it.

## Unlocking

Unlock your session using `loginctl unlock-session`.

This will mark the session as unlocked, start `unlock.target`, and stop
`lock.target`. 

Service that are marked `PartOf=lock.target` will be stopped when `lock.target`
stops.

## Suspending

Sleep your device using `systemctl suspend`.

This will start `sleep.target` along with any services that are `WantedBy` it.
This will happen _before_ the system is suspended.

Changelog
---------

## 2.2.0

- Also handle unlock events (and translate those to unlock.target).

## 2.1.0

- Minor bugfixes.
- Run as `Type=notify`.

## 2.0.0

- Rewrite in go.
- Move binary into /usr/lib.

## 1.1.0

- Use newer logind API.
- Events for other sessions are now correctly ignored.

## 1.0.0

Also handle sleep target.

## 0.1.0

Initial release.

LICENCE
-------

systemd-lock-handler is licensed under the ISC licence. See LICENCE for details.
