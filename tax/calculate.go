package tax

import (
	"errors"
	"math"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

var allowedAllowances = map[string]bool{
	"donation":  true,
	"personal":  true,
	"k-receipt": true,
}

func CalculateTax(c echo.Context) error {
	var requestBody TaxInfo
	var err error

	if err = c.Bind(&requestBody); err != nil {
		return err
	}

	if err := checkTaxInfoNotNegative(requestBody); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	if err = checkValidTaxAllowances(requestBody); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	personalAllowance, err := getPersonalAllowance()
	if err != nil {
		return err
	}

	tax := calculateTaxByLevels(requestBody.TotalIncome, personalAllowance)

	taxPayable := TaxPayable{
		Tax: (math.Round(tax*100) / 100),
	}

	taxPayable.Tax = taxPayable.Tax - requestBody.WHT
	taxPayable.Tax = math.Round(taxPayable.Tax*100) / 100

	if taxPayable.Tax >= 0 {
		return c.JSON(http.StatusOK, taxPayable)
	}

	taxReturnable := TaxReturnable{
		TaxRefund: math.Round(math.Abs(taxPayable.Tax)*100) / 100,
	}

	return c.JSON(http.StatusOK, taxReturnable)
}

func checkTaxInfoNotNegative(taxInfo TaxInfo) error {
	if taxInfo.TotalIncome < 0 || taxInfo.WHT < 0 {
		return errors.New("total income and wht cannot be less than 0")
	}
	return nil
}

func checkValidTaxAllowances(requestBody TaxInfo) error {
	for _, allowance := range requestBody.Allowances {
		if allowance.Amount < 0 {
			return errors.New("allowance amount cannot be less than 0")
		}

		if !allowedAllowances[strings.ToLower(allowance.AllowanceType)] {
			return errors.New("allowanceType not allowed")
		}
	}
	return nil
}

func calculateTaxByLevels(totalIncome, allowance float64) float64 {
	netIncome := totalIncome - allowance

	switch {
	case netIncome <= 150000:
		return 0

	case netIncome <= 500000:
		return (netIncome - 150000) * 0.1

	case netIncome <= 1000000:
		return (netIncome-500000)*0.15 + (500000-150000)*0.10

	case netIncome <= 2000000:
		return (netIncome-1000000)*0.20 + (1000000-500000)*0.15 + (500000-150000)*0.10

	default:
		return (netIncome-2000000)*0.35 + (2000000-1000000)*0.20 + (1000000-500000)*0.15 + (500000-150000)*0.10
	}
}
