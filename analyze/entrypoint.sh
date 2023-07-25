#!/bin/bash

cd /app
go test -v -run TestAnalyze -expectedJson /github/workspace/$INPUT_EXPECTED_JSON -pprofPath /github/workspace/$INPUT_PPROF_PATH
