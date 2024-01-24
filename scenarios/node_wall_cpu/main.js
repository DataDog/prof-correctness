const {isMainThread, Worker, workerData} = require('worker_threads');
const {pbkdf2} = require('crypto')
const {promisify} = require('util');
const pbkdf2Async = promisify(pbkdf2)

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

async function c() {
  await pbkdf2Async('secret', 'salt', 200000, 64, 'sha512');
}

async function foo(nsecs) {
  await new Promise((resolve) => setTimeout(resolve, 100));

  let done = false;
  const niter = 100e6
  setTimeout(() => done = true, nsecs * 1000);

  let bCumulativeTime = 0;
  let cCumulativeTime = 0;

  while (!done) {
    a(niter);
    const t0 = Date.now();
    b(2*niter);
    const t1 = Date.now();
    bCumulativeTime += t1-t0;

    // Ensure that we spend as much time in c as in b
    while (cCumulativeTime < bCumulativeTime) {
      const t0 = Date.now();
      await c();
      const t1 = Date.now();
      cCumulativeTime += t1-t0;
    }
  }
}

let executionTime = 0

if (isMainThread) {
  executionTime = process.argv[2] || process.env.EXECUTION_TIME_SEC || 2
  const nworkers = process.argv[3] || process.env.NWORKERS || 1
  console.log(`executionTime: ${executionTime}, nworkers: ${nworkers}`)
  
  for (let i = 0; i < nworkers; i++) {
    new Worker(__filename, {workerData: {executionTime: executionTime/2}});
  }
} else {
  executionTime = workerData.executionTime;
}

foo(executionTime)
