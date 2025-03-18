import java.util.concurrent.Executors;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.TimeUnit;

public class CpuBurner {
    public static void main(String[] args) {
        int numThreads = Runtime.getRuntime().availableProcessors();
        ExecutorService executor = Executors.newFixedThreadPool(numThreads);

        int executionTimeSec = Integer.parseInt(System.getenv().getOrDefault("EXECUTION_TIME_SEC", "60"));

        System.out.println("Starting CPU intensive tasks on " + numThreads + " threads for " + executionTimeSec + " seconds...");

        for (int i = 0; i < numThreads; i++) {
            executor.submit(() -> {
                while (!Thread.currentThread().isInterrupted()) {
                    double value = Math.pow(Math.random(), Math.random());
                }
            });
        }

        try {
            Thread.sleep(executionTimeSec * 1000L);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }

        executor.shutdownNow();
        System.out.println("CPU intensive tasks stopped.");
    }
}
