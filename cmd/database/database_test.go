package database

import (
	"fmt"
	"os/user"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateDatabaseNameWithHomeDirectory(t *testing.T) {
	user, err := user.Current()
	if err != nil {
		require.NoError(t, err)
	}

	// These two paths are identical and need to produce the same database name.
	path1 := "~/code/my_projects/blog"
	path2 := fmt.Sprintf("/Users/%s/code/my_projects/blog", user.Username)

	dbName1, err := generateDatabaseName(path1)
	if err != nil {
		require.NoError(t, err)
	}

	dbName2, err := generateDatabaseName(path2)
	if err != nil {
		require.NoError(t, err)
	}

	require.Equal(t, dbName1, dbName2)
}

func TestGenerateDatabaseNameCaseInsensitive(t *testing.T) {
	user, err := user.Current()
	if err != nil {
		require.NoError(t, err)
	}

	// These two paths are identical and need to produce the same database name.
	path1 := fmt.Sprintf("/Users/%s/Code/my_projects/Blog", user.Username)
	path2 := fmt.Sprintf("/Users/%s/code/my_projects/blog", user.Username)

	dbName1, err := generateDatabaseName(path1)
	if err != nil {
		require.NoError(t, err)
	}

	dbName2, err := generateDatabaseName(path2)
	if err != nil {
		require.NoError(t, err)
	}

	require.Equal(t, dbName1, dbName2)
}

func TestGenerateDatabaseNameSlashes(t *testing.T) {
	user, err := user.Current()
	if err != nil {
		require.NoError(t, err)
	}

	// These two paths are identical and need to produce the same database name.
	path1 := fmt.Sprintf("/Users/%s/code/my_projects/blog/", user.Username)
	path2 := fmt.Sprintf("/Users/%s/code/my_projects/blog", user.Username)

	dbName1, err := generateDatabaseName(path1)
	if err != nil {
		require.NoError(t, err)
	}

	dbName2, err := generateDatabaseName(path2)
	if err != nil {
		require.NoError(t, err)
	}

	require.Equal(t, dbName1, dbName2)
}

func TestGenerateDatabaseNameIncludesLowerCaseKeel(t *testing.T) {
	user, err := user.Current()
	if err != nil {
		require.NoError(t, err)
	}

	path := fmt.Sprintf("/Users/%s/code/my_projects/Blog", user.Username)

	dbName, err := generateDatabaseName(path)
	if err != nil {
		require.NoError(t, err)
	}

	require.True(t, strings.HasPrefix(dbName, "keel_"))
}
