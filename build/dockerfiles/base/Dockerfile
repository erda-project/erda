FROM amd64/golang:1.16.3-alpine3.12 as golang-image
FROM unitedwardrobe/golang-librdkafka:alpine3.12-golang1.15.0-librdkafka1.5.0

# use aliyun repo
RUN echo "http://mirrors.aliyun.com/alpine/v3.12/main/" > /etc/apk/repositories && \
    echo "http://mirrors.aliyun.com/alpine/v3.12/community/" >> /etc/apk/repositories

RUN apk update \
&& apk add tzdata \
&& cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
&& echo "Asia/Shanghai" > /etc/timezone \
&& apk add bash curl make cmake git openssh libressl-dev make curl libcurl alpine-sdk pkgconfig tzdata openjdk8 && \
   rm -rf /var/cache/apk/*

ENV PKG_CONFIG_PATH /lib64/pkgconfig:$PKG_CONFIG_PATH

RUN git clone git://github.com/libgit2/libgit2.git && \
    cd ./libgit2 && \
    git checkout v1.0.1 && rm -rf build && mkdir build && cd build && \
    cmake .. -DCMAKE_INSTALL_PREFIX=$TARGET -DBUILD_CLAR=OFF && cmake --build . --target install

RUN rm -rf /usr/local/go
COPY --from=golang-image "/usr/local/go" "/usr/local/go"
