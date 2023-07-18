# Description

A test for graalVM native images with ddprof.

It looks like unwinding is not going well.

You can run this manually
```
TEST_SCENARIOS=".*graal.*" go test -v -run TestScenarios
```

# Sources

https://github.com/oracle/graal/issues/916
