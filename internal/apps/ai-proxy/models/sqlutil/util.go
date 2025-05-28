package sqlutil

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func HandleOrderBy(sql *gorm.DB, orderBys []string) (*gorm.DB, error) {
	// order by
	if len(orderBys) == 0 {
		sql = sql.Order("updated_at desc")
	} else {
		for _, orderBy := range orderBys {
			// get is desc or asc
			parts := strings.Split(orderBy, " ")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid order by: %s", orderBy)
			}
			sql = sql.Order(clause.OrderByColumn{
				Column: clause.Column{Name: parts[0], Raw: false},
				Desc:   strings.EqualFold(parts[1], "desc"),
			})
		}
	}
	return sql, nil
}
