package mysql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mono83/cache"
	"hash/crc32"
	"time"
)

// DatabasePrepare is alias func for sql.db.Prepare
type DatabasePrepare func(string) (*sql.Stmt, error)

func newNode(ttl int, key string, value interface{}) (*mysqlNode, error) {
	n := new(mysqlNode)
	n.createdAt = time.Now().Unix()
	n.expiryAt = n.createdAt + int64(ttl)
	n.key = key

	// Serializing into JSON
	body, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	n.value = body

	return n, nil
}

type mysqlNode struct {
	createdAt int64
	expiryAt  int64
	key       string
	value     []byte
}

func (n mysqlNode) hash() uint32 {
	return crc32.ChecksumIEEE([]byte(n.key))
}

func (n mysqlNode) unmarshall(target interface{}) error {
	return json.Unmarshal(n.value, target)
}

// Cacher implements caching using MySQL
type cacher struct {
	prepare DatabasePrepare
	ttl     int

	table string

	selectStmt string
	insertStmt string
}

// New creates and returns new MySQL cache built over sql.DB.Prepare func
func New(db DatabasePrepare, table string, ttl int) (cache.Interface, error) {
	if db == nil {
		return nil, errors.New("Empty database provided")
	}
	if ttl < 1 || ttl > 2678400 {
		return nil, fmt.Errorf("Invalid TTL %d. It ,ust be in range [1, 2678400]", ttl)
	}

	m := new(cacher)
	m.table = table
	m.prepare = db
	m.ttl = ttl
	m.selectStmt = fmt.Sprintf(
		"SELECT `createdAt`, `expiryAt`, `value` FROM `%s` WHERE `keyHash` = ? AND `key` = ? ORDER BY `createdAt` DESC LIMIT 1",
		table,
	)
	m.insertStmt = fmt.Sprintf(
		"INSERT INTO `%s`(`createdAt`, `expiryAt`, `key`, `keyHash`, `value`) VALUES (?, ?, ?, ?, ?)",
		table,
	)

	return m, nil
}

// Get method reads data from MySQL cache
func (m *cacher) Get(key string, value interface{}) error {
	if key == "" {
		return errors.New("Empty key")
	}

	node := mysqlNode{key: key}
	stmt, err := m.prepare(m.selectStmt)
	if err != nil {
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(node.hash(), node.key)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return cache.NewErrCacheMiss(key)
	}

	err = rows.Scan(&node.createdAt, &node.expiryAt, &node.value)
	if err != nil {
		return err
	}

	// Checking expiry time
	if node.expiryAt < time.Now().Unix() {
		// Expired
		return cache.NewErrCacheMiss(key)
	}

	err = node.unmarshall(value)
	if err != nil {
		err = fmt.Errorf("MySQL Cacher error: %s", err.Error())
	}

	return err
}

// Put method stores data into MySQL cache
func (m *cacher) Put(key string, value interface{}) error {
	stmt, err := m.prepare(m.insertStmt)
	if err != nil {
		return err
	}
	defer stmt.Close()

	node, err := newNode(m.ttl, key, value)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(node.createdAt, node.expiryAt, node.key, node.hash(), node.value)
	return err
}
