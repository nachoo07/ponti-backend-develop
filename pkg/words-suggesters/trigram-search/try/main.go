package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	pkggorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	pgksuggester "github.com/alphacodinggroup/ponti-backend/pkg/words-suggesters/trigram-search"
)

func main() {
	// Forzar Postgres; usar credenciales por ENV o defaults
	ensureEnvDefault("DB_TYPE", "postgres")
	ensureEnvDefault("DB_HOST", "127.0.0.1")
	ensureEnvDefault("DB_PORT", "5432")
	ensureEnvDefault("DB_USER", "postgres")
	ensureEnvDefault("DB_PASSWORD", "postgres")
	ensureEnvDefault("SSL_MODE", "disable")

	// Crear DB temporal
	tempDB := fmt.Sprintf("trigram_demo_%d", time.Now().UnixNano())
	repo, err := pkggorm.Bootstrap("", "", "", "", tempDB, "", 0)
	if err != nil {
		log.Fatalf("failed to initialize repository (Postgres?): %v", err)
	}
	defer dropTempDB(tempDB)

	// Parámetros del suggester + defaults amigables
	table := strings.TrimSpace(os.Getenv("SUGGESTER_TABLE"))
	column := strings.TrimSpace(os.Getenv("SUGGESTER_COLUMN"))
	if table == "" {
		table = "words_demo"
	}
	if column == "" {
		column = "name"
	}
	prefix := strings.TrimSpace(os.Getenv("SUGGESTER_PREFIX"))
	if prefix == "" {
		prefix = "pa"
	}
	limit := envInt("SUGGESTER_LIMIT", 10)
	offset := envInt("SUGGESTER_OFFSET", 0)
	modeStr := strings.ToLower(strings.TrimSpace(os.Getenv("SUGGESTER_MODE")))
	schema := strings.TrimSpace(os.Getenv("SUGGESTER_SCHEMA"))
	if schema == "" {
		schema = "public"
	}
	idcol := strings.TrimSpace(os.Getenv("SUGGESTER_IDCOL"))
	if idcol == "" {
		idcol = "id"
	}
	threshold := envFloat("SUGGESTER_THRESHOLD", 0.3)
	qTimeout := time.Duration(envInt("SUGGESTER_TIMEOUT_MS", 3000)) * time.Millisecond
	ensureExt := true

	// Crear adapter y datos de prueba (extensiones + tabla + índice + inserts)
	adapter := pgksuggester.NewPkggormAdapter(repo)
	if err := bootstrapDemo(context.Background(), repo, schema, table, column, ensureExt); err != nil {
		log.Fatalf("bootstrap demo failed: %v", err)
	}

	// Inicializar WordsSuggester
	suggester, err := pgksuggester.Bootstrap(
		pgksuggester.WithDB(adapter),
		pgksuggester.WithSchema(schema),
		pgksuggester.WithIDColumn(idcol),
		selectMode(modeStr),
		pgksuggester.WithThreshold(threshold),
		pgksuggester.WithQueryTimeout(qTimeout),
		pgksuggester.WithEnsureExtensions(false),
	)
	if err != nil {
		log.Fatalf("bootstrap error: %v", err)
	}
	defer func() { _ = suggester.Close() }()

	// Ejecutar sugerencias con paginado
	ctx, cancel := context.WithTimeout(context.Background(), qTimeout)
	defer cancel()
	results, total, err := suggester.Suggest(ctx, table, column, prefix, limit, offset)
	if err != nil {
		log.Fatalf("suggest error: %v", err)
	}

	fmt.Printf("DB temporal: %s\n", tempDB)
	fmt.Printf("Total resultados posibles: %d\n", total)
	for _, r := range results {
		fmt.Printf("ID:%d Text:%s\n", r.ID, r.Text)
	}
}

func selectMode(s string) pgksuggester.Option {
	switch s {
	case "prefix":
		return pgksuggester.WithMode(pgksuggester.ModePrefix)
	case "trigram":
		return pgksuggester.WithMode(pgksuggester.ModeTrigram)
	default:
		return pgksuggester.WithMode(pgksuggester.ModeHybrid)
	}
}

func envInt(k string, def int) int {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envFloat(k string, def float64) float64 {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func ensureEnvDefault(key, def string) {
	if strings.TrimSpace(os.Getenv(key)) == "" {
		_ = os.Setenv(key, def)
	}
}

func bootstrapDemo(ctx context.Context, repo *pkggorm.Repository, schema, table, column string, ensureExt bool) error {
	if ensureExt {
		if err := repo.Client().WithContext(ctx).Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm").Error; err != nil {
			log.Printf("warn: cannot create extension pg_trgm: %v", err)
		}
		if err := repo.Client().WithContext(ctx).Exec("CREATE EXTENSION IF NOT EXISTS unaccent").Error; err != nil {
			log.Printf("warn: cannot create extension unaccent: %v", err)
		}
	}

	fqtn := fmt.Sprintf("%s.%s", schema, table)
	createTable := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            id BIGSERIAL PRIMARY KEY,
            %s TEXT NOT NULL
        )`, fqtn, column)
	if err := repo.Client().WithContext(ctx).Exec(createTable).Error; err != nil {
		return fmt.Errorf("create table: %w", err)
	}
	createIdx := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_%s_trgm ON %s USING gin (unaccent(lower(%s)) gin_trgm_ops)`, table, column, fqtn, column)
	if err := repo.Client().WithContext(ctx).Exec(createIdx).Error; err != nil {
		log.Printf("warn: cannot create trigram index: %v", err)
	}

	samples := []string{"Pablo", "Paula", "Pedro", "Patricia", "Paz", "Pablo Gomez", "Pablito", "Pampa", "Pantera", "Paloma"}
	for _, name := range samples {
		stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (?)", fqtn, column)
		if err := repo.Client().WithContext(ctx).Exec(stmt, name).Error; err != nil {
			log.Printf("warn: insert sample '%s': %v", name, err)
		}
	}
	return nil
}

// dropTempDB elimina la base temporal (best-effort)
func dropTempDB(dbName string) {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	ssl := os.Getenv("SSL_MODE")
	port, _ := strconv.Atoi(os.Getenv("DB_PORT"))
	adminRepo, err := pkggorm.Bootstrap("postgres", host, user, pass, "postgres", ssl, port)
	if err != nil {
		return
	}
	_ = adminRepo.Client().Exec("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = ?", dbName).Error
	_ = adminRepo.Client().Exec(fmt.Sprintf("DROP DATABASE IF EXISTS \"%s\"", dbName)).Error
}
