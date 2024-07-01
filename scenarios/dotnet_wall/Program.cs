using System;
using System.Collections.Generic;
using System.IO;
using System.Threading;

class Program
{
    static void Main(string[] args)
    {
        string filePath = "queries.txt";

        if (!File.Exists(filePath))
        {
            Console.WriteLine("Query file not found.");
            return;
        }

        while (true)
        {
            string[] queries = File.ReadAllLines(filePath);
            Queue<string> queryQueue = new Queue<string>(queries);

            List<Thread> threads = new List<Thread>();
            for (int i = 1; i <= 20; i++)
            {
                int duration = i % 2 == 0 ? 4 : 3; // Alternating durations for demonstration
                Thread thread = new Thread(() => ProcessQueries(queryQueue, duration)) { Name = $"Thread{i}" };
                threads.Add(thread);
            }

            foreach (var thread in threads)
            {
                thread.Start();
            }

            foreach (var thread in threads)
            {
                thread.Join();
            }
        }
    }

    static void ProcessQueries(Queue<string> queryQueue, int duration)
    {
        while (true)
        {
            string query;
            lock (queryQueue)
            {
                if (queryQueue.Count == 0) break;
                query = queryQueue.Dequeue();
            }

            if (query.StartsWith("CPU"))
            {
                Console.WriteLine($"Processing CPU intensive query on thread {Thread.CurrentThread.ManagedThreadId} for {duration} seconds");
                DateTime end = DateTime.Now.AddSeconds(duration);
                while (DateTime.Now < end)
                {
                    // Simulate CPU work
                    for (int i = 0; i < 1000; i++) ;
                }
            }
            else if (query.StartsWith("Sleep"))
            {
                Console.WriteLine($"Processing Sleep query on thread {Thread.CurrentThread.ManagedThreadId} for {duration} seconds");
                Thread.Sleep(duration * 1000);
            }
        }
    }
}
