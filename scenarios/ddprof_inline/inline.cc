#include <iostream>
#include <chrono>
#include <cstdlib>

__attribute__((noinline)) void intensiveTask(std::chrono::steady_clock::time_point endTime) {
    while (std::chrono::steady_clock::now() < endTime) {
        volatile int result = 0;
        for (int i = 0; i < 10000; i++) {
            result += i * i;
        }
    }
}

inline void moreIntensiveTask(std::chrono::steady_clock::time_point endTime) {
    while (std::chrono::steady_clock::now() < endTime) {
        volatile int result = 0;
        for (int i = 0; i < 20000; i++) {
            result += i * i;
        }
    }
}

int main() {
    const char* envVar = std::getenv("EXECUTION_TIME_SEC");
    if (!envVar) {
        std::cerr << "EXECUTION_TIME_SEC environment variable is not set." << std::endl;
        return 1;
    }

    int executionTime = std::atoi(envVar);
    for (int i = 0; i < executionTime; ++i) {
        auto totalStartTime = std::chrono::steady_clock::now();
        auto intensiveEndTime = totalStartTime + std::chrono::milliseconds(333); // 1/3 of a second
        intensiveTask(intensiveEndTime);

        auto moreIntensiveEndTime = totalStartTime + std::chrono::milliseconds(1000); // full second
        moreIntensiveTask(moreIntensiveEndTime);
    }
    return 0;
}
