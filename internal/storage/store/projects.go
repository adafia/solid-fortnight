package store

import (
	"database/sql"
)

type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type ProjectStore struct {
	db *sql.DB
}

func NewProjectStore(db *sql.DB) *ProjectStore {
	return &ProjectStore{db: db}
}

// CreateProject creates a new project in the database.
func (s *ProjectStore) CreateProject(project *Project) error {
	query := `
		INSERT INTO projects (name, description)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at`
	err := s.db.QueryRow(
		query,
		project.Name,
		project.Description,
	).Scan(&project.ID, &project.CreatedAt, &project.UpdatedAt)
	return err
}

// GetProject retrieves a project from the database by its ID.
func (s *ProjectStore) GetProject(id string) (*Project, error) {
	project := &Project{}
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM projects
		WHERE id = $1`
	err := s.db.QueryRow(query, id).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return project, nil
}
