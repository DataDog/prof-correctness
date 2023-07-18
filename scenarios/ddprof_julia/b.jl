@noinline function b(p::String)
    x = 0
    i = 0
    while i < 20000000
        x += i
        i += 1
    end
    open(p, "a") do file
        write(file, "$(x)\n")
    end
end
