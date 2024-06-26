package tax

type Allowances struct {
	AllowanceType string  `json:"allowanceType"`
	Amount        float64 `json:"amount"`
}

type TaxInfo struct {
	TotalIncome float64      `json:"totalIncome"`
	WHT         float64      `json:"wht"`
	Allowances  []Allowances `json:"allowances"`
}

type TaxPayable struct {
	Tax       float64    `json:"tax"`
	TaxLevels []TaxLevel `json:"taxLevel"`
}

type TaxReturnable struct {
	TaxRefund float64    `json:"taxRefund"`
	TaxLevels []TaxLevel `json:"taxLevel"`
}

type TaxLevel struct {
	Level string  `json:"level"`
	Tax   float64 `json:"tax"`
}

type AllowancesPersonalDeduction struct {
	PersonalDeduction float64 `json:"personalDeduction"`
}

type AllowancesKReceiptDeduction struct {
	KReceipt float64 `json:"kReceipt"`
}
