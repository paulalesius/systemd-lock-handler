#!/usr/bin/env python

import logging

from twisted.internet import defer, reactor
from txdbus import client


logger = logging.getLogger(__name__)
logging.basicConfig(level=logging.INFO)


@defer.inlineCallbacks
def onLock():
    logger.info("Session is being locked.")
    try:
        cli = yield client.connect(reactor, 'session')

        robj = yield cli.getRemoteObject(
            'org.freedesktop.systemd1',
            '/org/freedesktop/systemd1',
        )
        yield robj.callRemote('StartUnit', 'lock.target', 'replace')
    except Exception:
        logger.exception('Error starting lock.target.')


@defer.inlineCallbacks
def onSleep(suspended):
    logger.info("System going into sleep.")
    try:
        cli = yield client.connect(reactor, 'session')

        robj = yield cli.getRemoteObject(
            'org.freedesktop.systemd1',
            '/org/freedesktop/systemd1',
        )
        yield robj.callRemote('StartUnit', 'sleep.target', 'replace')
    except Exception:
        logger.exception('Error starting lock.target.')


@defer.inlineCallbacks
def main():
    try:
        cli = yield client.connect(reactor, 'system')

        lock_obj = yield cli.getRemoteObject(
            'org.freedesktop.login1',
            '/org/freedesktop/login1/session/c1',
        )
        sleep_obj = yield cli.getRemoteObject(
            'org.freedesktop.login1',
            '/org/freedesktop/login1',
        )

        lock_obj.notifyOnSignal('Lock', onLock)
        sleep_obj.notifyOnSignal('PrepareForSleep', onSleep)

        logger.info("Ready and waiting for events.")
    except Exception:
        logger.exception('Error listening for lock events.')


reactor.callWhenRunning(main)
reactor.run()
