package database

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/url"
	"strconv"

	"github.com/faizan1250/taskify/internal/config"

	"github.com/jackc/pgx/v5"
	tern "github.com/jackc/tern/v2/migrate"
	"github.com/rs/zerolog"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Migrate(ctx context.Context, logger *zerolog.Logger, cfg *config.Config) error {
	hostPort := net.JoinHostPort(cfg.Database.Host, strconv.Itoa(cfg.Database.Port))

	// URL-encode the password
	encodedPassword := url.QueryEscape(cfg.Database.Password)
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		cfg.Database.User,
		encodedPassword,
		hostPort,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close(ctx)
	}()

	m, err := tern.NewMigrator(ctx, conn, "schema_version")
	if err != nil {
		return fmt.Errorf("constructing database migrator: %w", err)
	}
	subtree, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("retrieving database migrations subtree: %w", err)
	}
	if err := m.LoadMigrations(subtree); err != nil {
		return fmt.Errorf("loading database migrations: %w", err)
	}
	from, err := m.GetCurrentVersion(ctx)
	if err != nil {
		return errors.New("retreiving current database migration version")
	}
	if err := m.Migrate(ctx); err != nil {
		return err
	}
	if int(from) == len(m.Migrations) {
		logger.Info().Msgf("database schema up to date, version %d", len(m.Migrations))
	} else {
		logger.Info().Msgf("migrated database schema, from %d to %d", from, len(m.Migrations))
	}
	return nil
}
