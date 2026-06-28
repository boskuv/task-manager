package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/repository"
)

var _ repository.AnalyticsRepository = (*AnalyticsRepo)(nil)

const doneTasks7dWindow = `DATE_SUB(UTC_TIMESTAMP(), INTERVAL 7 DAY)`

// AnalyticsRepo implements repository.AnalyticsRepository with MySQL.
type AnalyticsRepo struct {
	db *sql.DB
}

// NewAnalyticsRepo returns a MySQL-backed analytics repository.
func NewAnalyticsRepo(db *sql.DB) *AnalyticsRepo {
	return &AnalyticsRepo{db: db}
}

// ListTeamStats returns teams with member counts and done-task totals for the last 7 days.
func (r *AnalyticsRepo) ListTeamStats(ctx context.Context) ([]domain.TeamStats, error) {
	summaries, err := r.listTeamSummaries(ctx)
	if err != nil {
		return nil, err
	}

	memberStats, err := r.listMemberDoneStats(ctx)
	if err != nil {
		return nil, err
	}

	membersByTeam := make(map[int64][]domain.TeamMemberDoneStats, len(summaries))
	for _, member := range memberStats {
		membersByTeam[member.TeamID] = append(membersByTeam[member.TeamID], member.Stats)
	}

	stats := make([]domain.TeamStats, 0, len(summaries))
	for _, summary := range summaries {
		members := membersByTeam[summary.TeamID]
		if members == nil {
			members = []domain.TeamMemberDoneStats{}
		}
		stats = append(stats, domain.TeamStats{
			TeamID:      summary.TeamID,
			Name:        summary.Name,
			MemberCount: summary.MemberCount,
			DoneTasks7d: summary.DoneTasks7d,
			Members:     members,
		})
	}

	return stats, nil
}

type teamSummaryRow struct {
	TeamID      int64
	Name        string
	MemberCount int
	DoneTasks7d int
}

func (r *AnalyticsRepo) listTeamSummaries(ctx context.Context) ([]teamSummaryRow, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT
			t.id,
			t.name,
			COUNT(DISTINCT tm.user_id) AS member_count,
			COUNT(DISTINCT CASE
				WHEN tk.status = 'done'
				 AND tk.updated_at >= `+doneTasks7dWindow+`
				THEN tk.id
			END) AS done_tasks_7d
		FROM teams t
		LEFT JOIN team_members tm ON tm.team_id = t.id
		LEFT JOIN tasks tk ON tk.team_id = t.id
		GROUP BY t.id, t.name
		ORDER BY t.id`,
	)
	if err != nil {
		return nil, fmt.Errorf("list team summaries: %w", err)
	}
	defer rows.Close()

	summaries := make([]teamSummaryRow, 0)
	for rows.Next() {
		var row teamSummaryRow
		if err := rows.Scan(&row.TeamID, &row.Name, &row.MemberCount, &row.DoneTasks7d); err != nil {
			return nil, fmt.Errorf("scan team summary: %w", err)
		}
		summaries = append(summaries, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate team summaries: %w", err)
	}

	return summaries, nil
}

type memberDoneStatsRow struct {
	TeamID int64
	Stats  domain.TeamMemberDoneStats
}

func (r *AnalyticsRepo) listMemberDoneStats(ctx context.Context) ([]memberDoneStatsRow, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT
			tm.team_id,
			u.id,
			u.email,
			tm.role,
			COUNT(tk.id) AS done_tasks_7d
		FROM team_members tm
		INNER JOIN users u ON u.id = tm.user_id
		LEFT JOIN tasks tk ON tk.team_id = tm.team_id
			AND tk.assignee_id = tm.user_id
			AND tk.status = 'done'
			AND tk.updated_at >= `+doneTasks7dWindow+`
		GROUP BY tm.team_id, u.id, u.email, tm.role
		ORDER BY tm.team_id, done_tasks_7d DESC, u.email`,
	)
	if err != nil {
		return nil, fmt.Errorf("list member done stats: %w", err)
	}
	defer rows.Close()

	rowsOut := make([]memberDoneStatsRow, 0)
	for rows.Next() {
		var row memberDoneStatsRow
		var role string
		if err := rows.Scan(
			&row.TeamID,
			&row.Stats.UserID,
			&row.Stats.Email,
			&role,
			&row.Stats.DoneTasks7d,
		); err != nil {
			return nil, fmt.Errorf("scan member done stats: %w", err)
		}
		row.Stats.Role = domain.TeamRole(role)
		rowsOut = append(rowsOut, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate member done stats: %w", err)
	}

	return rowsOut, nil
}
