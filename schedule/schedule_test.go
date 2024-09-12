package schedule

import (
	"context"
	"testing"
	"time"

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

	weekAhead := time.Now().Add(time.Hour * 24 * 7)
	monthBack := weekAhead.Add(-time.Hour * 24 * 30)

	result, err := GetClassDates("5d", s, monthBack, weekAhead)
	require.NoError(t, err)
	spew.Dump(result)
}
