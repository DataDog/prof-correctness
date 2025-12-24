# Python Flask correctness check

This test validates that Flask apps are properly profiled.

## Test application

The Flask app runs a web server with two endpoints:

- `/`: Returns "Hello, World!" after performing CPU-intensive computation (0.5s of busy work)
- `/stop`: Terminates the process via SIGINT

A background thread makes HTTP requests to the main endpoint for 5 seconds, then calls the stop endpoint to gracefully
terminate the application. This ensures the profiler captures Flask request handling and CPU activity.
