@noinline function a(p::String)
    x = 0
    i = 0
    while i < 10000000
        x += i
        i += 1
    end
    # write to a temporary file
    open(p, "a") do file
        write(file, "$(x)\n")
    end
end
