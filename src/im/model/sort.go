package model

type ColumnOrder struct {
	Column string
	Type   string
}

func GetOrder(s string) ColumnOrder {

	order := ColumnOrder{Type: "ASC", Column: "Id"}

	if len(s) > 0 {
		var descending = false
		if s[0:1] == "-" {
			descending = true
			s = s[1:]
		}

		if descending == true {
			order.Type = "DESC"
		}
	}

	order.Column = s

	return order
}

func (s *ColumnOrder) Validate() bool {
	if (s.Type == "DESC" || s.Type == "ASC") && (len(s.Column) >= 2) {
		return true
	}
	return false
}
