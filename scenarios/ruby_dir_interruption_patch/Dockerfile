FROM ruby:3.3

ENV DD_PROFILING_ENABLED true
ENV DD_PROFILING_DIR_INTERRUPTION_WORKAROUND_ENABLED true

COPY ./scenarios/ruby_dir_interruption_patch/*.rb /app/
RUN chmod 755 /app/*

WORKDIR /app
RUN bundle install

# The `DirMonkeyPatch` issue does not affect Ruby 3.4+. We thus pin to this spec version, to avoid breakages from
# ruby-spec/mspec changes for newer Rubies (and to avoid issues once 3.3 support gets dropped).
RUN wget https://github.com/ruby/spec/archive/7240267784230e3e0bb9181dd4ea82216d136956.zip -O spec.zip
RUN wget https://github.com/ruby/mspec/archive/dffcdf7d00612c13465983224203e457a30cf124.zip -O mspec.zip
RUN unzip spec.zip
RUN unzip mspec.zip

# Since we aggressively drop permissions when running, we need to manually create a folder for rubyspec to use
RUN mkdir /app/rubyspec_temp
RUN chmod 777 /app/rubyspec_temp

# Needed by one of the ruby specs (which is testing access to home)
# @ivoanjo: Locally I get uid 1000, and in CI I'm seeing 1001. I am not sure why they differ, or if any other uids
# can show up. But, if we're missing this for an user, it'll cause the tests to fail so we definitely won't miss it.
RUN useradd -u 1000 -ms /bin/bash someuser
RUN useradd -u 1001 -ms /bin/bash someuser2

CMD bundle exec ddprofrb exec ruby mspec-dffcdf7d00612c13465983224203e457a30cf124/bin/mspec-run --config mspec_config.rb spec-7240267784230e3e0bb9181dd4ea82216d136956/core/dir/
