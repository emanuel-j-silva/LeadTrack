package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWeekStart(t *testing.T) {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	assert.NoError(t, err)

	// Wednesday, Oct 1, 2025
	dt := time.Date(2025, 10, 1, 15, 30, 0, 0, loc)
	ws := WeekStart(dt, loc)

	// Monday of that week should be Sep 29, 2025
	expected := time.Date(2025, 9, 29, 0, 0, 0, 0, loc)
	assert.True(t, expected.Equal(ws), "expected %s, got %s", expected, ws)
}
