FROM ruby:3.3

ENV DD_PROFILING_EXPERIMENTAL_ALLOCATION_ENABLED true
ENV DD_TRACE_DEBUG true
ENV DD_PROFILING_PPROF_PREFIX="/app/data/profiles_"

# Copy the Ruby program into the container
COPY ./scenarios/ruby_allocations/gems.rb ./scenarios/ruby_allocations/main.rb /app/
RUN chmod 644 /app/*

# Set the working directory to the location of the program
WORKDIR /app

RUN bundle install

# Run the program when the container starts
ENV LOOPS_PER_SEC 1000
CMD bundle exec ddtracerb exec ruby main.rb
