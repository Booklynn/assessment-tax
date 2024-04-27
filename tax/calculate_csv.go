package tax

import (
	"encoding/csv"
	"errors"
	"io"
	"math"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func CalculateTaxWithCSV(c echo.Context) error {
	file, err := c.FormFile("taxes.csv")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "key should be taxes.csv")
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	reader := csv.NewReader(src)

	header, err := reader.Read()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	var taxCSV []TaxCSV

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		totalIncome, wht, err := convertIncomeWthRowToFloat64(row)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		allowances, err := getAllowancesListCSV(row, header)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		taxInfo := TaxInfo{
			TotalIncome: totalIncome,
			WHT:         wht,
			Allowances:  allowances,
		}

		if err := checkTaxInfoNotNegative(taxInfo); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		if err = checkValidTaxAllowances(taxInfo); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		personalAllowanceAmount, err := getPersonalAllowance()
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		otherAllowancesAmount, err := getAllowancesAmount(taxInfo)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		allowancesAmount := personalAllowanceAmount + otherAllowancesAmount
		tax := calculateTaxByLevels(taxInfo.TotalIncome, allowancesAmount)

		taxPayable := TaxCSV{
			TotalIncome: taxInfo.TotalIncome,
			Tax:         (math.Round(tax*100) / 100),
		}

		taxPayable.Tax = taxPayable.Tax - taxInfo.WHT
		taxPayable.Tax = math.Round(taxPayable.Tax*100) / 100

		if taxPayable.Tax >= 0 {
			taxCSV = append(taxCSV, taxPayable)
		} else {
			taxCSV = append(taxCSV, TaxCSV{
				TotalIncome: taxInfo.TotalIncome,
				TaxRefund:   math.Round(math.Abs(taxPayable.Tax)*100) / 100,
			})
		}

	}

	taxCSVResponse := TaxResponseCSV{
		Taxes: taxCSV,
	}

	return c.JSON(http.StatusOK, taxCSVResponse)
}

func convertIncomeWthRowToFloat64(row []string) (float64, float64, error) {
	totalIncomeStr := row[0]
	totalIncome, err := strconv.ParseFloat(totalIncomeStr, 64)
	if err != nil {
		return 0, 0, errors.New("cannot parse str to float64")
	}

	whtStr := row[1]
	wht, err := strconv.ParseFloat(whtStr, 64)
	if err != nil {
		return 0, 0, errors.New("cannot parse str to float64")
	}
	return totalIncome, wht, nil
}

func getAllowancesListCSV(row []string, header []string) ([]Allowances, error) {
	var allowances []Allowances

	for i := 2; i < len(row); i++ {
		allowanceStr := row[i]
		allowance, err := strconv.ParseFloat(allowanceStr, 64)
		if err != nil {
			return nil, errors.New("cannot parse str to float64")
		}

		allowances = append(allowances, Allowances{
			AllowanceType: header[i],
			Amount:        allowance,
		})
	}
	return allowances, nil
}
