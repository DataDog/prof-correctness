# Docker image that allows to run docker in gitlab
FROM 486234852809.dkr.ecr.us-east-1.amazonaws.com/docker:20.10.3
USER root

ENV DEBIAN_FRONTEND=noninteractive

RUN set -x \
    && apt-get update \
    && apt-get -y install --no-install-recommends build-essential curl php-cli python3 python3-pip xz-utils \
    && apt-get -y clean \
    && rm -rf /var/lib/apt/lists/*
RUN set -x \
    && update-alternatives --install /usr/bin/python python /usr/bin/python3 1

RUN set -x \
    && curl -OL "https://go.dev/dl/go1.25.1.linux-amd64.tar.gz" \
    && echo "7716a0d940a0f6ae8e1f3b3f4f36299dc53e31b16840dbd171254312c41ca12e go1.25.1.linux-amd64.tar.gz" | sha256sum -c - \
    && tar -C /usr/local -xzf go1.25.1.linux-amd64.tar.gz \
    && chmod +x /usr/local/go/bin/go \
    && ln -s /usr/local/go/bin/go /usr/local/bin/go \
    && rm go1.25.1.linux-amd64.tar.gz

