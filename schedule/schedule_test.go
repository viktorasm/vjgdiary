package schedule

import (
	"context"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

func Test_parseSchedule(t *testing.T) {
	//contents, err := os.ReadFile("rttdata.json")
	//if err != nil {
	//	t.Skipf("failed to read sample rttdata.json: %v", err)
	//}
	//s := Schedule{}
	// require.NoError(t, json.Unmarshal(contents, &s))
	d, err := NewDownloader()
	require.NoError(t, err)
	s, err := d.GetSchedule(context.Background())
	require.NoError(t, err)

	result, err := GetNextClassDates("5d", s)
	require.NoError(t, err)
	spew.Dump(result)
}
