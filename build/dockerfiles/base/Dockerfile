# syntax = registry.erda.cloud/retag/dockerfile:latest
FROM m.daocloud.io/docker.io/bitnami/git:latest AS git-from

FROM --platform=$TARGETPLATFORM m.daocloud.io/docker.io/golang:1.24.6-bookworm
ARG TARGETPLATFORM

# fetch latest git (which is not available in debian stable repo now) at first for later use
RUN apt remove -yq git
COPY --from=git-from /opt/bitnami/git /opt/bitnami/git
ENV PATH="/opt/bitnami/git/bin:${PATH}"
# Set safe.directory to prevent "dubious ownership" errors, see: https://github.com/actions/runner-images/issues/6775#issuecomment-1410270956
# use `--system` instead of `--global` to make it work for all users
RUN mkdir -p /opt/bitnami/git/etc && touch /opt/bitnami/git/etc/gitconfig && git config --system --add safe.directory '*'

# set timezone to CST
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

RUN apt update && apt install -y --no-install-recommends --no-install-suggests \
    cmake \
    gcc \
    g++ \
    ca-certificates \
    curl \
    gnupg \
    lsb-release \
    libssl-dev \
    vim \
    jq \
    python3-pip \
    librdkafka-dev \
    net-tools \
    nginx \
    nginx-extras \
 && rm -fr /var/lib/apt/lists/*

# nodejs
RUN curl -sL https://deb.nodesource.com/setup_lts.x | bash - && apt update && apt install -y nodejs && rm -fr /var/lib/apt/lists/*

# docker-cli
RUN mkdir -p /etc/apt/keyrings \
 && curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg \
 && echo \
    "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
    $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null \
 && apt update && apt install -y --no-install-recommends --no-install-suggests \
    docker-ce-cli \
 && rm -fr /var/lib/apt/lists/*

# librdkafka
# doc: https://github.com/confluentinc/confluent-kafka-go#librdkafka
# summary: Prebuilt librdkafka binaries are included with the Go client and librdkafka does not need to be installed separately on the build or target system.
# see: /go/pkg/mod/github.com/confluentinc/confluent-kafka-go@v1.5.2/kafka/librdkafka
# ------
# UPDATE: ARM arch doesn't have prebuilt binary.
# check $TARGETPLATFORM equal to linux/arm64

# libgit2 - normal installation if specific version cannot install through APT.
## doc: https://github.com/libgit2/git2go#installing
## summary: This project wraps the functionality provided by libgit2. If you're using a versioned branch, install it to your system via your system's package manager and then install git2go.
ENV TARGET /usr
RUN cd "$HOME" && \
    git clone https://github.com/libgit2/libgit2.git -b v1.3.2 --depth 1 && \
    cd libgit2 && \
    rm -rf build && mkdir build && cd build && \
    cmake .. -DBUILD_CLAR=OFF -DCMAKE_INSTALL_PREFIX=$TARGET && cmake --build . --target install
