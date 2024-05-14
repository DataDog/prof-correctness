x = 0
i = 0

def main():
    global x, i
    for _ in range(50):
        a()
        b()
    # We add a print to prevent optimization that could turn this into a no-op program
    print(x)

def a():
    global x, i
    i = 0
    while i < 1000000:
        x += i
        i += 1

def b():
    global x, i
    i = 0
    while i < 2000000:
        x += i
        i += 1

main()
