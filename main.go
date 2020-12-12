package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/kyokomi/emoji"
	"github.com/spf13/cobra"
)

func main() {
	bump := cmdBump{}
	cmd := bump.New()

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

type cmdBump struct {
	bumpVersion

	gitDir string
	modDir string
	push   bool
}

type bumpVersion struct {
	major bool
	minor bool
	patch bool
}

func (b *cmdBump) New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gump",
		Short: "bump up git tag version",
		RunE:  b.RunE,
	}

	cmd.Flags().StringVarP(&b.gitDir, "git-dir", "g", ".", "repository root (the .git directory)")
	cmd.Flags().StringVarP(&b.modDir, "gomod-dir", "m", ".", "go module root (the go.mod file)")
	cmd.Flags().BoolVarP(&b.push, "push", "p", false, "push tags")

	cmd.Flags().BoolVarP(&b.major, "major", "", false, "increment major version")
	cmd.Flags().BoolVarP(&b.minor, "minor", "", false, "increment minor version")
	cmd.Flags().BoolVarP(&b.patch, "patch", "", true, "increment patch version")

	return cmd
}

func (b *cmdBump) RunE(cmd *cobra.Command, args []string) error {
	repo, err := git.PlainOpen(filepath.Clean(b.gitDir))
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// go moduleはsemver prefixにrootからのパス名を接頭辞に付ける
	absRootDir, err := filepath.Abs(b.gitDir)
	if err != nil {
		return err
	}
	absModDir, err := filepath.Abs(b.modDir)
	if err != nil {
		return err
	}
	modPrefix, err := filepath.Rel(absRootDir, absModDir)
	if err != nil {
		return err
	}

	latest, err := getModVersion(repo, modPrefix)
	if err != nil {
		return err
	}
	if err := bumpPush(repo, latest, b.bumpVersion, modPrefix, b.push); err != nil {
		return err
	}

	return nil
}

func getModVersion(repo *git.Repository, modPrefix string) (*semver.Version, error) {
	tagRefs, err := repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	// 混乱を防ぐために末尾スラッシュがある状態に統一する
	if !strings.HasSuffix(modPrefix, "/") {
		modPrefix = fmt.Sprintf("%s/", modPrefix)
	}

	var vers semver.Collection
	err = tagRefs.ForEach(func(r *plumbing.Reference) error {
		// "refs/tags/v42.0.0" のフォーマットでtagが取得できる
		name := strings.TrimPrefix(r.Name().String(), "refs/tags/")

		// module対応("tools/ops/bump/v0.0.1")のtagを探索する
		if strings.HasPrefix(name, modPrefix) {
			ver := strings.TrimPrefix(name, modPrefix)
			v, err := semver.NewVersion(ver)
			if err != nil {
				// 該当のmodule対応ではないのでスキップ
				return nil
			}
			vers = append(vers, v)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// tag未作成
	if len(vers) == 0 {
		return semver.NewVersion("0.0.0")
	}

	sort.Sort(vers)
	return vers[len(vers)-1], nil
}

func bumpPush(repo *git.Repository, v *semver.Version, b bumpVersion, modPrefix string, push bool) error {
	var next semver.Version
	if b.major {
		next = v.IncMajor()
	} else if b.minor {
		next = v.IncMinor()
	} else {
		next = v.IncPatch()
	}

	h, err := repo.Head()
	if err != nil {
		return err
	}

	nextTag := modSemver(next, modPrefix)
	_, err = repo.CreateTag(nextTag, h.Hash(), nil)
	if err != nil {
		return err
	}
	emoji.Printf(":check_mark: create tag %s\n", nextTag)

	if !push {
		return nil
	}

	_, err = exec.Command("git", "push", "origin", nextTag).Output()
	if err != nil {
		return err
	}
	emoji.Printf(":check_mark: push tag %s\n", nextTag)

	return nil
}

func modSemver(v semver.Version, modPrefix string) string {
	p := strings.TrimSuffix(modPrefix, "/")
	return fmt.Sprintf("%s/v%s", p, v.String())
}
