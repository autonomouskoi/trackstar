package pb

import (
	"errors"
	"fmt"

	"github.com/autonomouskoi/datastruct/mapset"
)

func (cfg *Config) Validate() error {
	seenTags := mapset.MapSet[string]{}
	for _, tag := range cfg.Tags {
		if tag.Tag == "" {
			return errors.New("tag cannot be empty")
		}
		if seenTags.Has(tag.Tag) {
			return fmt.Errorf("duplicate tag %q", tag.Tag)
		}
		seenTags.Add(tag.Tag)
	}
	return nil
}
