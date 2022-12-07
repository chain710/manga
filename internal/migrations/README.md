# database migration

## tool usage

see doc: https://github.com/golang-migrate/migrate/tree/master/cmd/migrate

## create new migration

create empty migration(up and down) sql file

```
migrate create -dir ./internal/migrations/pg/ -ext sql -seq -digits 6 $name
```


## migrate up

```bash
migrate -source file:///path/pg -database "postgres://user:pass@localhost:5432/db?sslmode=disable" up 
```

## update requirements

### v2

run ```create extension pg_jieba``` as superuser in database
