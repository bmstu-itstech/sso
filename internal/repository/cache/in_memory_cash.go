package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bmstu-itstech/sso/internal/config"
	"github.com/bmstu-itstech/sso/internal/domain/models"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

func NewAppCache(cfg config.Config) (*AppCache, error) {
	connectionString := cfg.GetPostgresPath()
	db, err := sqlx.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	cache := AppCache{
		mu:   &sync.RWMutex{},
		data: make(map[int32]models.AppRepos),
		db:   db,
	}
	err = cache.Reload()
	if err != nil {
		return nil, err
	}
	go UpdateListener(connectionString, &cache)
	return &cache, nil
}

type AppCache struct {
	mu   *sync.RWMutex
	data map[int32]models.AppRepos
	db   *sqlx.DB
}

func (c *AppCache) App(ctx context.Context, appId int32) (models.AppRepos, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data[appId], nil
}

func (c *AppCache) Reload() error {
	var rows []models.AppRepos
	err := c.db.Select(&rows, "SELECT * FROM apps")
	if err != nil {
		return err
	}

	newData := make(map[int32]models.AppRepos, len(rows))
	for _, row := range rows {
		newData[row.Id] = row
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = newData
	fmt.Println("Cash reloaded")
	return nil
}

func UpdateListener(connString string, cash *AppCache) {
	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Println("Listener error:", err)
		}
	}
	listener := pq.NewListener(connString, 10*time.Second, time.Minute, reportProblem)
	if err := listener.Listen("update_cache"); err != nil {
		panic(err)
	}

	fmt.Println("Start monitoring PostgreSQL...")

	go func() {
		for {
			select {
			case <-listener.Notify:
				fmt.Println("PostgreSQL updated")
				fmt.Printf("cache: %+v\n", cash.data)
				_ = cash.Reload()
			case <-time.After(1 * time.Minute):
				go listener.Ping()
			}
		}
	}()
}
