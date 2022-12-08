FROM golang:alpine as build
COPY . /src/
WORKDIR /src
RUN CGO_ENABLED=0 go build -o /bin/server
RUN rm -rf /src

FROM alpine:3.16.0
COPY --from=build /bin/server /bin/server
RUN apk add --no-cache tini
EXPOSE 8080
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/bin/server"]