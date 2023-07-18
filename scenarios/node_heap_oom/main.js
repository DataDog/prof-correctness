

async function foo(size) {
  const n = size / 8
  const x = []
  x.length = n
  for (let i = 0; i < n; i++) { x[i] = Math.random() }
}

const durationMs = (process.argv[2] || process.env.EXECUTION_TIME || 2) * 1000
setTimeout(() => foo(1024 * 1024 * 50), durationMs)
