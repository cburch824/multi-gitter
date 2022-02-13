package multigitter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCloneDirectory_FromGitlab(t *testing.T) {
	repoUrl := "https://gitlab.com/group1/group2/group3/reponame"
	expected := "group1/group2/group3/reponame"

	actualResult, actualErr := getCloneDirectory(repoUrl)

	require.NoError(t, actualErr)
	assert.Equal(t, expected, actualResult)
}

func TestGetCloneDirectory_FromGithub(t *testing.T) {
	repoUrl := "https://github.com/org1/group1/group2/reponame"
	expected := "org1/group1/group2/reponame"

	actualResult, actualErr := getCloneDirectory(repoUrl)

	require.NoError(t, actualErr)
	assert.Equal(t, expected, actualResult)
}

func TestGetCloneDirectory_WithCapitalLetters(t *testing.T) {
	repoUrl := "https://gitlab.com/Group1/gRoup2/grouP3/repoName"
	expected := "Group1/gRoup2/grouP3/repoName"

	actualResult, actualErr := getCloneDirectory(repoUrl)

	require.NoError(t, actualErr)
	assert.Equal(t, expected, actualResult)
}

func TestGetCloneDirectory_WithHyphens(t *testing.T) {
	repoUrl := "https://gitlab.com/g-roup1/grou-p2/group-3/repo-name"
	expected := "g-roup1/grou-p2/group-3/repo-name"

	actualResult, actualErr := getCloneDirectory(repoUrl)

	require.NoError(t, actualErr)
	assert.Equal(t, expected, actualResult)
}
