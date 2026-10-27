package consumer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Version struct {
	Full      string
	Version   string
	Build     int
	Interface int
}

var versionPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]?)\.(0|[1-9][0-9]?)\.([1-9][0-9]*)(\n)?$`)

func ReadVersion(root string) (Version, error) {
	path := filepath.Join(root, ".wow-version")
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 64 {
		return Version{}, fmt.Errorf("missing or invalid .wow-version; expected major.minor.patch.build")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Version{}, err
	}
	match := versionPattern.FindStringSubmatch(string(data))
	if match == nil {
		return Version{}, fmt.Errorf("invalid .wow-version; expected one major.minor.patch.build without leading zeros")
	}
	values := make([]int, 4)
	for i := range values {
		values[i], err = strconv.Atoi(match[i+1])
		if err != nil {
			return Version{}, fmt.Errorf("invalid .wow-version component: %w", err)
		}
	}
	return Version{strings.TrimSuffix(string(data), "\n"), strings.Join(match[1:4], "."), values[3], values[0]*10000 + values[1]*100 + values[2]}, nil
}

func (v Version) Require(identity map[string]any) error {
	if identity["version"] != v.Full || intValue(identity["build"]) != v.Build {
		return fmt.Errorf(".wow-version selects %s; source cohort is %v. Stage and review the selected release first", v.Full, identity["version"])
	}
	return nil
}
