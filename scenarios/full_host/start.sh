#!/bin/bash

echo "Starting dd-otel-host-profiler daemon..."

# Start the profiler daemon in the background
# Configure it to output pprof files locally with 10s reporting period
DD_HOST_PROFILING_UPLOAD_SYMBOLS=false \
dd-otel-host-profiler \
    --pprof-prefix=/app/data/profiles_ \
    --upload-period=5s \
    --split-by-service=true \
    --verbose &

PROFILER_PID=$!

# Give the profiler a moment to start up
sleep 2

echo "Starting test application for ${EXECUTION_TIME_SEC} seconds..."

# Show what processes are running before starting our app
echo "Processes before starting clang_pie:"
ps aux | grep -E "(clang_pie|PID)" || true

# Run the test application with DD_SERVICE set
DD_SERVICE=clang_pie_test timeout ${EXECUTION_TIME_SEC}s /app/clang_pie &
APP_PID=$!

echo "Started clang_pie with PID: $APP_PID"
echo "Processes after starting clang_pie:"
ps aux | grep -E "(clang_pie|PID)" || true

# Wait for the application to complete
wait $APP_PID || true

echo "Test application finished, stopping profiler..."

# Stop the profiler daemon
kill $PROFILER_PID 2>/dev/null || true
wait $PROFILER_PID 2>/dev/null || true

# Give it a moment to flush data
sleep 2

echo "Profiling complete. Output files:"
ls -la /app/data/
