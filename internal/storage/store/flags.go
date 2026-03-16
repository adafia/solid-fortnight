package store

import (
	"database/sql"
)

type Flag struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Tags        []byte `json:"tags"`
	CreatedBy   string `json:"created_by"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	Archived    bool   `json:"archived"`
}

type FlagStore struct {
	db *sql.DB
}

func NewFlagStore(db *sql.DB) *FlagStore {
	return &FlagStore{db: db}
}

// CreateFlag creates a new flag in the database.
func (s *FlagStore) CreateFlag(flag *Flag) error {
	query := `
		INSERT INTO flags (project_id, key, name, description, type, tags, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`
	err := s.db.QueryRow(
		query,
		flag.ProjectID,
		flag.Key,
		flag.Name,
		flag.Description,
		flag.Type,
		flag.Tags,
		flag.CreatedBy,
	).Scan(&flag.ID, &flag.CreatedAt, &flag.UpdatedAt)
	return err
}

// GetFlag retrieves a flag from the database by its ID.
func (s *FlagStore) GetFlag(id string) (*Flag, error) {
	flag := &Flag{}
	query := `
		SELECT id, project_id, key, name, description, type, tags, created_by, created_at, updated_at, archived
		FROM flags
		WHERE id = $1 AND archived = false`
	err := s.db.QueryRow(query, id).Scan(
		&flag.ID,
		&flag.ProjectID,
		&flag.Key,
		&flag.Name,
		&flag.Description,
		&flag.Type,
		&flag.Tags,
		&flag.CreatedBy,
		&flag.CreatedAt,
		&flag.UpdatedAt,
		&flag.Archived,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Or a custom not found error
		}
		return nil, err
	}
	return flag, nil
}

// GetFlagByKey retrieves a flag from the database by its project and key.
func (s *FlagStore) GetFlagByKey(projectID, key string) (*Flag, error) {
	flag := &Flag{}
	query := `
		SELECT id, project_id, key, name, description, type, tags, created_by, created_at, updated_at, archived
		FROM flags
		WHERE project_id = $1 AND key = $2 AND archived = false`
	err := s.db.QueryRow(query, projectID, key).Scan(
		&flag.ID,
		&flag.ProjectID,
		&flag.Key,
		&flag.Name,
		&flag.Description,
		&flag.Type,
		&flag.Tags,
		&flag.CreatedBy,
		&flag.CreatedAt,
		&flag.UpdatedAt,
		&flag.Archived,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return flag, nil
}

// UpdateFlag updates a flag in the database.
func (s *FlagStore) UpdateFlag(flag *Flag) error {
	query := `
		UPDATE flags
		SET name = $1, description = $2, type = $3, tags = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
		RETURNING updated_at`
	err := s.db.QueryRow(
		query,
		flag.Name,
		flag.Description,
		flag.Type,
		flag.Tags,
		flag.ID,
	).Scan(&flag.UpdatedAt)
	return err
}

// DeleteFlag marks a flag as archived in the.
func (s *FlagStore) DeleteFlag(id string) error {
	query := "UPDATE flags SET archived = true, updated_at = CURRENT_TIMESTAMP WHERE id = $1"
	_, err := s.db.Exec(query, id)
	return err
}
