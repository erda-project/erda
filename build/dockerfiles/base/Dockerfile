FROM unitedwardrobe/golang-librdkafka:alpine3.12-golang1.15.0-librdkafka1.5.0

RUN apk update \
&& apk add tzdata \
&& cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
&& echo "Asia/Shanghai" > /etc/timezone \
&& apk add curl \
&& apk add bash make git