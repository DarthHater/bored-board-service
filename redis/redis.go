// Liberally stole a bunch of this from a Heroku example, thank you Heroku lords
package redis

import (
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/websocket"
	"log"
)

type RedisReciever struct {
	pool           *redis.Pool
	messages       chan []byte
	newConnections chan *websocket.Conn
	rmConnections  chan *websocket.Conn
}

func NewRedisReciever(pool *redis.Pool) RedisReciever {
	return RedisReciever{
		pool:           pool,
		messages:       make(chan []byte, 1000),
		newConnections: make(chan *websocket.Conn),
		rmConnections:  make(chan *websocket.Conn),
	}
}

func (rr *RedisReciever) Run(channel string) error {
	conn := rr.pool.Get()
	defer conn.Close()
	psc := redis.PubSubConn{Conn: conn}
	psc.Subscribe(channel)
	go rr.ConnHandler()
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			// if _, err := validateMessage(v.Data); err != nil {
			// 	l.WithField("err", err).Error("Error unmarshalling message from Redis")
			// 	continue
			// }
			rr.Broadcast(v.Data)
		case redis.Subscription:
			log.Printf("Redis Subscription > kind:%s, count:%d", v.Kind, v.Count)
		case error:
			return v
		default:
			// l.WithField("v", v).Info("Unknown Redis receive during subscription")
		}
	}
}

func (rr *RedisReciever) Broadcast(msg []byte) {
	rr.messages <- msg
}

func (rr *RedisReciever) Register(conn *websocket.Conn) {
	rr.newConnections <- conn
}

func (rr *RedisReciever) DeRegister(conn *websocket.Conn) {
	rr.rmConnections <- conn
}

func (rr *RedisReciever) ConnHandler() {
	conns := make([]*websocket.Conn, 0)
	for {
		select {
		case msg := <-rr.messages:
			for _, conn := range conns {
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					log.Print("Error writing data to connection")
					conns = RemoveConn(conns, conn)
				}
			}
		case conn := <-rr.newConnections:
			conns = append(conns, conn)
		case conn := <-rr.rmConnections:
			conns = RemoveConn(conns, conn)
		}
	}
}

func RemoveConn(conns []*websocket.Conn, remove *websocket.Conn) []*websocket.Conn {
	var i int
	var found bool
	for i = 0; i < len(conns); i++ {
		if conns[i] == remove {
			found = true
			break
		}
	}
	if !found {
		panic("Conns not found")
	}
	copy(conns[i:], conns[i+1:])
	conns[len(conns)-1] = nil
	return conns[:len(conns)-1]
}

type RedisWriter struct {
	pool     *redis.Pool
	messages chan []byte
}

func NewRedisWriter(pool *redis.Pool) RedisWriter {
	return RedisWriter{
		pool:     pool,
		messages: make(chan []byte, 1000),
	}
}

func (rw *RedisWriter) Run(channel string) error {
	conn := rw.pool.Get()
	defer conn.Close()

	for data := range rw.messages {
		if err := WriteToRedis(conn, data, channel); err != nil {
			rw.Publish(data)
			return err
		}
	}
	return nil
}

func WriteToRedis(conn redis.Conn, data []byte, channel string) error {
	if err := conn.Send("PUBLISH", channel, data); err != nil {
		return err
	}
	if err := conn.Flush(); err != nil {
		return err
	}
	return nil
}

func (rw *RedisWriter) Publish(data []byte) {
	rw.messages <- data
}
