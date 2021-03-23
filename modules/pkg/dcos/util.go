package dcos

import (
	"fmt"
	"strconv"
)

func outputRows(rows [][]string) {
	var cols []int
	for _, row := range rows {
		for i, col := range row {
			if i < len(cols) {
				if len(col) > cols[i] {
					cols[i] = len(col)
				}
			} else {
				cols = append(cols, len(col))
			}
		}
	}
	format := ""
	for _, col := range cols {
		format += "%-" + strconv.Itoa(col+4) + "s"
	}
	format += "\n"
	args := make([]interface{}, len(cols))
	for _, row := range rows {
		args = args[:0]
		for _, col := range row {
			args = append(args, col)
		}
		fmt.Printf(format, args...)
	}
}

func app2Row(m map[string]interface{}) (a []string) {
	a = []string{
		m["id"].(string),
		strconv.FormatFloat(m["mem"].(float64), 'f', -1, 64),
		strconv.FormatFloat(m["cpus"].(float64), 'f', -1, 64),
	}
	tasksRunning := int(m["tasksRunning"].(float64))
	a = append(a, fmt.Sprintf("%d/%d", tasksRunning, int(m["instances"].(float64))))
	_, ok := m["healthChecks"]
	if ok {
		a = append(a, fmt.Sprintf("%d/%d", int(m["tasksHealthy"].(float64)), tasksRunning))
	} else {
		a = append(a, "---")
	}
	_, ok = m["overdue"]
	if ok {
		a = append(a, "True")
	} else {
		a = append(a, "False")
	}
	return
}

func OutputApps(a []map[string]interface{}) {
	var rows [][]string
	rows = append(rows, []string{"ID", "MEM", "CPUS", "TASKS", "HEALTH", "WAITING"})
	for _, m := range a {
		rows = append(rows, app2Row(m))
	}
	outputRows(rows)
}
