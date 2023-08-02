#include <iostream>
#include <thread>
#include <vector>
#include <chrono>
#include <atomic>

std::atomic<bool> running{true};

void allocate_memory(size_t size) {
    char* data = new char[size];
    data[0] = 'a';
    data[size - 1] = 'z';
    delete[] data;
}

char* leak_function(int i) {
    if (i % 3 == 0) {
        return static_cast<char*>(malloc(1280));
    }
    return nullptr;
}

void thread_function() {
    int i = 0;
    while (running) {
        ++i;
        allocate_memory(1024);
        allocate_memory(2048);
        allocate_memory(4096);
    
        char* ptr = leak_function(i);
        
        if (ptr) {
            ptr[0] = 'c';
        }
        std::this_thread::sleep_for(std::chrono::milliseconds(100));
    }
}

int main(int argc, char** argv) {
    int test_duration = 10;
    const char* exec_time_env = getenv("EXECUTION_TIME_SEC");
    if (exec_time_env) {
        test_duration = atoi(exec_time_env);
        if (test_duration == 0) {
            exit(1);
        }
    }
    printf("Executable %s starting for %d seconds\n", argv[0], test_duration);

#ifdef THREADED_TEST
    const int num_threads = 10;
    std::vector<std::thread> threads;
    for (int i = 0; i < num_threads; ++i) {
        threads.emplace_back(thread_function);
    }
#endif
    time_t end = time(NULL) + test_duration;
    int i = 0;
    while (time(NULL) < end) {
        ++i;
        allocate_memory(512);
        allocate_memory(1024);
        allocate_memory(2048);
        allocate_memory(4096);
        allocate_memory(8192);

        std::this_thread::sleep_for(std::chrono::milliseconds(20));

        char* ptr = leak_function(i);
        
        if (ptr) {
            ptr[0] = 'c';
        }
    }

    running = false;
#ifdef THREADED_TEST
    for (auto& thread : threads) {
        thread.join();
    }
#endif

    return 0;
}
