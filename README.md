# MangaDepot

![alt screen](docs/screen.png)

## what

Simple manga reader in browser. Written in go

inspired by [kavitareader](https://www.kavitareader.com/) and [komga](https://komga.org/)

copy most front end code from komga, with improvement in reader and better archive support

## limitation

- no user management nor authentication
- only support zip/rar files (files supported by [go-unarr](https://github.com/gen2brain/go-unarr)).
- database: only support postrges

## build

```bash
# build docker image
docker build --no-cache --progress=plain -t chain710/manga-depot:latest .
```

## run

```bash
# setup database
docker run -it --rm manga-depot:v0.0.1 migrate up --dsn 'postgres://manga:123456@localhost:5432/manga?sslmode=disable'
# create extension if necessary
# in postgres with superuser
create extension pg_jieba
# run service
docker run -d -v /host_books:/container_books -p 8080:8080 manga-depot:v0.0.1 serve --dsn 'postgres://manga:123456@localhost:5432/manga?sslmode=disable'
```

## TODO

### server

- [ ] improve smart cover selection
- [ ] list ongoing tasks(scan lib mostly)
- [ ] better sql builder

### front end

- [ ] mark volume read/unread
- [ ] reader original mode
- [ ] book meta edit
