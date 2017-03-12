FROM alpine:3.5
MAINTAINER SocialEngine

ENV DUMP_INIT_VERSION 1.2.0

ADD dist/rancher-cron /usr/bin/rancher-cron
ADD https://github.com/Yelp/dumb-init/releases/download/v${DUMP_INIT_VERSION}/dumb-init_${DUMP_INIT_VERSION}_amd64 /usr/bin/dumb-init

RUN \
  adduser -s /sbin/nologin -u 1000 -D rancher-cron && \
  apk add --no-cache ca-certificates && \
  chmod +x /usr/bin/dumb-init

USER rancher-cron

CMD [ "dumb-init", "rancher-cron" ]
