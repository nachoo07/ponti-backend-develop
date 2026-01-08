package pkgsuggester

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type WordsSuggester struct {
	db           DB
	limit        int
	threshold    float64
	logger       Logger
	schema       string
	idColumn     string
	mode         MatchMode
	countMode    CountMode
	queryTimeout time.Duration
}

func newSuggester(cfg *Config) *WordsSuggester {
	// Opcionalmente asegurar extensiones (mejor en migraciones; aquí sólo si se configuro).
	if cfg.EnsureExtensions {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := cfg.DB.WithContext(ctx).Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm").Error(); err != nil {
			cfg.logger.Error("failed to create pg_trgm extension", err)
		}
		if err := cfg.DB.WithContext(ctx).Exec("CREATE EXTENSION IF NOT EXISTS unaccent").Error(); err != nil {
			cfg.logger.Error("failed to create unaccent extension", err)
		}
	}

	return &WordsSuggester{
		db:           cfg.DB,
		limit:        cfg.Limit,
		threshold:    cfg.Threshold,
		logger:       cfg.logger,
		schema:       cfg.Schema,
		idColumn:     cfg.IDColumn,
		mode:         cfg.Mode,
		countMode:    cfg.Count,
		queryTimeout: cfg.QueryTimeout,
	}
}

func (s *WordsSuggester) Suggest(
	ctx context.Context,
	table, column, prefix string,
	limit, offset int,
) ([]Suggestion, int64, error) {
	if !validIdentifier.MatchString(table) {
		return nil, 0, fmt.Errorf("invalid table name: %s", table)
	}
	if !validIdentifier.MatchString(column) {
		return nil, 0, fmt.Errorf("invalid column name: %s", column)
	}
	q := strings.TrimSpace(prefix)
	if q == "" {
		return nil, 0, nil
	}

	fullTable := s.schema + "." + table
	if !validIdentifier.MatchString(s.schema) || !validIdentifier.MatchString(table) {
		return nil, 0, fmt.Errorf("invalid schema/table")
	}
	if !validIdentifier.MatchString(s.idColumn) {
		return nil, 0, fmt.Errorf("invalid id column")
	}

	if limit <= 0 {
		limit = s.limit
	}
	if offset < 0 {
		offset = 0
	}

	if s.queryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.queryTimeout)
		defer cancel()
	}

	// 1. Query paginada según modo
	var sqlStmt string
	var args []any
	switch s.mode {
	case ModePrefix:
		sqlStmt = fmt.Sprintf(
			`SELECT %[1]s AS id, %[2]s AS text
             FROM %s
             WHERE unaccent(lower(%[2]s)) LIKE unaccent(lower(?)) || '%%'
             ORDER BY %[2]s ASC
             LIMIT ? OFFSET ?`,
			s.idColumn, column, fullTable,
		)
		args = []any{q, limit, offset}
	case ModeTrigram:
		sqlStmt = fmt.Sprintf(
			`SELECT %[1]s AS id, %[2]s AS text
             FROM %s
             WHERE unaccent(lower(%[2]s)) %% unaccent(lower(?))
               AND similarity(unaccent(lower(%[2]s)), unaccent(lower(?))) >= ?
             ORDER BY unaccent(lower(%[2]s)) <-> unaccent(lower(?)) ASC
             LIMIT ? OFFSET ?`,
			s.idColumn, column, fullTable,
		)
		args = []any{q, q, s.threshold, q, limit, offset}
	case ModeHybrid:
		sqlStmt = fmt.Sprintf(
			`WITH params AS (SELECT unaccent(lower(?)) AS q),
            pref AS (
              SELECT %[1]s AS id, %[2]s AS text, 0 AS bucket, 0::float AS dist
              FROM %s t, params p
              WHERE unaccent(lower(%[2]s)) LIKE p.q || '%%'
            ),
            fuzzy AS (
              SELECT %[1]s AS id, %[2]s AS text, 1 AS bucket,
                     (unaccent(lower(%[2]s)) <-> p.q) AS dist
              FROM %s t, params p
              WHERE unaccent(lower(%[2]s)) %% p.q
                AND similarity(unaccent(lower(%[2]s)), p.q) >= ?
            )
            SELECT id, text FROM (
              SELECT * FROM pref
              UNION ALL
              SELECT * FROM fuzzy
            ) s
            ORDER BY bucket ASC, dist ASC, text ASC
            LIMIT ? OFFSET ?`,
			s.idColumn, column, fullTable, fullTable,
		)
		args = []any{q, s.threshold, limit, offset}
	default:
		return nil, 0, fmt.Errorf("unknown match mode: %d", s.mode)
	}
	var out []Suggestion
	start := time.Now()
	res := s.db.WithContext(ctx).Raw(sqlStmt, args...).Scan(&out)
	if err := res.Error(); err != nil {
		s.logger.Error("suggest query failed", err)
		return nil, 0, fmt.Errorf("suggest: %w", err)
	}

	total, err := s.count(ctx, fullTable, column, q, limit)
	if err != nil {
		return nil, 0, err
	}

	s.logger.Debug(fmt.Sprintf("latency: %s", time.Since(start)))
	return out, total, nil
}

func (s *WordsSuggester) Close() error { return nil }

func (s *WordsSuggester) Health(ctx context.Context) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// count calcula el total según el CountMode y el MatchMode activos.
func (s *WordsSuggester) count(ctx context.Context, fullTable, column, q string, pageLimit int) (int64, error) {
	switch s.countMode {
	case CountApproxNone:
		return 0, nil
	case CountPrefixThenFuzzyIfNeeded:
		var pref int64
		sqlPref := fmt.Sprintf(
			`SELECT COUNT(*) FROM %s WHERE unaccent(lower(%s)) LIKE unaccent(lower(?)) || '%%'`,
			fullTable, column,
		)
		if err := s.db.WithContext(ctx).Raw(sqlPref, q).Scan(&pref).Error(); err != nil {
			s.logger.Error("count prefix failed", err)
			return 0, fmt.Errorf("count-prefix: %w", err)
		}
		if pref >= int64(pageLimit) || s.mode == ModePrefix {
			return pref, nil
		}
		var fuzzy int64
		sqlFuzzy := fmt.Sprintf(
			`SELECT COUNT(*) FROM %s
             WHERE unaccent(lower(%s)) %% unaccent(lower(?))
               AND similarity(unaccent(lower(%s)), unaccent(lower(?))) >= ?`,
			fullTable, column, column,
		)
		if err := s.db.WithContext(ctx).Raw(sqlFuzzy, q, q, s.threshold).Scan(&fuzzy).Error(); err != nil {
			s.logger.Error("count fuzzy failed", err)
			return 0, fmt.Errorf("count-fuzzy: %w", err)
		}
		return pref + fuzzy, nil
	case CountExact:
		var total int64
		switch s.mode {
		case ModePrefix:
			sqlCount := fmt.Sprintf(
				`SELECT COUNT(*) FROM %s WHERE unaccent(lower(%s)) LIKE unaccent(lower(?)) || '%%'`,
				fullTable, column,
			)
			if err := s.db.WithContext(ctx).Raw(sqlCount, q).Scan(&total).Error(); err != nil {
				s.logger.Error("count prefix exact failed", err)
				return 0, fmt.Errorf("count: %w", err)
			}
		case ModeTrigram:
			sqlCount := fmt.Sprintf(
				`SELECT COUNT(*) FROM %s
                 WHERE unaccent(lower(%s)) %% unaccent(lower(?))
                   AND similarity(unaccent(lower(%s)), unaccent(lower(?))) >= ?`,
				fullTable, column, column,
			)
			if err := s.db.WithContext(ctx).Raw(sqlCount, q, q, s.threshold).Scan(&total).Error(); err != nil {
				s.logger.Error("count trigram exact failed", err)
				return 0, fmt.Errorf("count: %w", err)
			}
		case ModeHybrid:
			sqlCount := fmt.Sprintf(
				`WITH params AS (SELECT unaccent(lower(?)) AS q),
                  pref AS (
                    SELECT 1 FROM %s t, params p
                    WHERE unaccent(lower(%s)) LIKE p.q || '%%'
                  ),
                  fuzzy AS (
                    SELECT 1 FROM %s t, params p
                    WHERE unaccent(lower(%s)) %% p.q
                      AND similarity(unaccent(lower(%s)), p.q) >= ?
                  )
                  SELECT (SELECT COUNT(*) FROM pref) + (SELECT COUNT(*) FROM fuzzy) AS total`,
				fullTable, column, fullTable, column, column,
			)
			if err := s.db.WithContext(ctx).Raw(sqlCount, q, s.threshold).Scan(&total).Error(); err != nil {
				s.logger.Error("count hybrid exact failed", err)
				return 0, fmt.Errorf("count: %w", err)
			}
		}
		return total, nil
	default:
		return 0, fmt.Errorf("unknown count mode: %d", s.countMode)
	}
}
