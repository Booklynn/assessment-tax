package tax

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func SetPersonalAllowanceAmount(c echo.Context) error {
	var requestBody Allowances

	if err := c.Bind(&requestBody); err != nil {
		return err
	}

	if requestBody.Amount > 100000 || requestBody.Amount < 10000 {
		errorMessage := "Your error message here"
		return c.String(http.StatusBadRequest, errorMessage)
	}

	_, err := conn.Exec("UPDATE allowances SET personal = $1 WHERE id = $2", requestBody.Amount, 1)
	if err != nil {
		return err
	}

	allowancesDeduction := AllowancesDeduction{
		PersonalDeduction: requestBody.Amount,
	}

	return c.JSON(http.StatusOK, allowancesDeduction)
}
