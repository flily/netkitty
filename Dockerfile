# Github: https://github.com/flily/netkitty

######## build phase ########

FROM golang:1.16.5-alpine3.13 AS build
WORKDIR /build
COPY . /build/
RUN go build -o netkitty

######## package phase ########

FROM alpine:3.13.5
COPY --from=build /build/netkitty /usr/local/bin/netkitty
CMD [ "netkitty" ]
