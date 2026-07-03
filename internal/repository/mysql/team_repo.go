package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/model"
	"github.com/boskuv/task-manager/internal/repository"
)

var _ repository.TeamRepository = (*TeamRepo)(nil)

const teamSelectColumns = `t.id, t.name, t.created_by, t.created_at`

// TeamRepo implements repository.TeamRepository with MySQL.
type TeamRepo struct {
	db *sql.DB
}

// NewTeamRepo returns a MySQL-backed team repository.
func NewTeamRepo(db *sql.DB) *TeamRepo {
	return &TeamRepo{db: db}
}

// Create inserts a new team and returns the persisted entity.
func (r *TeamRepo) Create(ctx context.Context, team domain.Team) (domain.Team, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO teams (name, created_by) VALUES (?, ?)`,
		team.Name,
		team.CreatedBy,
	)
	if err != nil {
		return domain.Team{}, fmt.Errorf("insert team: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.Team{}, fmt.Errorf("last insert id: %w", err)
	}

	return r.GetByID(ctx, id)
}

// ListByUserID returns teams where the user is a member.
func (r *TeamRepo) ListByUserID(ctx context.Context, userID int64) ([]domain.Team, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+teamSelectColumns+`
		 FROM teams t
		 INNER JOIN team_members tm ON tm.team_id = t.id
		 WHERE tm.user_id = ?
		 ORDER BY t.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list teams by user: %w", err)
	}
	defer rows.Close()

	teams := make([]domain.Team, 0)
	for rows.Next() {
		team, err := scanTeam(rows)
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate teams: %w", err)
	}

	return teams, nil
}

// GetByID returns a team by id.
func (r *TeamRepo) GetByID(ctx context.Context, id int64) (domain.Team, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, created_by, created_at FROM teams WHERE id = ?`,
		id,
	)
	return scanTeam(row)
}

// AddMember adds a user to a team with the given role.
func (r *TeamRepo) AddMember(ctx context.Context, member domain.TeamMember) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO team_members (team_id, user_id, role) VALUES (?, ?, ?)`,
		member.TeamID,
		member.UserID,
		string(member.Role),
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("insert team member: %w", err)
	}
	return nil
}

// GetMemberRole returns a user's role in a team.
func (r *TeamRepo) GetMemberRole(ctx context.Context, teamID, userID int64) (domain.TeamRole, error) {
	var role string
	err := r.db.QueryRowContext(ctx,
		`SELECT role FROM team_members WHERE team_id = ? AND user_id = ?`,
		teamID,
		userID,
	).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrNotFound
		}
		return "", fmt.Errorf("get member role: %w", err)
	}
	return domain.TeamRole(role), nil
}

func scanTeam(row rowScanner) (domain.Team, error) {
	var m model.Team
	if err := row.Scan(&m.ID, &m.Name, &m.CreatedBy, &m.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Team{}, domain.ErrNotFound
		}
		return domain.Team{}, fmt.Errorf("scan team: %w", err)
	}
	return m.ToDomain(), nil
}
