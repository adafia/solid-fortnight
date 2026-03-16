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

type Environment struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	Key       string `json:"key"`
	SortOrder int    `json:"sort_order"`
	CreatedAt string `json:"created_at"`
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

// CreateEnvironment creates a new environment in the database.
func (s *ProjectStore) CreateEnvironment(env *Environment) error {
	query := `
		INSERT INTO environments (project_id, name, key, sort_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`
	err := s.db.QueryRow(
		query,
		env.ProjectID,
		env.Name,
		env.Key,
		env.SortOrder,
	).Scan(&env.ID, &env.CreatedAt)
	return err
}

// GetEnvironments retrieves all environments for a project.
func (s *ProjectStore) GetEnvironments(projectID string) ([]Environment, error) {
	query := `
		SELECT id, project_id, name, key, sort_order, created_at
		FROM environments
		WHERE project_id = $1
		ORDER BY sort_order ASC`
	
	rows, err := s.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var environments []Environment
	for rows.Next() {
		var env Environment
		err := rows.Scan(
			&env.ID,
			&env.ProjectID,
			&env.Name,
			&env.Key,
			&env.SortOrder,
			&env.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		environments = append(environments, env)
	}
	return environments, nil
}
