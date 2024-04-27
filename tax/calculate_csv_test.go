package tax

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertIncomeWthRowToFloat64(t *testing.T) {
	testCases := []struct {
		row            []string
		expectedIncome float64
		expectedWht    float64
		expectedError  error
	}{
		{[]string{"1000", "200"}, 1000, 200, nil},
		{[]string{"10.5", "20.75"}, 10.5, 20.75, nil},
		{[]string{"invalid", "20.75"}, 0, 0, errors.New("cannot parse str to float64")},
		{[]string{"100.5", "invalid"}, 0, 0, errors.New("cannot parse str to float64")},
	}

	for _, tt := range testCases {
		income, wht, err := convertIncomeWthRowToFloat64(tt.row)
		assert.Equal(t, tt.expectedIncome, income)
		assert.Equal(t, tt.expectedWht, wht)
		assert.Equal(t, tt.expectedError, err)
	}
}

func TestGetAllowancesListCSV(t *testing.T) {
	tests := []struct {
		row            []string
		header         []string
		expectedResult []Allowances
		expectedError  error
	}{
		{
			[]string{"600000", "40000", "20000"},
			[]string{"totalIncome", "wht", "donation"},
			[]Allowances{
				{AllowanceType: "donation", Amount: 20000},
			},
			nil,
		},
		{
			[]string{"600000", "40000", "invalid"},
			[]string{"totalIncome", "wht", "donation"},
			nil,
			errors.New("cannot parse str to float64"),
		},
	}

	for _, test := range tests {
		result, err := getAllowancesListCSV(test.row, test.header)

		assert.Equal(t, test.expectedError, err)

		if test.expectedResult != nil {
			assert.Len(t, result, len(test.expectedResult))
			for i := range test.expectedResult {
				assert.Equal(t, test.expectedResult[i].AllowanceType, result[i].AllowanceType)
				assert.Equal(t, test.expectedResult[i].Amount, result[i].Amount)
			}
		} else {
			assert.Nil(t, result)
		}
	}
}
