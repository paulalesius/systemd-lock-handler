systemd-lock-handler
====================

`logind` (part of systemd) emits events when the system is locked or goes into
sleep.

These events however, are simple D-Bus events, and don't actually run anything.
There are no facilities to easily _run_ anything on these events either (e.g.:
a screen locker).

`systemd-lock-handler` application fills this gap.

When the system is either locked, or about to go into sleep, this service will
start the `lock.target` and `sleep.target` systemd user targets respectively.

You can then have any of your own services (including screen lockers and other
one-shot commands) run when this event is activated.

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

You should service files for anything you intend to intend to run on lock. For
example, `enabling` this service file would run `slock`:

    [Unit]
    Description=A simple X screen locker
    Requisite=xorg.target  # Optional, non-standard
    PartOf=lock.target  # Stop this unit if lock.target is stopped.

    [Service]
    ExecStart=/usr/bin/slock
    # Explicitly run `unlock.target` when locker exits. This stops
    # `lock.target`, and other units that are wanted by it.
    ExecStopPost=/usr/bin/systemctl --user start unlock.target

    [Install]
    WantedBy=lock.target
    WantedBy=sleep.target

You'll also need to enable and start this service:

    systemctl --user enable --now systemd-lock-handler.service

## Locking

Lock your session using `loginctl lock-session`.

This will mark your session as locked, and start `lock.target` along with any
services that are `WantedBy` it.

## Suspending

Sleep your device using `systemctl suspend`.

This will start `sleep.target` along with any services that are `WantedBy` it.
This will happen _before_ the system is suspended.

LICENCE
-------

systemd-lock-handler is licensed under the ISC licence. See LICENCE for details.
