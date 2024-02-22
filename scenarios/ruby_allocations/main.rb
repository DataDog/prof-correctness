# The Ruby profiler does not (yet) include a way of exporting a pprof to a file, so we implement it here:
class ExportToFile
  PPROF_PREFIX = ENV.fetch('DD_PROFILING_PPROF_PREFIX')

  def export(flush)
    File.write("#{PPROF_PREFIX}#{flush.start.strftime('%Y%m%dT%H%M%SZ')}.pprof", flush.pprof_data)
    true
  end
end

Datadog.configure do |c|
  c.profiling.enabled = true
  c.profiling.exporter.transport = ExportToFile.new
end

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
    sleep_budget += time_per_loop
    iterations += 1
    elapsed = stop - start
    sleep_budget -= elapsed
    time_to_sleep = sleep_budget.positive? ? sleep_budget : 0
    sleep(time_to_sleep)
    slept = Process.clock_gettime(Process::CLOCK_MONOTONIC) - stop
    sleep_budget -= slept
    sum_sleep_time += slept
  end
  puts "Thread #{Thread.current.name} finished with #{iterations} iterations and #{sum_sleep_time} total sleep time"
end

threads = []
threads << Thread.new do
  Thread.current.name = 'thread75'
  allocate_stuff(loops_per_sec: 75) # 75 * 5=375 allocs/sec
end
threads << Thread.new do
  Thread.current.name = 'thread25'
  allocate_stuff(loops_per_sec: 25) # 25 * 5=125 allocs/sec
end

threads.each(&:join) # Total expectation of 500 allocs/sec

puts "Executable #{__FILE__} finished successfully"
