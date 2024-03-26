FROM ruby:3.3

ENV DD_PROFILING_ALLOCATION_ENABLED true
ENV DD_PROFILING_EXPERIMENTAL_HEAP_ENABLED true
ENV DD_PROFILING_EXPERIMENTAL_HEAP_SIZE_ENABLED true
ENV DD_TRACE_DEBUG true
ENV DD_PROFILING_PPROF_PREFIX="/app/data/profiles_"

# Copy the Ruby program into the container
COPY ./scenarios/ruby_heap/gems.rb ./scenarios/ruby_heap/main.rb /app/
RUN chmod 644 /app/*

# Set the working directory to the location of the program
WORKDIR /app

RUN bundle install

# By disabling reshuffles, we will always do allocations in the same order
# and thus (theoretically) increase the effect of bias from interval sampling.
ENV DISABLE_RESHUFFLE="true"

# Run the program when the container starts
CMD bundle exec ddtracerb exec ruby main.rb
