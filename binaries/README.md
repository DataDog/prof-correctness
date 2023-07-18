# Binaries

This folder can be used to store profiler binaries that should be tested.
You can then add following line to have the binaries available in the docker build step.

```
COPY ./binaries/ /app/binaries/
```

The profiler install script should then look within this folder.
