package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"
)

/*
#include <stdio.h>
#include <stdlib.h>

void cAllocateMemory() {
    // Allocate some memory in the C function
	int size_alloc = 10000;
    int* data = (int*)malloc(size_alloc * sizeof(int));
	int sum = 0;
    // Use the allocated memory to avoid compiler optimizations
    for (int i = 0; i < size_alloc; i++) {
        data[i] = i;
    }
	for (int i = 0; i < size_alloc; i++) {
		sum += data[i];
	}
	printf("%d\n", sum);
    // Free the allocated memory
    free(data);
}
*/
import "C"

func burnCPU() {
	// Simulate CPU-intensive work
	for i := 0; i < 1000000000; i++ {
		_ = i * i
	}
}

func allocateMemory() {
	// Allocate some memory in Go
	data := make([]int, 100)

	// Use the allocated memory to avoid compiler optimizations
	for i := 0; i < 100; i++ {
		data[i] = i
	}
}

func main() {
	// Get the execution time from the environment variable
	executionTimeStr := os.Getenv("EXECUTION_TIME")
	executionTime, err := strconv.Atoi(executionTimeStr)
	if err != nil {
		fmt.Println("Error parsing EXECUTION_TIME:", err)
		return
	}

	// Calculate the end time based on the execution time
	endTime := time.Now().Add(time.Duration(executionTime) * time.Second)

	// Run the loop until the specified duration is reached
	for time.Now().Before(endTime) {
		burnCPU()
		allocateMemory()
		C.cAllocateMemory()
		time.Sleep(100 * time.Millisecond) // Sleep to simulate some delay
	}

	// Print some memory stats at the end
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Allocated memory (bytes): %v\n", m.Alloc)
	fmt.Printf("Total memory allocated (bytes): %v\n", m.TotalAlloc)
}
