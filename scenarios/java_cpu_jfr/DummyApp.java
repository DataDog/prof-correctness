/**
 * Simple CPU-intensive app for JFR correctness testing.
 *
 * Continuously computes Fibonacci numbers to keep the CPU busy.
 * The JFR recording is started by the JVM flags passed on the command line;
 * this app just runs for the requested duration and exits.
 *
 * The EXECUTION_TIME_SEC environment variable (default: 30) controls how long
 * the app runs, matching the convention used by prof-correctness.
 */
public class DummyApp {
    public static void main(String[] args) throws InterruptedException {
        int seconds = 30;
        String env = System.getenv("EXECUTION_TIME_SEC");
        if (env != null && !env.isEmpty()) {
            seconds = Integer.parseInt(env.trim());
        }
        System.out.println("Running DummyApp for " + seconds + " seconds");

        long endMs = System.currentTimeMillis() + seconds * 1000L;
        int n = 40; // fib(40) takes ~0.4 s per call, keeps CPU fully busy
        while (System.currentTimeMillis() < endMs) {
            fibonacci(n);
        }
        System.out.println("DummyApp finished: " + fibonacci(n));
    }

    /** Intentionally naive recursive Fibonacci — the hot function we assert on. */
    static long fibonacci(int n) {
        if (n <= 1) return n;
        return fibonacci(n - 1) + fibonacci(n - 2);
    }
}
