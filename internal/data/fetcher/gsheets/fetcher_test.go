package gsheets

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseRow_ValidData(t *testing.T) {
	i := &impl{}

	row := []any{
		"22/12/2024",    // date
		"13,224,723.00", // GMV
		"123",           // users
		"456,789.50",    // marketing cost
		"10",            // FE pods
		"5",             // BE pods
	}

	daily, err := i.parseRow(row, 1)
	assert.NoError(t, err)
	assert.Equal(t, 13224723.00, daily.GMV)
}

func TestParseRow_DateParsed(t *testing.T) {
	i := &impl{}

	row := []any{
		"22/12/2024",
		"1000",
		"10",
		"50",
	}

	daily, err := i.parseRow(row, 2)
	assert.NoError(t, err)

	expectedDate, _ := time.Parse(dateLayout, "22/12/2024")
	assert.Equal(t, expectedDate, daily.Date)
}

func TestParseRow_OptionalFieldsMissing(t *testing.T) {
	i := &impl{}

	row := []any{
		"01/01/2025",
		"1000",
		"1",
		"50",
		"", // FE pods missing
		"", // BE pods missing
	}

	daily, err := i.parseRow(row, 3)
	assert.NoError(t, err)
	assert.Nil(t, daily.FEPods)
	assert.Nil(t, daily.BEPods)
}

func TestParseRow_InvalidDate(t *testing.T) {
	i := &impl{}

	row := []any{
		"invalid-date",
		"1000",
		"1",
		"50",
	}

	_, err := i.parseRow(row, 4)
	assert.Error(t, err)
}

func TestParseRow_InvalidGMV(t *testing.T) {
	i := &impl{}

	row := []any{
		"01/01/2025",
		"13,abc",
		"1",
		"50",
	}

	_, err := i.parseRow(row, 5)
	assert.Error(t, err)
}

func TestParseRow_InvalidUsers(t *testing.T) {
	i := &impl{}

	row := []any{
		"01/01/2025",
		"1000",
		"abc",
		"50",
	}

	_, err := i.parseRow(row, 6)
	assert.Error(t, err)
}

func TestParseRow_InvalidMarketingCost(t *testing.T) {
	i := &impl{}

	row := []any{
		"01/01/2025",
		"1000",
		"1",
		"abc",
	}

	_, err := i.parseRow(row, 7)
	assert.Error(t, err)
}
