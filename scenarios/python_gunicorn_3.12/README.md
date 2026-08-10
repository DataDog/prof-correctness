# Description

Correctness check for `gunicorn`.

Because we want to run the test for a few seconds, then to exit automatically, the script is both the app definition and
a worker thread that issues HTTP requests to the app. After a few seconds, a special request is sent which makes the app
exit, terminating the process (at which point we look at the resulting Profiles from the testing framework).

In the expected Profile, we look both at how much CPU and time the app consumes, and how much CPU and time the Requester
thread uses.
