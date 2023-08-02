include("a.jl")
include("b.jl")


test_duration = 60
exec_time_env = ENV["EXECUTION_TIME_SEC"]
if exec_time_env != nothing
    test_duration = parse(Int, exec_time_env)
    if test_duration == 0
        exit(1)
    end
end
println("Executable $(@__FILE__) starting for $test_duration seconds")
end_time = time() + test_duration
path = "/app/data/temp.txt"
while time() < end_time
    # println("start exec == a->b")
    a(path)
    b(path)
end
println("Executable $(@__FILE__) finished successfully")
rm(path)
