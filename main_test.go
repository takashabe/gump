package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/go-cmp/cmp"
)

const (
	testModA = "a/b"
	testModB = "c"
)

func FakeMultipleTagRepository(t *testing.T) *git.Repository {
	repo, err := git.Init(memory.NewStorage(), memfs.New())
	if err != nil {
		t.Fatal(err)
	}

	mustCreateTag(t, repo, fmt.Sprintf("%s/v0.0.1", testModA))
	mustCreateTag(t, repo, fmt.Sprintf("%s/v0.0.2", testModA))
	mustCreateTag(t, repo, fmt.Sprintf("%s/v0.1.0", testModA))
	mustCreateTag(t, repo, fmt.Sprintf("%s/v1.0.0", testModB))

	// create first commit
	wt, _ := repo.Worktree()
	wt.Filesystem.Create("fake")
	wt.Add("fake")
	if _, err := wt.Commit("", &git.CommitOptions{}); err != nil {
		t.Fatal(err)
	}

	return repo
}

func mustCreateTag(t *testing.T, repo *git.Repository, tag string) {
	t.Helper()
	h := plumbing.NewHash("b8e471f58bcbca63b07bda20e428190409c2db47")
	if _, err := repo.CreateTag(tag, h, nil); err != nil {
		t.Fatal(err)
	}
}

func Test_getModVersion(t *testing.T) {
	tests := []struct {
		name      string
		modPrefix string
		want      *semver.Version
	}{
		{
			name:      "should zero version",
			modPrefix: "",
			want:      semver.MustParse("v0.0.0"),
		},
		{
			name:      "should returns latest modA tag",
			modPrefix: testModA,
			want:      semver.MustParse("v0.1.0"),
		},
		{
			name:      "should returns latest modB tag",
			modPrefix: testModB,
			want:      semver.MustParse("v1.0.0"),
		},
		{
			name:      "should zero version when new-mod",
			modPrefix: "foo",
			want:      semver.MustParse("v0.0.0"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := FakeMultipleTagRepository(t)
			got, err := getModVersion(repo, tt.modPrefix)
			if err != nil {
				t.Fatalf("want no error: got %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_bumpPush(t *testing.T) {
	tests := []struct {
		name      string
		current   *semver.Version
		bump      bumpVersion
		modPrefix string
		want      string
	}{
		{
			name:      "bump patch",
			current:   semver.MustParse("0.0.0"),
			bump:      bumpVersion{false, false, true},
			modPrefix: "a/b/c",
			want:      "a/b/c/v0.0.1",
		},
		{
			name:      "bump minor",
			current:   semver.MustParse("0.1.1"),
			bump:      bumpVersion{false, true, false},
			modPrefix: "a",
			want:      "a/v0.2.0",
		},
		{
			name:      "bump major",
			current:   semver.MustParse("1.1.0"),
			bump:      bumpVersion{true, true, true},
			modPrefix: "a",
			want:      "a/v2.0.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := FakeMultipleTagRepository(t)
			if err := bumpPush(repo, tt.current, tt.bump, tt.modPrefix, false); err != nil {
				t.Fatalf("want no error: got %v", err)
			}

			got, err := repo.Tag(tt.want)
			if err != nil {
				t.Fatalf("want no error: got %v", err)
			}
			gotCleanTag := strings.TrimPrefix(got.Name().String(), "refs/tags/")
			if diff := cmp.Diff(tt.want, gotCleanTag); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
