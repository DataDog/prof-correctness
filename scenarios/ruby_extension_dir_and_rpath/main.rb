require 'timeout'

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

Datadog::Profiling.wait_until_running

def a
  x = 0
  i = 0
  while i < 10_000_000
    x += i
    i += 1
  end
end

def b
  x = 0
  i = 0
  while i < 20_000_000
    x += i
    i += 1
  end
end

test_duration = 5
exec_time_env = ENV['EXECUTION_TIME_SEC']
if exec_time_env
  test_duration = exec_time_env.to_i
  exit(1) if test_duration == 0
end

puts "Executable #{__FILE__} starting for #{test_duration} seconds"
end_time = Time.now + test_duration
while Time.now < end_time
  a
  b
end
puts "Executable #{__FILE__} finished successfully"
