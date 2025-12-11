package migration

import (
	"errors"
	"io"
	"io/fs"

	"github.com/golang-migrate/migrate/v4/source"
)

// LoadAll reads all migrations from the provided source.Driver in ascending version order.
// Caller is responsible for closing the source driver.
func LoadAll(src source.Driver) ([]Migration, error) {
	first, err := src.First()
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []Migration{}, nil
		}
		return nil, err
	}
	version := first
	migrations := make([]Migration, 0)

	for {
		up, ident, err := src.ReadUp(version)
		if err != nil {
			return nil, err
		}
		upBytes, err := io.ReadAll(up)
		up.Close()
		if err != nil {
			return nil, err
		}

		down, _, err := src.ReadDown(version)
		if err != nil {
			return nil, err
		}
		downBytes, err := io.ReadAll(down)
		down.Close()
		if err != nil {
			return nil, err
		}

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
			return nil, err
		}
		version = next
	}

	return migrations, nil
}
