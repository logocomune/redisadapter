package redisadapter

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/googollee/go-socket.io"
	"github.com/oklog/ulid"
	"math/rand"
	"sync"
	"time"
)

type broadcast struct {
	instanceKey string
	rooms       map[string]map[string]socketio.Socket
	redisPull   *redis.Pool
	log         Logger
	sync.RWMutex
}

// The Conf struct is used to configure the adapter
type Conf struct {
	Host   string
	Port   string
	Prefix string
	Logger Logger
}

// The PublishedMessage is the base struct to propagate events through redis
type PublishedMessage struct {
	Ignore  interface{}
	Room    string
	Message string
	Event   string
	Args    []interface{}
}

func newPool(host string, port string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", host+":"+port) },
	}
}

var delayBetweenConnections = 500 * time.Millisecond

// NewRedisAdapter create the redis adapter for socket.io
func NewRedisAdapter(conf Conf) socketio.BroadcastAdaptor {
	b := &broadcast{
		rooms: make(map[string]map[string]socketio.Socket),
	}
	b.log = NewNoLog()
	if conf.Logger != nil {
		b.log = conf.Logger
	}

	host := "127.0.0.1"
	if conf.Host != "" {
		host = conf.Host
	}
	b.log.Debugf("Redis host: %s\n", host)

	port := "6379"
	if conf.Port != "" {
		port = conf.Port
	}
	b.log.Debugf("Redis port: %s\n", port)

	prefix := "redis_socket_io"
	if conf.Prefix != "" {
		prefix = conf.Prefix
	}
	b.log.Debugf("Redis event prefix: %s\n", prefix)

	b.log.Debugln("Start redis pool")
	b.redisPull = newPool(host, port)

	b.instanceKey = generateUniqueKey(prefix)

	b.subscriberWorker(prefix)

	return b
}

func generateUniqueKey(prefix string) string {

	t1 := time.Now()
	entropy := rand.New(rand.NewSource(t1.UnixNano()))
	return prefix + "#" + ulid.MustNew(ulid.Now(), entropy).String()

}

func (b *broadcast) subscriberWorker(prefix string) {
	go func() {
		for {
			c := b.redisPull.Get()

			subConn := redis.PubSubConn{Conn: c}
			subConn.PSubscribe(prefix + "#*")
			for c.Err() == nil {

				switch n := subConn.Receive().(type) {
				case redis.Message:
					b.log.Debugf("Redis Message - channel: '%s' data: '%s'\n", n.Channel, n.Data)
				case redis.PMessage:
					ignored := " ignored"
					if n.Channel != b.instanceKey {
						b.subscribeMessage(n.Data)
						ignored = ""
					}
					b.log.Debugf("Redis PMessage%s - pattern: '%s' channel: '%s' data: '%s'\n", ignored, n.Pattern, n.Channel, n.Data)

				case redis.Subscription:
					b.log.Debugf("Redis Subscription - kind: '%s' channel: '%s' count: '%d'\n", n.Kind, n.Channel, n.Count)

				case error:
					b.log.Errorf("Redis Receive error: %v\n", n)
				}
			}

			c.Close()
			b.log.Errorf("Subscription Connection error....retry in %s", delayBetweenConnections)
			time.Sleep(delayBetweenConnections)

		}

	}()
}

func (b *broadcast) subscribeMessage(data []byte) error {

	p := PublishedMessage{}
	err := json.Unmarshal(data, &p)
	if err != nil {

		b.log.Errorln("Error during json decode", err, string(data))
		return nil
	}

	ignore, ok := p.Ignore.(socketio.Socket)
	if !ok {
		b.log.Debugln("Ignore field is not a Socket type")
		ignore = nil
	}

	return b.broadcastLocal(ignore, p.Room, p.Event, p.Args...)
}

func (b *broadcast) Join(room string, socket socketio.Socket) error {
	b.Lock()
	defer b.Unlock()
	sockets, ok := b.rooms[room]
	if !ok {
		sockets = make(map[string]socketio.Socket)
	}
	sockets[socket.Id()] = socket
	b.rooms[room] = sockets
	return nil
}

func (b *broadcast) Leave(room string, socket socketio.Socket) error {
	b.Lock()
	defer b.Unlock()
	sockets, ok := b.rooms[room]
	if !ok {
		return nil
	}
	delete(sockets, socket.Id())
	if len(sockets) == 0 {
		delete(b.rooms, room)
		return nil
	}
	b.rooms[room] = sockets
	return nil
}

func (b *broadcast) Len(room string) int {
	b.RLock()
	defer b.RUnlock()
	return len(b.rooms[room])
}

func (b *broadcast) Send(ignore socketio.Socket, room, event string, args ...interface{}) error {
	b.broadcastLocal(ignore, room, event, args)
	remoteError := b.publishToRemote(ignore, room, event, args)
	return remoteError

}

func (b *broadcast) broadcastLocal(ignore socketio.Socket, room, event string, args ...interface{}) error {
	b.RLock()
	defer b.RUnlock()
	sockets := b.rooms[room]
	for id, s := range sockets {
		if ignore != nil && ignore.Id() == id {
			continue
		}
		s.Emit(event, args...)
	}
	return nil
}

func (b *broadcast) publishToRemote(ignore socketio.Socket, room, event string, args ...interface{}) error {

	p := PublishedMessage{
		Ignore: ignore,
		Room:   room,
		Event:  event,
		Args:   args,
	}

	buf, _ := json.Marshal(&p)
	conn := b.redisPull.Get()
	defer conn.Close()
	_, err := conn.Do("PUBLISH", b.instanceKey, buf)
	return err
}
