package util

import (
	"errors"
	"fmt"
)

func MakePath(path, child string) string {
	return fmt.Sprintf("%s/%s", path, child)
}

func ValidatePath(path string) error {
	if path == "" {
		return errors.New("Path cannot be null")
	}

	if path[0] != '/' {
		return errors.New("Path must start with / character")
	}

	if len(path) == 1 {
		return nil
	}

	if path[len(path)-1] == '/' {
		return errors.New("Path must not end with / character")
	}

	for i, nextCharacter := 1, '/'; i < len(path); i++ {
		if path[i] == 0 {
			return errors.New(fmt.Sprintf("null character not allowed @%d", i))
		}

		if path[i] == '/' && nextCharacter == '/' {
			return errors.New(fmt.Sprintf("empty node name specified @%d", i))
		}

		// /../foo or /..
		if path[i] == '.' && nextCharacter == '.' {
			if path[i-2] == '/' && (i+1 == len(path) || path[i+1] == '/') {
				return errors.New(fmt.Sprintf("relative paths not allowed @%d", i))
			}
		}

		nextCharacter = int32(path[i])
	}

	return nil
}
