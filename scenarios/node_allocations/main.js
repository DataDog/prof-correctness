function a(size, refs) {
  // prevent string interning
  refs.push((' ' + 'a'.repeat(size)).slice(1))
}

function b(size, refs) {
  // prevent string interning
  refs.push((' ' + 'b'.repeat(size)).slice(1))
}

async function foo(iterCount, allocPeriodMs) {
  let count = 0;
  function work() {
    count += 1
    a(allocSize, refs);
    b(allocSize * 2, refs);

    if (count == iterCount) {
      refs.forEach((x, i) => { if ((Math.floor(i / 2) % 2) == 0) { refs[i] = undefined; } })
      if (global.gc) { global.gc(); }
      return;
    }

    setTimeout(work, allocPeriodMs)
  }

  setTimeout(work, 0);
}

var refs = []
const allocSize = 1024 * 1024 * 2
const allocPeriodMs = 100
const durationMs = (process.argv[2] || process.env.EXECUTION_TIME_SEC || 2) * 1000
const iterCount = durationMs / allocPeriodMs
setTimeout(() => foo(iterCount, allocPeriodMs), 100)
