# Redis adapter for go-socket.io

[![Build Status](https://travis-ci.org/logocomune/redisadapter.svg?branch=master)](https://travis-ci.org/logocomune/redisadapter)

It's a redis adapter for [go-socket.io](https://github.com/googollee/go-socket.io).

By running go-socket.io with this adapter, you can run multiple socket.io instances in different processes or servers that can all broadcast and emit events to and from each other.

## Install

```bash
go get "github.com/logocomune/redisadapter"
```


## Configuration

```go
type Conf struct {
	Host   string
	Port   string
	Prefix string
	Logger Logger
}
```
- Host: host to connect to redis on (127.0.0.1)
- Port: port to connect to redis on (6379)
- Prefix: prefix for keys on publish/subscribe events (redis_socket_io)
- Logger: logger (redisadapter.NewNoLog())
   - redisadapter.NewStdLog(): enable all logs
   - redisadapter.NewNoLog(): disable all logs
   - logrus.New(): it's compatible with redisadapter.Logger interface
    
### Logger interface
```go
type Logger interface {
	Debugf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	Debug(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})

	Debugln(args ...interface{})
	Warnln(args ...interface{})
	Errorln(args ...interface{})
}

```


## Usage

```go
package main

import (
    "net/http"
    "log"
    "github.com/googollee/go-socket.io"
    "github.com/logocomune/redisadapter"
)

func main() {
    server, err := socketio.NewServer(nil)
   	if err != nil {
   		log.Fatal(err)
   	}

    conf := redisadapter.Conf{
    		Host:"localhost",
    		Prefix:"my_chat_app",
    		Logger: redis.NewStdLog(),
    	}
    server.SetAdaptor(redisadapter.NewRedisAdapter(conf))

    server.On("connection", func(so socketio.Socket) {
        
    	so.Join("room_1")
        
        so.On("chat message", func(msg string) {
            
           log.Println("Msg: ",msg)
        })
        
        so.On("disconnection", func() {
            so.BroadcastTo("room_1", "chat message", "Good bye")
        })
    })
    server.On("error", func(so socketio.Socket, err error) {
        log.Println("error:", err)
    })

    http.Handle("/socket.io/", server)
    log.Println("Serving at localhost:8008...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```
## Inspired and influenced

- https://github.com/satyakb/go-socket.io-redis
- https://github.com/Automattic/socket.io-redis
