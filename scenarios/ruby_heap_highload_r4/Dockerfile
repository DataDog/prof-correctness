FROM ruby:4.0.1

ENV DD_PROFILING_ALLOCATION_ENABLED true
ENV DD_PROFILING_EXPERIMENTAL_HEAP_ENABLED true
ENV DD_PROFILING_EXPERIMENTAL_HEAP_SIZE_ENABLED true
ENV DD_TRACE_DEBUG true
ENV DD_PROFILING_PPROF_PREFIX="/app/data/profiles_"
ENV DD_PROFILING_EXPERIMENTAL_HEAP_RUBY4_ENABLED true

# Copy the Ruby program into the container
COPY ./scenarios/ruby_heap_r4/gems.rb ./scenarios/ruby_heap/main.rb /app/
RUN chmod 644 /app/*

# Set the working directory to the location of the program
WORKDIR /app

RUN bundle install

ENV LOOPS_PER_SEC=500

# Run the program when the container starts
CMD bundle exec ddprofrb exec ruby main.rb
