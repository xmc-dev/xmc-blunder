# XMC's Account Service (account-srv)

`account-srv` is XMC's account managing component for users (humans) and services (machines).

## Fire it up

1. Get the code

```shell
go get github.com/xmc-dev/account-srv
```

2. Start consul

```shell
consul agent -server -data-dir /tmp/consul -bootstrap-expect 1
```

3. Start postgres (make sure the database is created and it has the `uuid-ossp`
   extension loaded)

4. Fire

```shell
./auth-srv --database_url="root:root@/auth?parseTime=True"
```

For more information regarding database URLs, click [here](https://github.com/go-sql-driver/mysql#dsn-data-source-name).

## API

See `.proto` files.

To use it:

``` shell
micro query xmc.srv.account <method> <params>
```

## To Do

* [ ] Docker file
