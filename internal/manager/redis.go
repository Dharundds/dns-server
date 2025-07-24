package manager

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type Redis struct {
	client *redis.Client

	host           string
	port           string
	password       string
	db             int
	maxConn        int
	minConn        int
	maxRetries     int
	connTimeout    time.Duration
	readTimeout    time.Duration
	writeTimeout   time.Duration
	reconnectDelay time.Duration
}

type RedisOption func(r *Redis)

func NewRedisManager(opts ...RedisOption) (*Redis, error) {
	log.Info().Msg("Creating RedisManager")

	r := &Redis{
		host:           "localhost",
		port:           "6379",
		password:       "",
		db:             0,
		maxConn:        5,
		minConn:        1,
		maxRetries:     3,
		connTimeout:    10 * time.Second,
		readTimeout:    5 * time.Second,
		writeTimeout:   5 * time.Second,
		reconnectDelay: 10 * time.Second,
	}

	for _, opt := range opts {
		opt(r)
	}

	// Create connection immediately
	err := r.createNewConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis connection: %w", err)
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	log.Info().Msg("Redis connection established successfully")
	return r, nil
}

func WithRedisAddress(host, port string) RedisOption {
	return func(r *Redis) {
		r.host = host
		r.port = port
	}
}

func WithRedisAuth(password string, db int) RedisOption {
	return func(r *Redis) {
		r.password = password
		r.db = db
	}
}

func WithRedisPoolConfig(maxConn, minConn int) RedisOption {
	return func(r *Redis) {
		r.maxConn = maxConn
		r.minConn = minConn
	}
}

func WithRedisTimeouts(connTimeout, readTimeout, writeTimeout time.Duration) RedisOption {
	return func(r *Redis) {
		r.connTimeout = connTimeout
		r.readTimeout = readTimeout
		r.writeTimeout = writeTimeout
	}
}

func (r *Redis) GetClient() *redis.Client {
	return r.client
}

func (r *Redis) Name() string {
	return "Redis"
}

func (r *Redis) Reconnect(ctx context.Context) error {
	// Close existing connection
	if r.client != nil {
		r.client.Close()
	}

	err := r.createNewConnection()
	if err != nil {
		return err
	}

	return r.client.Ping(ctx).Err()
}

// Pinger implementation Ends -------

func (r *Redis) createNewConnection() error {
	log.Info().Msgf("Creating Redis connection to %s:%s", r.host, r.port)

	opts := &redis.Options{
		Addr:         fmt.Sprintf("%s:%s", r.host, r.port),
		Password:     r.password,
		DB:           r.db,
		PoolSize:     r.maxConn,
		MinIdleConns: r.minConn,
		MaxRetries:   r.maxRetries,
		DialTimeout:  r.connTimeout,
		ReadTimeout:  r.readTimeout,
		WriteTimeout: r.writeTimeout,
	}

	r.client = redis.NewClient(opts)
	return nil
}

// Redis operations

// Get retrieves a value by key
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	result, err := r.client.Get(ctx, key).Result()
	return result, err
}

// Set stores a key-value pair with optional TTL
func (r *Redis) Set(ctx context.Context, key string, value any, ttl ...time.Duration) error {
	var expiration time.Duration
	if len(ttl) > 0 {
		expiration = ttl[0]
	}
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Del deletes one or more keys
func (r *Redis) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists checks if keys exist
func (r *Redis) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

// Expire sets a timeout on a key
func (r *Redis) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// Keys returns all keys matching a pattern
func (r *Redis) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}

// ScanKeys returns keys using SCAN command (more efficient for large datasets)
func (r *Redis) ScanKeys(ctx context.Context, pattern string) ([]string, error) {
	var keys []string
	var cursor uint64

	for {
		var scanKeys []string
		var err error
		scanKeys, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, scanKeys...)
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}

func (r *Redis) Close() {
	if r.client != nil {
		r.client.Close()
	}
}
