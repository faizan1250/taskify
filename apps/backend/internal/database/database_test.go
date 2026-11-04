package database_test

import (
	"math"
	"testing"
	"time"

	"github.com/faizan1250/taskify/internal/config"
	"github.com/faizan1250/taskify/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPoolConfig_AppliesConnectionPoolSettings(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "postgres",
			Password:        "secret",
			Name:            "tasker",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    10,
			ConnMaxLifetime: 300,
			ConnMaxIdleTime: 60,
		},
	}

	poolCfg, err := database.NewPoolConfig(cfg, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, poolCfg)

	assert.Equal(t, int32(25), poolCfg.MaxConns)
	assert.Equal(t, int32(10), poolCfg.MinConns)
	assert.Equal(t, 300*time.Second, poolCfg.MaxConnLifetime)
	assert.Equal(t, 60*time.Second, poolCfg.MaxConnIdleTime)
}

func TestNewPoolConfig_PreservesDefaultsWhenUnsetOrZero(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "postgres",
			Password:        "secret",
			Name:            "tasker",
			SSLMode:         "disable",
			MaxOpenConns:    0,
			MaxIdleConns:    0,
			ConnMaxLifetime: 0,
			ConnMaxIdleTime: 0,
		},
	}

	poolCfg, err := database.NewPoolConfig(cfg, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, poolCfg)

	defaultPoolCfg, err := pgxpool.ParseConfig("postgres://postgres:secret@localhost:5432/tasker?sslmode=disable")
	require.NoError(t, err)

	assert.Equal(t, defaultPoolCfg.MaxConns, poolCfg.MaxConns)
	assert.Equal(t, defaultPoolCfg.MinConns, poolCfg.MinConns)
	assert.Equal(t, defaultPoolCfg.MaxConnLifetime, poolCfg.MaxConnLifetime)
	assert.Equal(t, defaultPoolCfg.MaxConnIdleTime, poolCfg.MaxConnIdleTime)
}

func TestNewPoolConfig_LocalEnvTracer(t *testing.T) {
	t.Parallel()

	logger := zerolog.Nop()
	cfg := &config.Config{
		Primary: config.Primary{
			Env: "local",
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "secret",
			Name:     "tasker",
			SSLMode:  "disable",
		},
	}

	poolCfg, err := database.NewPoolConfig(cfg, &logger, nil)
	require.NoError(t, err)
	require.NotNil(t, poolCfg)
	assert.NotNil(t, poolCfg.ConnConfig.Tracer)
}

func TestNewPoolConfig_InvalidDSN(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     9999999,
			User:     "postgres",
			Password: "secret",
			Name:     "tasker",
			SSLMode:  "disable",
		},
	}

	poolCfg, err := database.NewPoolConfig(cfg, nil, nil)
	assert.Error(t, err)
	assert.Nil(t, poolCfg)
}

func TestNewPoolConfig_IgnoresOverflowValues(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:         "localhost",
			Port:         5432,
			User:         "postgres",
			Password:     "secret",
			Name:         "tasker",
			SSLMode:      "disable",
			MaxOpenConns: math.MaxInt64,
			MaxIdleConns: math.MaxInt64,
		},
	}

	poolCfg, err := database.NewPoolConfig(cfg, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, poolCfg)

	defaultPoolCfg, err := pgxpool.ParseConfig("postgres://postgres:secret@localhost:5432/tasker?sslmode=disable")
	require.NoError(t, err)

	assert.Equal(t, defaultPoolCfg.MaxConns, poolCfg.MaxConns)
	assert.Equal(t, defaultPoolCfg.MinConns, poolCfg.MinConns)
}
