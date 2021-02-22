package main

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/archdx/gpool"

	"github.com/davecgh/go-spew/spew"
)

func main() {
	pool := gpool.NewPool(64)
	defer pool.Close()

	var mu sync.Mutex
	var users []User

	idCh := make(chan int)

	gr := pool.GGroup(8)

	gr.Exec(func() {
		for id := range idCh {
			if user, ok := storage.queryUser(id); ok {
				mu.Lock()
				users = append(users, user)
				mu.Unlock()
			}
		}
	})

	for _, id := range storage.getRandomIDs() {
		idCh <- id
	}

	close(idCh)
	gr.WaitAndRelease()

	spew.Dump(users)
}

var storage Storage

type User struct {
	ID   int
	Name string
}

type Storage struct {
	mu       sync.RWMutex
	usersMap map[int]User
}

func (s *Storage) queryUser(id int) (User, bool) {
	s.mu.RLock()
	user, ok := s.usersMap[id]
	s.mu.RUnlock()

	return user, ok
}

func (s *Storage) getRandomIDs() (ids []int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for id := range s.usersMap {
		if rand.Intn(len(s.usersMap)/10) == 0 {
			ids = append(ids, id)
		}
	}

	return
}

func (s *Storage) init(usersCount int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.usersMap = make(map[int]User, usersCount)

	for id := 1; id <= usersCount; id++ {
		s.usersMap[id] = User{
			ID:   id,
			Name: fmt.Sprintf("User %d", id),
		}
	}
}

func init() {
	const usersCount = 256
	storage.init(usersCount)
}
