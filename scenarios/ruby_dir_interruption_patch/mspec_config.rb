Datadog::Profiling.wait_until_running

if Dir.ancestors.first == Datadog::Profiling::Ext::DirInstanceMonkeyPatches &&
  Dir.singleton_class.ancestors.first == Datadog::Profiling::Ext::DirClassMonkeyPatches
  puts "Dir interruption patch is present!"
else
  raise 'Dir interruption patch is not present!'
end
