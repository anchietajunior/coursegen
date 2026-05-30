// Package assets embeds the CourseGen planning skills into the binary, so a
// `brew install`ed coursegen carries them and `coursegen setup` can install
// them into the user's agent without any extra download.
//
// The tree under skills/ is GENERATED from the repo-root skills/*.md (source of
// truth) by `make sync-skills`. Do not edit it by hand.
package assets

import (
	"embed"
	"io/fs"
)

//go:embed skills
var embedded embed.FS

// SkillsFS returns the embedded skills tree rooted at "skills".
func SkillsFS() fs.FS { return embedded }
