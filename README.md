# MangaDepot

## what

Manga reader in browser. Written in go

inspired by [kavitareader](https://www.kavitareader.com/) and [komga](https://komga.org/)

copy most front end code from komga, with improvement in reader and better archive support

## limitation

- no user management nor authentication
- no support for files other than archive(zip, rar)
- database: only support postrges

## build

```bash
# build docker image
docker build --no-cache --progress=plain -t chain710/manga-depot:latest .
```

## run

```bash
# setup database
docker run -it --rm chain710/manga-depot:v0.0.1 migrate up --dsn 'postgres://manga:123456@localhost:5432/manga?sslmode=disable'
# run service
docker run -d -v /host_books:/container_books -p 8080:8080 chain710/manga-depot:v0.0.1 serve --dsn 'postgres://manga:123456@localhost:5432/manga?sslmode=disable'
```

## TODO 

### server
- [ ] improve cover selection
- [ ] list ongoing tasks(scan lib mostly)
- [ ] code refine: sql builder

### front end
- [ ] reader original mode
- [ ] book summary responsive improve
- [ ] book meta edit