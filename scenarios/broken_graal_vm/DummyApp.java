public class DummyApp {
  public static void main(String[] args) throws InterruptedException {
    // Consume half a CPU for several minutes
    long start = System.currentTimeMillis();
    long end = start + (args.length > 0 ? Integer.parseInt(args[0]) : 60) * 1000; // run for 60 seconds by default
    int n = args.length > 1 ? Integer.parseInt(args[1]) : 40; // calculate the 40th Fibonacci number by default
    while (System.currentTimeMillis() < end) {
      fibonacci(n); // calculate the nth Fibonacci number
    }
    System.out.println("Fibonacci application finished: " + fibonacci(n)); // print the final result
  }

  private static int fibonacci(int n) {
    if (n <= 1) return n;
    return fibonacci(n - 1) + fibonacci(n - 2);
  }
}
