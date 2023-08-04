#include <time.h>
#include <stdint.h>
#include <stdlib.h>
#include <stdio.h>

void a() {
    int64_t x = 0;
    int64_t i = 0;
    while (i < 100000000) {
        x += i;
        i += 1;
    }
}

void b() {
    int64_t x = 0;
    int64_t i = 0;
    while (i < 200000000) {
        x += i;
        i += 1;
    }
}

int main(int argc, char *argv[]) {
    int test_duration = 60;
    const char *exec_time_env = getenv("EXECUTION_TIME_SEC");
    if (exec_time_env) {
        test_duration = atoi(exec_time_env);
        if (test_duration == 0) {
            exit(1);
        }
    }
    printf("Executable %s starting for %d seconds\n", argv[0], test_duration);
    time_t end = time(NULL) + test_duration;
    while (time(NULL) < end) {
        a();
        b();
    }
    printf("Executable %s finished successfully\n", argv[0]);
    return 0;
}
