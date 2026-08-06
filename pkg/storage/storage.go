package storage

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/iavianm/raindrop_go_bot/pkg/raindrop"
	"go.etcd.io/bbolt"
)

const tokensBucket = "tokens"
const statesBucket = "states"

type Storage struct {
	db *bbolt.DB
}

type StoredTokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func New(path string) (*Storage, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(tokensBucket)); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(statesBucket)); err != nil {
			return err
		}
		return err
	})

	if err != nil {
		db.Close()
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) SaveTokens(chatID int64, t *raindrop.Tokens) error {
	stored := StoredTokens{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(t.ExpiresIn) * time.Second),
	}

	value, err := json.Marshal(stored)
	if err != nil {
		return fmt.Errorf("storage: marshal tokens: %w", err)
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(tokensBucket)).Put(chatIDKey(chatID), value)
	})
}

func (s *Storage) GetTokens(chatID int64) (*StoredTokens, error) {
	var stored *StoredTokens

	err := s.db.View(func(tx *bbolt.Tx) error {
		value := tx.Bucket([]byte(tokensBucket)).Get(chatIDKey(chatID))
		if value == nil {
			return nil
		}

		stored = new(StoredTokens)
		return json.Unmarshal(value, stored)
	})
	if err != nil {
		return nil, fmt.Errorf("storage: get tokens: %w", err)
	}

	return stored, nil
}

func chatIDKey(chatID int64) []byte {
	return []byte(strconv.FormatInt(chatID, 10))
}

func (s *Storage) SaveState(state string, chatID int64) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(statesBucket)).Put([]byte(state), chatIDKey(chatID))
	})
}

func (s *Storage) PopState(state string) (int64, error) {
	var chatID int64

	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(statesBucket))
		key := []byte(state)

		value := b.Get(key)
		if value == nil {
			return nil
		}

		parsed, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return err
		}
		chatID = parsed

		return b.Delete(key)
	})
	if err != nil {
		return 0, fmt.Errorf("storage: pop state: %w", err)
	}

	return chatID, nil
}

func (s *Storage) DeleteTokens(chatID int64) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(tokensBucket)).Delete(chatIDKey(chatID))
	})
}
