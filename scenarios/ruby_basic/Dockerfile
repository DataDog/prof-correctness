FROM ruby:2.7

ENV DD_PROFILING_FORCE_ENABLE_NEW true
ENV DD_TRACE_DEBUG true
ENV DD_PROFILING_PPROF_PREFIX="/app/data/profiles_"

# Copy the Ruby program into the container
COPY ./scenarios/ruby_basic/gems.rb ./scenarios/ruby_basic/main.rb /app/
RUN chmod 644 /app/*

# Set the working directory to the location of the program
WORKDIR /app

RUN bundle install

# Run the program when the container starts
CMD bundle exec ddtracerb exec ruby main.rb
