require 'objspace'
require 'timeout'

# For expecation calculations refer to
# https://docs.google.com/spreadsheets/d/1ZrR2UYsU78SnvBdE4OMOHeY5yKX7diussurgr2NAlQs

$community_pool = {
  a: [],
  b: [],
  c: [],
  d: []
}

# The Ruby profiler does not (yet) include a way of exporting a pprof to a file, so we implement it here:
class ExportToFile
  PPROF_PREFIX = ENV.fetch('DD_PROFILING_PPROF_PREFIX')

  def export(flush)
    $community_pool.each_entry do |element|
      puts("#{element[0]} #{element[1].size} #{ObjectSpace.memsize_of(element[1].last)}")
    end
    File.write("#{PPROF_PREFIX}#{flush.start.strftime('%Y%m%dT%H%M%SZ')}.pprof", flush.pprof_data)
    File.write("#{PPROF_PREFIX}#{flush.start.strftime('%Y%m%dT%H%M%SZ')}.internal_metadata.json",
               flush.internal_metadata_json)
    true
  end
end

Datadog.configure do |c|
  c.profiling.enabled = true
  c.profiling.exporter.transport = ExportToFile.new
end

setup_end = Process.clock_gettime(Process::CLOCK_MONOTONIC)

Datadog::Profiling.wait_until_running

# Approx 1KB per living object (memsize_of says 1041) (e.g. 5MB for 5k live objects)
def a
  $community_pool[:a] << ('a' * 1000)
end

# Approx 40 bytes + 100 * 64 bits=840 bytes per living object (e.g. 8MB for 10k live objects)
# https://ivoanjo.me/blog/2021/02/11/looking-into-array-memory-usage/
def b
  $community_pool[:b] << Array.new(100)
end

# Basic object data (40 bytes) + separate IV array
#
# https://github.com/ruby/ruby/blob/57748ef2a26cbdfe075347c6f7064eb419b4949a/gc.c#L4904
#
# There are a total of 10 fields.
# But capacities grow in powers of 2 so capacity will be 16
#
# Total size per live object = 40 + 16 * 8 = ~168 bytes (memsize_of says 160 bytes)
# (e.g. for 25000 entries = ~4.2MB)
class ObjectWithShape
  def initialize
    @field1 = 1
    @field2 = 2
    @field3 = 3
    @field4 = 4
    @field5 = 5
    @field6 = 6
    @field7 = 7
    @field8 = 8
    @field9 = 9
    @field10 = 10
  end
end

# Basic object data (40 bytes) + hash table capable of holding 10 fields
#
# https://github.com/ruby/ruby/blob/57748ef2a26cbdfe075347c6f7064eb419b4949a/gc.c#L4901
#
# Hash table holding 10 fields
# * https://github.com/ruby/ruby/blob/2d80b6093f3b0c21c89db72eebacfef4a535b149/st.c#L666-L673
# * 2^3 < 10 < 2^4 => entry_power = 4
# * 4 words*sizeof(st_index_t)=4*8=32 bytes for storing bins
# * 16 allocated entries * sizeof(st_table_entry) = 16 * 3 * 8 = 384 bytes for storing entries
# * sizeof(st_table) ~= 55 bytes
#
# Total size per live object = 40 + 55 + 32 + 384 = ~511 bytes (memsize_of says 600 bytes)
# (e.g. for 10000 entries = ~5.1MB)
class TooComplexObject
  def initialize(shape_init_i)
    10.times do |i|
      instance_variable_set("@field_shape_#{shape_init_i}_#{i}", i)
    end
  end
end

# We're creating TooComplexObject with 8 different shapes initially
# to trigger this: https://github.com/ruby/ruby/blob/609bbad15da6fe91904bdcd139f9e24e3cf61d4b/shape.c#L733-L741
# After that, any new shape (such as the default triggered by the constructor with no parameters) will be
# tagged as SHAPE_OBJ_TOO_COMPLEX which triggers storage of fields in a st_table.
8.times { |i| TooComplexObject.new(i) }

def c
  $community_pool[:c] << ObjectWithShape.new
end

def d
  $community_pool[:d] << TooComplexObject.new(8)
end

test_duration = 50
exec_time_env = ENV['EXECUTION_TIME_SEC']
if exec_time_env
  test_duration = exec_time_env.to_i
  exit(1) if test_duration.zero?
end

loops_per_sec = (ENV['LOOPS_PER_SEC'] || 50).to_i
total_loops = loops_per_sec * test_duration
toss_into_pool_per_loop = {
  a: 2,
  b: 4,
  c: 8,
  d: 4
}
drown_per_loop = {
  a: 1,
  b: 2,
  c: 3,
  d: 2
}
toss_order = toss_into_pool_per_loop.keys
expected_swimming_at_end = toss_into_pool_per_loop.keys.map do |k|
  [k, (toss_into_pool_per_loop[k] - drown_per_loop[k]) * total_loops]
end.to_h

puts "Executable #{__FILE__} starting for #{test_duration} seconds"

sum_sleep_time = 0
end_time = Process.clock_gettime(Process::CLOCK_MONOTONIC) + test_duration
iterations = 0
time_per_loop = (1.0 / loops_per_sec)
disable_reshuffle = ENV['DISABLE_RESHUFFLE'] == 'true'
puts "Reshuffles enabled?: #{disable_reshuffle ? 'no' : 'yes'}"
reshuffle_secs = 0.5
gc_every_secs = 1
sleep_budget = 0
while (start = Process.clock_gettime(Process::CLOCK_MONOTONIC)) < end_time
  # Shift allocationg ordering a bit. In a real app there would naturally be some
  # variance to requests/tasks but here each loop looks much the same as the one before.
  # This makes it easy to reason about expectations but, in combination with the interval
  # sampling done by allocation+heap sampling, may make it extra prone to biases.
  # By shuffling the tossing order we replicate some of the variety you'd see in real
  # workloads.
  toss_order.shuffle if !disable_reshuffle && (iterations % (loops_per_sec * reshuffle_secs).to_i).zero?
  toss_order.each { |sym| toss_into_pool_per_loop[sym].times { Object.send(sym) } }

  # Remove some objects from the pool
  drown_per_loop.each_entry do |sym, num_losses|
    $community_pool[sym].pop(num_losses)
  end

  GC.start if (iterations % (loops_per_sec * gc_every_secs)).zero?

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

GC.start

expected_swimming_at_end.each_entry do |k, expected|
  actual = $community_pool[k].size
  puts "Expected #{expected} '#{k}'s in the pool at the end, found #{actual}" if actual != expected
end

puts "Executable #{__FILE__} finished successfully in #{iterations} iterations and after sleeping #{sum_sleep_time} seconds"
puts "Time since setup end: #{Process.clock_gettime(Process::CLOCK_MONOTONIC) - setup_end} secs"
