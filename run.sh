#!/bin/zsh

latest_wheel=$(ls -t ddtrace*.whl | head -n 1)
echo "Using $latest_wheel"

REPO_ROOT=$(pwd)
echo "Using $REPO_ROOT"

cd scenarios/python_asyncio_3.11

cp $REPO_ROOT/$latest_wheel .

docker build -t python_asyncio_3.11 --build-arg LATEST_WHEEL=$latest_wheel . || exit 1
# docker run -it python_asyncio_3.11

cd ../../

NETWORK_HOST=YES TEST_SCENARIOS="python_asyncio" go test -v -run TestScenarios
