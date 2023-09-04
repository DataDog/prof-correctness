const refs = []

async function foo(size) {
  const n = size / 8
  const x = []
  x.length = n
  for (let i = 0; i < n; i++) { x[i] = Math.random() }
  refs.push(x)
}

const allocSize = 1024 * 1024 * 5
const allocPeriodMs = 200
const interval = setInterval(() => foo(allocSize), allocPeriodMs)
