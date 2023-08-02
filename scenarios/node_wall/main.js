function busyWait(ms) {
  return new Promise(resolve => {
    let done = false;
    console.log('resolve')
    function work() {
      console.log('work')
      if (done) return;
      let sum = 0;
      for (let i = 0; i < 100e6; i++) {
        sum += i;
      }
      setImmediate(work);
    }
    setImmediate(work);
    setTimeout(() => {
      console.log('done = true')
      done = true;
      resolve();
    }, ms);
  });
}

function a(niter) {
  let sum = 0;
  for (let i = 0; i < niter; i++) {
    sum += i;
  }
  return sum
}

function b(niter) {
  let sum = 0;
  for (let i = 0; i < niter; i++) {
    sum += i;
  }
  return sum;
}

async function foo(nsecs) {
  let done = false;  
  function work() {
    if (done) {
      return;
    }
    const niter = 100e6
    a(niter);
    b(niter*2);
    setImmediate(work)
  }

  setTimeout(() => done = true, nsecs * 1000);
  setImmediate(work);
}

const executionTime = process.argv[2] || process.env.EXECUTION_TIME_SEC || 2
setTimeout(() => foo(executionTime), 100)
