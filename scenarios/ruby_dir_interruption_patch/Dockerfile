FROM ruby:3.3

ENV DD_PROFILING_ENABLED true
ENV DD_PROFILING_DIR_INTERRUPTION_WORKAROUND_ENABLED true

COPY ./scenarios/ruby_dir_interruption_patch/*.rb /app/
RUN chmod 755 /app/*

WORKDIR /app
RUN bundle install

RUN wget https://github.com/ruby/spec/archive/refs/heads/master.zip -O spec-master.zip
RUN wget https://github.com/ruby/mspec/archive/refs/heads/master.zip -O mspec-master.zip
RUN unzip spec-master.zip
RUN unzip mspec-master.zip

# Since we aggressively drop permissions when running, we need to manually create a folder for rubyspec to use
RUN mkdir /app/rubyspec_temp
RUN chmod 777 /app/rubyspec_temp

# Needed by one of the ruby specs (which is testing access to home)
# @ivoanjo: Locally I get uid 1000, and in CI I'm seeing 1001. I am not sure why they differ, or if any other uids
# can show up. But, if we're missing this for an user, it'll cause the tests to fail so we definitely won't miss it.
RUN useradd -u 1000 -ms /bin/bash someuser
RUN useradd -u 1001 -ms /bin/bash someuser2

CMD bundle exec ddprofrb exec ruby mspec-master/bin/mspec-run --config mspec_config.rb spec-master/core/dir/
