systemd-lock-handler
====================

``logind`` (part of systemd) emits events when the system is lock or goes into
sleep.

These events though, are simple D-Bus events, and don't actually run anything.
There are no facilities to easily _run_ anything on these events either (e.g.:
a screen locker).

This application fills this gap.

When the system is either locked, or about to go into sleep, this service will
start the ``lock.target`` ``sleep.target`` systemd targets respectively.

You can then have any of your own services (including screen lockers and other
one-shot commands) run when this event is activated.

Note that systemd already has a ``sleep.target``, however, that's a
system-level target, and your user-level units can't rely on it. The one
included in this package does not conflict, but rather compliments that one.

Installation
------------

* Download the latest release_.
* ``tar xf v1.0.0``
* ``cd systemd-lock-handler-1.0.0/``
* ``sudo make install``

.. _release: https://github.com/WhyNotHugo/systemd-lock-handler/releases

Usage
-----

You should service files for anything you intend to intend to run on lock. For
example, ``enabling`` this service file would run ``slock``::

    [Unit]
    Description=A simple X screen locker
    Requisite=xorg.target

    [Service]
    ExecStart=/usr/bin/slock
    ExecStopPost=/usr/bin/systemctl --user start unlock.target

    [Install]
    WantedBy=lock.target

Keep in mind that, for this to work a few steps need to be taken:

Steps if you'll be using ``lock.target``
----------------------------------------

* Enable this service ``systemctl --user enable --now systemd-lock-handler.service``.
* Lock your session using ``loginctl lock-session``.

Steps if you'll be using ``sleep.target``
-----------------------------------------

* Enable this service ``systemctl --user enable --now systemd-lock-handler.service``.
* Sleep your device using ``systemctl suspend``.

LICENCE
-------

systemd-lock-handler is licensed under the ISC licence. See LICENCE for details.
