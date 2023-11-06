package ioutil

import (
	"os"
	"os/user"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpandUser(t *testing.T) {
	assert := require.New(t)
	var v string
	var err error

	u, _ := user.Current()

	v, err = ExpandUser(`/dev/null`)
	assert.Equal(v, `/dev/null`)
	assert.Nil(err)

	v, err = ExpandUser(`~`)
	assert.Equal(v, u.HomeDir)
	assert.Nil(err)

	v, err = ExpandUser("~" + u.Name)
	assert.Equal(v, u.HomeDir)
	assert.Nil(err)

	v, err = ExpandUser("~/test-123")
	assert.Equal(v, u.HomeDir+"/test-123")
	assert.Nil(err)

	v, err = ExpandUser("~" + u.Name + "/test-123")
	assert.Equal(v, u.HomeDir+"/test-123")
	assert.Nil(err)

	v, err = ExpandUser("~/test-123/~/123")
	assert.Equal(v, u.HomeDir+"/test-123/~/123")
	assert.Nil(err)

	v, err = ExpandUser("~" + u.Name + "/test-123/~" + u.Name + "/123")
	assert.Equal(v, u.HomeDir+"/test-123/~"+u.Name+"/123")
	assert.Nil(err)

	assert.False(IsNonemptyFile(`/nonexistent.txt`))
	assert.False(IsNonemptyDir(`/nonexistent/dir`))

	assert.True(IsNonemptyFile(`/etc/hosts`))
	assert.True(IsNonemptyDir(`/etc`))

	x, err := os.Executable()
	assert.NoError(err)
	assert.True(IsNonemptyExecutableFile(x))
}

func TestMatchPath(t *testing.T) {
	require.True(t, MatchPath(`**`, `/hello/there.txt`))
	require.True(t, MatchPath(`*.txt`, `/hello/there.txt`))
	require.True(t, MatchPath(`**/*.txt`, `/hello/there.txt`))
	require.False(t, MatchPath(`**/*.txt`, `/hello/there.jpg`))
	require.True(t, MatchPath("* ?at * eyes", "my cat has very bright eyes"))
	require.True(t, MatchPath("", ""))
	require.False(t, MatchPath("", "b"))
	require.True(t, MatchPath("*ä", "åä"))
	require.True(t, MatchPath("abc", "abc"))
	require.True(t, MatchPath("a*c", "abc"))
	require.True(t, MatchPath("a*c", "a12345c"))
	require.True(t, MatchPath("a?c", "a1c"))
	require.True(t, MatchPath("?at", "cat"))
	require.True(t, MatchPath("?at", "fat"))
	require.True(t, MatchPath("*", "abc"))
	require.True(t, MatchPath(`\*`, "*"))
	require.False(t, MatchPath("?at", "at"))
	require.True(t, MatchPath("*test", "this is a test"))
	require.True(t, MatchPath("this*", "this is a test"))
	require.True(t, MatchPath("*is *", "this is a test"))
	require.True(t, MatchPath("*is*a*", "this is a test"))
	require.True(t, MatchPath("**test**", "this is a test"))
	require.True(t, MatchPath("**is**a***test*", "this is a test"))
	require.False(t, MatchPath("*is", "this is a test"))
	require.False(t, MatchPath("*no*", "this is a test"))
	require.True(t, MatchPath("[!a]*", "this is a test3"))
	require.True(t, MatchPath("*abc", "abcabc"))
	require.True(t, MatchPath("**abc", "abcabc"))
	require.True(t, MatchPath("???", "abc"))
	require.True(t, MatchPath("?*?", "abc"))
	require.True(t, MatchPath("?*?", "ac"))
	require.False(t, MatchPath("sta", "stagnation"))
	require.True(t, MatchPath("sta*", "stagnation"))
	require.False(t, MatchPath("sta?", "stagnation"))
	require.False(t, MatchPath("sta?n", "stagnation"))
	require.True(t, MatchPath("{abc,def}ghi", "defghi"))
	require.True(t, MatchPath("{abc,abcd}a", "abcda"))
	require.True(t, MatchPath("{a,ab}{bc,f}", "abc"))
	require.True(t, MatchPath("{*,**}{a,b}", "ab"))
	require.False(t, MatchPath("{*,**}{a,b}", "ac"))
	require.True(t, MatchPath("/{rate,[a-z][a-z][a-z]}*", "/rate"))
	require.True(t, MatchPath("/{rate,[0-9][0-9][0-9]}*", "/rate"))
	require.True(t, MatchPath("/{rate,[a-z][a-z][a-z]}*", "/usd"))
}
