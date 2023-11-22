'use strict'

const DDTrace = require('dd-trace')
const tracer = DDTrace.init()

const OUTER_ROUNDS = 3
const INNER_ROUNDS = 3

let counter = 0

setTimeout(runBusySpans, 100)

function runBusySpans () {
  tracer.trace('x' + counter, { type: 'web', resource: `endpoint-${counter}` }, (span, done) => {
    // Make the span ID deterministic for the test.
    // NOTE: this (obviously) depends on unsupported internals of dd-trace-js as
    // it is accessing _* named properties. This is fine in tests, but can
    // obviously cause test failures if these internals change. Since this is
    // all internal to Datadog, we can adjust the tests if they break due to
    // internal changes in dd-trace-js.
    const rootSpanId = ((INNER_ROUNDS + 1) * counter) + 1
    span._spanContext._spanId._buffer = [0, 0, 0, 0, 0, 0, 0, rootSpanId]

    setImmediate(() => {
      for (let i = 0; i < INNER_ROUNDS; ++i) {
        const z = i
        tracer.trace('y' + i, (span2, done2) => {
          // Make the span ID deterministic for the test as above
          span2._spanContext._spanId._buffer = [0, 0, 0, 0, 0, 0, 0, rootSpanId + z + 1]
          setTimeout(() => {
            // Make sure we have a (counter, z) specific frame on stack
            switch (counter) {
              case 0:
                switch (i) {
                  case 0: busyLoop00(); break
                  case 1: busyLoop01(); break
                  case 2: busyLoop02(); break
                }
                break
              case 1:
                switch (i) {
                  case 0: busyLoop10(); break
                  case 1: busyLoop11(); break
                  case 2: busyLoop12(); break
                }
                break
              case 2:
                switch (i) {
                  case 0: busyLoop20(); break
                  case 1: busyLoop21(); break
                  case 2: busyLoop22(); break
                }
                break
            }
            done2()
            if (z === (INNER_ROUNDS - 1)) {
              if (++counter < OUTER_ROUNDS) {
                setTimeout(runBusySpans, 0)
              }
              done()
            }
          }, 0)
        })
      }
    })
  })
}

function busyLoop () {
  const start = process.hrtime.bigint()
  for (;;) {
    const now = process.hrtime.bigint()
    if (now - start > 505050505n) {
      break
    }
  }
}

function busyLoop00 () {
  busyLoop()
}

function busyLoop01 () {
  busyLoop()
}

function busyLoop02 () {
  busyLoop()
}

function busyLoop10 () {
  busyLoop()
}

function busyLoop11 () {
  busyLoop()
}

function busyLoop12 () {
  busyLoop()
}

function busyLoop20 () {
  busyLoop()
}

function busyLoop21 () {
  busyLoop()
}

function busyLoop22 () {
  busyLoop()
}
