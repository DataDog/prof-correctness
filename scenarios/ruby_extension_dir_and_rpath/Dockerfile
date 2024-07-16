FROM ruby:3.3

ENV DD_TRACE_DEBUG true
ENV DD_PROFILING_PPROF_PREFIX="/app/data/profiles_"
ENV EXECUTION_TIME_SEC 5

# Copy the Ruby program into the container
COPY ./scenarios/ruby_extension_dir_and_rpath/gems.rb ./scenarios/ruby_basic/main.rb /app/
RUN chmod 644 /app/*

# Set the working directory to the location of the program
WORKDIR /app

RUN bundle config set --local path 'vendor/bundle'
RUN bundle install
RUN rm -f vendor/bundle/ruby/*/gems/*/lib/datadog*so vendor/bundle/ruby/*/bundler/gems/*/lib/datadog*so
RUN bundle config set --local path 'vendor-moved/bundle'
RUN mv vendor vendor-moved

# Run the program when the container starts
CMD bundle exec ddprofrb exec ruby main.rb
