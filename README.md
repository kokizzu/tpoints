

## Requirements

```
# go 1.14
sudo apt install golang-go

# dependencies or use go mod vendor 
go get -u -v github.com/labstack/echo/...
go get -u -v github.com/swaggo/swag/cmd/swag
go get -u -v github.com/swaggo/echo-swagger
go get -u -v go.mongodb.org/mongo-driver/mongo
go get -u -v github.com/kokizzu/gotro/L
go get -u -v github.com/beeker1121/goque
go get -u -v github.com/json-iterator/go

# mongodb https://stackoverflow.com/a/60603587/1620210
sudo apt install mongodb
```

## testing and benchmark

```
go test

# benchmark sync processing
go test -bench=Clean -run=^$ # clean database, server must not be running
go run main.go > /dev/null
go test -bench=Sync -run=^$ # server must run

BenchmarkSync-8             1045           1020670 ns/op
BenchmarkSync-8              634           1874949 ns/op


# benchmark async processing
go test -bench=Clean -run=^$ # clean database, server must not be running
go run main.go > /dev/null
go test -bench=Async -run=^$ # server must run

BenchmarkAsync-8            4771            252726 ns/op
BenchmarkAsync-8            4081            265111 ns/op


```

## note/todos

- read not cached
- no user authentication (user id passed as parameter)
