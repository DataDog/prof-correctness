# prof-correctness

A small tool to make it simple to write profiler correctness tests which run known test programs and make assertions on the resulting profiles.
Checkout #profiling-library-pager to get notified of test failures.

## Setup

### Pre-requisites

Install go >= 1.25.1: `brew install go`, `choco install go`, etc.
Install docker.

### Running Tests

```sh
go test -v -run TestScenarios # Run all scenarios

TEST_SCENARIOS="ddprof.*"  go test -v -run TestScenarios # Run ddprof scenarios
```

### Using as a GitHub Action

You may use this repo to run the analyzer on your profiler emitted pprof files.
You can do so by adding a GitHub Action Workflow to your repo, where you build
your profiler and run it on an example program. After you did this, you can add
a step to analyze your results and match it against your expectation:

```yaml
      - name: Check profiler correctness for allocations
        uses: Datadog/prof-correctness/analyze
        with:
          expected_json: profiling/tests/correctness/allocations.json
          pprof_path: profiling/tests/correctness/allocations/
```

You need to provide a JSON file with your expectations and a path to where to
find the pprof files.

## Creating new tests 

### Define the dockerfile 

Create a test scenario and prefix the folders with something relevant like `php`, `ddprof`, `go`...
The dockerfile specifies how to install the profiler and run the test app.
The dockerfile needs to follow rules
- set variable `EXECUTION_TIME_SEC` (which defines how long the tests runs for)
- output pprof data to the `/app/data/` folder
  The /app/data mirrors the data folder in this repository.
```
# Define OS / Install App...
# Install profiler...
# Run things
ENV EXECUTION_TIME_SEC="60"
# Default is that test data is dropped in the data folder
ENV DD_PROFILING_PPROF_PREFIX="/app/data/profiles_"
CMD ["ddprof", "-l", "notice", "/app/build/some_app" ]
```

The `./profilers` folder contains helpers.

### Describe the expected output

Describe what you expect in a json file within the same folder. *This data is captured at every run in the matching folder*, so you can use it as a reference to adjust your test. Here is an example of a json file

```
{
  "test_name":"some_app",
  "stacks": [
    {
      "profile-type": "cpu-time",
      "stack-content":
      [
        {
          "regular_expression":";_start;__libc_start_main.*;main;a$",
          "value": 33,
          "error_margin": 5
        },
        {
          "regular_expression":";_start;__libc_start_main.*;main;b$",
          "value": 66,
          "error_margin": 5
        }
      ]
    }
  ]
}
```

### Profile input formats (pprof and OTLP)

The analyzer's internal representation is **google/pprof**. It accepts two
on-disk input formats and converts OTLP into pprof so that assertions in
`expected_profile.json` work identically regardless of the input:

- **google/pprof** (`.pprof`, optionally lz4/zstd compressed) — the historical
  format emitted by ddprof and the standalone `dd-otel-host-profiler`. Used
  as-is.
- **OpenTelemetry profiles / OTLP** — the wire format of the eBPF full-host
  profiler and the datadog-agent host-profiler. Drop the
  `ExportProfilesServiceRequest` bytes as `.otlp` (protobuf) or `.otlp.json`
  (OTLP JSON). This is **converted to pprof** on read (see `analysis/otlp.go`):
  it is not a native OTLP assertion path. One OTLP request may carry several
  profiles (e.g. `alloc_space`, `alloc_objects`, `inuse_space`,
  `inuse_objects`, cpu `samples`); they are merged into one pprof profile whose
  sample-type columns are addressed by `profile-type` in the JSON.

Format is chosen by filename suffix (`.otlp`/`.pb` → OTLP proto,
`.otlp.json` → OTLP JSON, otherwise pprof); pprof files that fail to parse are
retried as OTLP, so `pprof-regex` can point at either.

> Note: because pprof is the intermediate model, anything the OTLP payload
> carries that does not map onto pprof (resource/sample attribute tables,
> span/trace links, per-sample timestamps) must be translated in the converter
> or it is dropped. Attribute→label mapping is not yet implemented, so
> label-based assertions currently only work for pprof inputs.

### Run your test

```
TEST_RUN_SECS=12 TEST_SCENARIOS="MyTestsName.*" go test -v -run TestScenarios
```

### Cleanup test data (optional)

The tests results are written to the data folder. You can periodically clean that folder.

## Troubleshooting 

### Export profiles through the agent

If you have an agent setup locally, you can run the command line with `NETWORK_HOST=YES`. Using network host will allow the docker instance to target the agent running locally. Example:

```
NETWORK_HOST=YES TEST_SCENARIOS="ddprof_allo.*"  TEST_RUN_SECS=13 go test -v -run TestScenarios
```

## Slack notifications

You'll find a slack notification call in the
[`test.yml`](.github/workflows/test.yml) that triggers when a prof correctness
run fails. This expects a `SLACK_WEBHOOK` secret being set in the repository
which must be a URL to a [Slack workflow
webhook](https://slack.com/help/articles/360041352714-Create-more-advanced-workflows-using-webhooks).
It passes the `scenario` and `failed_run_url` used for testing to the webhook,
please make sure that the webhook is configured to accept this.

