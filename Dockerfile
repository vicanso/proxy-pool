FROM node:12-alpine as webbuilder
ADD . /proxy-pool
RUN cd /proxy-pool/web \
  && yarn \
  && yarn build \
  && rm -rf node_module  

FROM golang:1.13-alpine as builder

COPY --from=webbuilder /proxy-pool /proxy-pool

RUN apk update \
  && apk add git make \
  && go get -u github.com/gobuffalo/packr/v2/packr2 \
  && cd /proxy-pool \
  && make build

FROM alpine 

EXPOSE 4000

RUN addgroup -g 1000 go \
  && adduser -u 1000 -G go -s /bin/sh -D go \
  && apk add --no-cache ca-certificates

COPY --from=builder /proxy-pool/proxypool /usr/local/bin/proxypool
COPY --from=builder /proxy-pool/script/entrypoint.sh /entrypoint.sh

USER go

WORKDIR /home/go

HEALTHCHECK --timeout=10s CMD [ "wget", "http://127.0.0.1:4000/ping", "-q", "-O", "-"]

ENTRYPOINT ["/entrypoint.sh"]
CMD ["proxypool"]
