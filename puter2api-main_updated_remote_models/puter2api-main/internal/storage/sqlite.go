package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Token 表示存储的 Puter 认证 Token
type Token struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`      // 用户自定义名称
	Token     string     `json:"token"`     // JWT Token
	IsActive  bool       `json:"is_active"` // 是否启用
	IsValid   bool       `json:"is_valid"`  // 是否有效（测试通过）
	LastUsed  *time.Time `json:"last_used,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Storage 数据库存储接口
type Storage struct {
	db *sql.DB
}

// New 创建新的存储实例
func New(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	s := &Storage{db: db}
	if err := s.init(); err != nil {
		db.Close()
		return nil, err
	}

	return s, nil
}

// init 初始化数据库表
func (s *Storage) init() error {
	query := `
	CREATE TABLE IF NOT EXISTS tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL DEFAULT '',
		token TEXT NOT NULL UNIQUE,
		is_active BOOLEAN NOT NULL DEFAULT 1,
		is_valid BOOLEAN NOT NULL DEFAULT 0,
		last_used DATETIME,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_tokens_is_active ON tokens(is_active);
	CREATE INDEX IF NOT EXISTS idx_tokens_is_valid ON tokens(is_valid);
	`
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	return nil
}

// Close 关闭数据库连接
func (s *Storage) Close() error {
	return s.db.Close()
}

// AddToken 添加新 Token
func (s *Storage) AddToken(name, token string) (*Token, error) {
	now := time.Now()
	result, err := s.db.Exec(
		`INSERT INTO tokens (name, token, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		name, token, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add token: %w", err)
	}

	id, _ := result.LastInsertId()
	return &Token{
		ID:        id,
		Name:      name,
		Token:     token,
		IsActive:  true,
		IsValid:   false,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// GetAllTokens 获取所有 Token
func (s *Storage) GetAllTokens() ([]Token, error) {
	rows, err := s.db.Query(
		`SELECT id, name, token, is_active, is_valid, last_used, created_at, updated_at
		 FROM tokens ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query tokens: %w", err)
	}
	defer rows.Close()

	var tokens []Token
	for rows.Next() {
		var t Token
		var lastUsed sql.NullTime
		err := rows.Scan(&t.ID, &t.Name, &t.Token, &t.IsActive, &t.IsValid, &lastUsed, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan token: %w", err)
		}
		if lastUsed.Valid {
			t.LastUsed = &lastUsed.Time
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}

// GetToken 根据 ID 获取 Token
func (s *Storage) GetToken(id int64) (*Token, error) {
	var t Token
	var lastUsed sql.NullTime
	err := s.db.QueryRow(
		`SELECT id, name, token, is_active, is_valid, last_used, created_at, updated_at
		 FROM tokens WHERE id = ?`, id,
	).Scan(&t.ID, &t.Name, &t.Token, &t.IsActive, &t.IsValid, &lastUsed, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	if lastUsed.Valid {
		t.LastUsed = &lastUsed.Time
	}
	return &t, nil
}

// GetActiveToken 获取一个可用的 Token（轮询策略）
func (s *Storage) GetActiveToken() (*Token, error) {
	var t Token
	var lastUsed sql.NullTime
	// 优先选择有效且最久未使用的 Token
	err := s.db.QueryRow(
		`SELECT id, name, token, is_active, is_valid, last_used, created_at, updated_at
		 FROM tokens WHERE is_active = 1 AND is_valid = 1
		 ORDER BY last_used ASC NULLS FIRST, created_at ASC LIMIT 1`,
	).Scan(&t.ID, &t.Name, &t.Token, &t.IsActive, &t.IsValid, &lastUsed, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active token: %w", err)
	}
	if lastUsed.Valid {
		t.LastUsed = &lastUsed.Time
	}
	return &t, nil
}

// UpdateTokenUsed 更新 Token 最后使用时间
func (s *Storage) UpdateTokenUsed(id int64) error {
	now := time.Now()
	_, err := s.db.Exec(
		`UPDATE tokens SET last_used = ?, updated_at = ? WHERE id = ?`,
		now, now, id,
	)
	return err
}

// UpdateTokenValid 更新 Token 有效性
func (s *Storage) UpdateTokenValid(id int64, isValid bool) error {
	now := time.Now()
	_, err := s.db.Exec(
		`UPDATE tokens SET is_valid = ?, updated_at = ? WHERE id = ?`,
		isValid, now, id,
	)
	return err
}

// UpdateTokenActive 更新 Token 启用状态
func (s *Storage) UpdateTokenActive(id int64, isActive bool) error {
	now := time.Now()
	_, err := s.db.Exec(
		`UPDATE tokens SET is_active = ?, updated_at = ? WHERE id = ?`,
		isActive, now, id,
	)
	return err
}

// DeleteToken 删除 Token
func (s *Storage) DeleteToken(id int64) error {
	_, err := s.db.Exec(`DELETE FROM tokens WHERE id = ?`, id)
	return err
}

// UpdateToken 更新 Token 信息
func (s *Storage) UpdateToken(id int64, name, token string) error {
	now := time.Now()
	_, err := s.db.Exec(
		`UPDATE tokens SET name = ?, token = ?, is_valid = 0, updated_at = ? WHERE id = ?`,
		name, token, now, id,
	)
	return err
}
