/**
 * Deterministic two-thread workload for the java_cpu_alloc prof-correctness scenario.
 *
 * One thread burns CPU in a tight loop (feeds the cpu-time profile), the other
 * allocates byte[] continuously (feeds the allocation profile). Named Thread
 * subclasses (not lambdas/anonymous classes) keep frame names stable across
 * JDK versions for the expected_profile.json regexes.
 */
public class Workload {

    static volatile boolean running = true;
    static volatile long sink = 0;

    static class CpuBurner extends Thread {
        CpuBurner() {
            super("cpu-burner");
        }

        @Override
        public void run() {
            while (running) {
                sink += spin(100_000);
            }
        }

        private long spin(int iterations) {
            long acc = 0;
            for (int i = 0; i < iterations; i++) {
                acc += (long) i * i;
            }
            return acc;
        }
    }

    static class AllocGenerator extends Thread {
        // Batch-and-park keeps this thread's own CPU footprint low, so CPU
        // samples stay dominated by CpuBurner while still producing a steady
        // stream of allocations for the alloc-samples profile.
        private static final int BATCH_SIZE = 64;
        private static final long PARK_NANOS = 2_000_000L;

        AllocGenerator() {
            super("alloc-generator");
        }

        @Override
        public void run() {
            while (running) {
                for (int i = 0; i < BATCH_SIZE; i++) {
                    sink += allocate().length;
                }
                java.util.concurrent.locks.LockSupport.parkNanos(PARK_NANOS);
            }
        }

        private byte[] allocate() {
            return new byte[1024];
        }
    }

    public static void main(String[] args) throws InterruptedException {
        long durationSec = Long.parseLong(System.getenv().getOrDefault("EXECUTION_TIME_SEC", "10"));

        CpuBurner cpuBurner = new CpuBurner();
        AllocGenerator allocGenerator = new AllocGenerator();
        cpuBurner.start();
        allocGenerator.start();

        Thread.sleep(durationSec * 1000L);
        running = false;

        cpuBurner.join();
        allocGenerator.join();

        System.out.println("sink=" + sink);
    }
}
