source 'https://rubygems.org'

# This test runs with the latest released stable and not latest master because installing from master uses different
# paths. See https://github.com/DataDog/prof-correctness/pull/39 for details.
gem 'datadog', '!= 2.23.0' # Avoid https://github.com/DataDog/dd-trace-rb/issues/5137
