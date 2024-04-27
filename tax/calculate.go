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

	personalAllowanceAmount, err := getPersonalAllowance()
	if err != nil {
		return err
	}

	otherAllowancesAmount, err := getAllowancesAmount(requestBody)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	allowancesAmount := personalAllowanceAmount + otherAllowancesAmount
	tax := calculateTaxByLevels(requestBody.TotalIncome, allowancesAmount)
	taxLevel := displayTaxLevel(requestBody.TotalIncome, allowancesAmount, tax)

	taxPayable := TaxPayable{
		Tax:       (math.Round(tax*100) / 100),
		TaxLevels: taxLevel,
	}

	taxPayable.Tax = taxPayable.Tax - requestBody.WHT
	taxPayable.Tax = math.Round(taxPayable.Tax*100) / 100

	if taxPayable.Tax >= 0 {
		return c.JSON(http.StatusOK, taxPayable)
	}

	taxReturnable := TaxReturnable{
		TaxRefund: math.Round(math.Abs(taxPayable.Tax)*100) / 100,
		TaxLevels: taxLevel,
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
	allowanceTypes := []string{}

	for _, allowance := range requestBody.Allowances {
		allowanceType := strings.ToLower(allowance.AllowanceType)

		if allowance.Amount < 0 {
			return errors.New("allowance amount cannot be less than 0")
		}

		found := false
		for _, existingType := range allowanceTypes {
			if existingType == allowanceType {
				found = true
				return errors.New("found allowanceType duplication")
			}
		}

		if !found {
			allowanceTypes = append(allowanceTypes, allowanceType)
		}

		if !allowedAllowances[allowanceType] {
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

func getAllowancesAmount(requestBody TaxInfo) (float64, error) {
	var allowancesAmount float64

	for _, allowance := range requestBody.Allowances {
		allowanceType := strings.ToLower(allowance.AllowanceType)

		if allowanceType == "donation" {
			donationAllowance, err := getDonationAllowance()
			if err != nil {
				return 0, err
			}

			if allowance.Amount > donationAllowance {
				allowance.Amount = donationAllowance
			}

			allowancesAmount += allowance.Amount
		}

		if allowanceType == "k-receipt" {
			kReceiptAllowance, err := getKReceiptAllowance()
			if err != nil {
				return 0, err
			}

			if allowance.Amount > kReceiptAllowance {
				allowance.Amount = kReceiptAllowance
			}

			allowancesAmount += allowance.Amount
		}
	}

	return allowancesAmount, nil
}

func displayTaxLevel(totalIncome, allowance, tax float64) []TaxLevel {
	taxLevels := []TaxLevel{
		{Level: "0-150,000", Tax: 0.0},
		{Level: "150,001-500,000", Tax: 0.0},
		{Level: "500,001-1,000,000", Tax: 0.0},
		{Level: "1,000,001-2,000,000", Tax: 0.0},
		{Level: "2,000,001 ขึ้นไป", Tax: 0.0},
	}

	switch {
	case totalIncome-allowance <= 150000:
		taxLevels[0].Tax = tax
	case totalIncome-allowance <= 500000:
		taxLevels[1].Tax = tax
	case totalIncome-allowance <= 1000000:
		taxLevels[2].Tax = tax
	case totalIncome-allowance <= 2000000:
		taxLevels[3].Tax = tax
	default:
		taxLevels[4].Tax = tax
	}

	return taxLevels
}
