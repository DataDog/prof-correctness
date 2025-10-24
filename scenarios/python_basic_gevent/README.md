# Description

gevent test

We don't expect to see anything else other than the event hub's internals in the
MainThread task, even though we run the same business logic as in the spawned
task. In real-life applications it is unlikely that business logic will ever run
directly on the event hub.
