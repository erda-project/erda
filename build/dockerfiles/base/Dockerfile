FROM --platform=linux/amd64 golang:1.17-bullseye

# set timezone to CST
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

RUN echo "deb http://deb.debian.org/debian bookworm main" >> /etc/apt/sources.list # for libgit2-1.3
RUN curl -sL https://deb.nodesource.com/setup_current.x | bash - && apt install -y nodejs
RUN apt update && \
    apt install -y --no-install-recommends --no-install-suggests \
    # for libgit2
    cmake libssl-dev libgit2-1.3 libgit2-dev \
    jq python3-pip \
    nodejs && \
    rm -fr /var/cache/apk

RUN apt install -y libgit2-1.3

# librdkafka
# doc: https://github.com/confluentinc/confluent-kafka-go#librdkafka
# summary: Prebuilt librdkafka binaries are included with the Go client and librdkafka does not need to be installed separately on the build or target system.
# see: /go/pkg/mod/github.com/confluentinc/confluent-kafka-go@v1.5.2/kafka/librdkafka

# libgit2 - normal installation if specific version cannot install through APT.
## doc: https://github.com/libgit2/git2go#installing
## summary: This project wraps the functionality provided by libgit2. If you're using a versioned branch, install it to your system via your system's package manager and then install git2go.
#RUN cd "$HOME" && \
#    git clone https://github.com/libgit2/libgit2.git -b v1.3.1 --depth 1 && \
#    cd libgit2 && \
#    rm -rf build && mkdir build && cd build && \
#    cmake .. -DBUILD_CLAR=OFF && cmake --build . --target install
