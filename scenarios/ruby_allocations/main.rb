require 'timeout'

# The Ruby profiler does not (yet) include a way of exporting a pprof to a file, so we implement it here:
class ExportToFile
  PPROF_PREFIX = ENV.fetch('DD_PROFILING_PPROF_PREFIX')

  def export(flush)
    File.write("#{PPROF_PREFIX}#{flush.start.strftime('%Y%m%dT%H%M%SZ')}.pprof", flush.pprof_data)
    File.write("#{PPROF_PREFIX}#{flush.start.strftime('%Y%m%dT%H%M%SZ')}.internal_metadata.json",
               flush.internal_metadata_json)
    true
  end
end

Datadog.configure do |c|
  c.profiling.enabled = true
  c.profiling.exporter.transport = ExportToFile.new
  c.telemetry.enabled = false
end

Datadog::Profiling.wait_until_running

def a
  # This is intentionally inneficient to see if we capture it. You may naively assume this leads
  # to 2 string allocations ('a' and the expansion of 'a' to size 1024 * 120) but the rules of
  # the * operator precedence means we'll do:
  # * Allocate one string directly (>a)
  # * Allocate another string of size 1024 as a result of the first call of * (>a>*)
  # * Allocate another string of size 1024*120 as a result of the second call of * (>a>*)
  'a' * 1024 * 120
end

class ComplexObject
  def initialize
    @data_storage = Array.new(1024)
  end
end

def b
  ComplexObject.new
end

$test_duration = 50
exec_time_env = ENV['EXECUTION_TIME_SEC']
if exec_time_env
  $test_duration = exec_time_env.to_i
  exit(1) if $test_duration.zero?
end

puts "Executable #{__FILE__} starting for #{$test_duration} seconds"

def allocate_stuff(loops_per_sec:)
  sum_sleep_time = 0
  end_time = Process.clock_gettime(Process::CLOCK_MONOTONIC) + $test_duration
  iterations = 0
  time_per_loop = (1.0 / loops_per_sec)
  sleep_budget = 0
  while (start = Process.clock_gettime(Process::CLOCK_MONOTONIC)) < end_time
    a # 3 String allocations per loop exec.
    b # 1 Object allocation + 1 Array allocation per loop exec.

    stop = Process.clock_gettime(Process::CLOCK_MONOTONIC)
    elapsed = stop - start
    # NOTE: When you sleep, you may actually end up sleeping for more time than you asked for.
    #       If we were purely setting expectations on relative weights we could get away
    #       with a simple `sleep(time_per_loop - elapsed)`. The end result is that we may not have
    #       actually run all the loops we would expect given our `test_duration` and `loops_per_sec`
    #       values but the relative percentages should still be ok. But in these correctness checks,
    #       we're also asserting that the total value is indeed `test_duration * loops_per_sec * allocations_per_loop`
    #       and thus we need to ensure that we actually ran for `test_duration * loops_per_sec` even in
    #       the presence of high system load or weird scheduling delays meaning our sleep calls should
    #       take into account the difference between expected vs real cumulative sleep time of past
    #       iterations (which is what the sleep_budget is for).
    sleep_budget += (time_per_loop - elapsed)
    time_to_sleep = sleep_budget.positive? ? sleep_budget : 0
    sleep(time_to_sleep)
    slept = Process.clock_gettime(Process::CLOCK_MONOTONIC) - stop
    sleep_budget -= slept

    sum_sleep_time += slept

    iterations += 1
  end
  puts "Thread #{Thread.current.name} finished with #{iterations} iterations and #{sum_sleep_time} total sleep time"
end

loops_per_sec = (ENV['LOOPS_PER_SEC'] || 100).to_i # Each loop does 5 allocations
threads = []
thread1_loops_per_sec = (loops_per_sec * 0.75).to_i # Thread 1 should use 3/4s of the flamegraph
threads << Thread.new do
  Thread.current.name = "thread#{thread1_loops_per_sec}"
  allocate_stuff(loops_per_sec: thread1_loops_per_sec)
end
thread2_loops_per_sec = (loops_per_sec * 0.25).to_i # Thread 2 should use 1/4 of the flamegraph
threads << Thread.new do
  Thread.current.name = "thread#{thread2_loops_per_sec}"
  allocate_stuff(loops_per_sec: thread2_loops_per_sec)
end

threads.each(&:join) # Total expectation of 5 * loops_per_sec allocs/sec (value-matching-sum in json)

puts "Executable #{__FILE__} finished successfully"
