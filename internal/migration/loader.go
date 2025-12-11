package migration

import (
	"errors"
	"io"
	"io/fs"
	"log/slog"

	"github.com/golang-migrate/migrate/v4/source"
)

// LoadAll reads all migrations from the provided source.Driver in ascending version order.
// Caller is responsible for closing the source driver.
func LoadAll(src source.Driver, logger *slog.Logger) ([]Migration, error) {
	if logger == nil {
		logger = slog.Default()
	}

	first, err := src.First()
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			logger.Debug("no migrations found", slog.String("reason", "source empty"))
			return []Migration{}, nil
		}
		logger.Error("read first migration", slog.Any("err", err))
		return nil, err
	}
	logger.Info("loading migrations", slog.Uint64("start_version", uint64(first)))

	version := first
	migrations := make([]Migration, 0)
	seen := make(map[uint]struct{})

	for {
		up, ident, err := src.ReadUp(version)
		if err != nil {
			logger.Error("read up migration", slog.Uint64("version", uint64(version)), slog.Any("err", err))
			return nil, err
		}
		upBytes, err := io.ReadAll(up)
		up.Close()
		if err != nil {
			logger.Error("read up content", slog.Uint64("version", uint64(version)), slog.Any("err", err))
			return nil, err
		}

		down, _, err := src.ReadDown(version)
		if err != nil {
			logger.Error("read down migration", slog.Uint64("version", uint64(version)), slog.Any("err", err))
			return nil, err
		}
		downBytes, err := io.ReadAll(down)
		down.Close()
		if err != nil {
			logger.Error("read down content", slog.Uint64("version", uint64(version)), slog.Any("err", err))
			return nil, err
		}

		if _, exists := seen[version]; exists {
			logger.Warn("duplicate migration version", slog.Uint64("version", uint64(version)), slog.String("identifier", ident))
		}
		seen[version] = struct{}{}

		logger.Debug("loaded migration", slog.Uint64("version", uint64(version)), slog.String("identifier", ident))
		migrations = append(migrations, Migration{
			Version:    version,
			Identifier: ident,
			Up:         upBytes,
			Down:       downBytes,
		})

		next, err := src.Next(version)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				break
			}
			logger.Error("find next migration", slog.Uint64("version", uint64(version)), slog.Any("err", err))
			return nil, err
		}
		version = next
	}

	logger.Info("loaded migrations complete", slog.Int("count", len(migrations)))
	return migrations, nil
}
