FROM golang:1.17.13-bullseye as go-build
WORKDIR /src
COPY . .
RUN go env -w GOPROXY=https://goproxy.cn,direct && make bin

FROM debian:bullseye-slim
COPY --from=go-build /src/bin/manga /usr/bin/
RUN addgroup --gid 1000 manga &&  adduser --home /manga --disabled-password --gecos "" --gid 1000 --uid 1000 manga
USER manga
WORKDIR /manga
ENTRYPOINT ["/usr/bin/manga"]
CMD ["serve"]
