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

        string[] queries = File.ReadAllLines(filePath);
        Queue<string> queryQueue = new Queue<string>(queries);

        Thread thread1 = new Thread(() => ProcessQueries(queryQueue, 4)) { Name = "LongRunningThread" };
        Thread thread2 = new Thread(() => ProcessQueries(queryQueue, 3)) { Name = "ShortRunningThread" };

        thread1.Start();
        thread2.Start();

        thread1.Join();
        thread2.Join();
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
