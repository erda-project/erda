// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dbmigrate",
	Short: "certain migration logic",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var alertExpressionCmd = &cobra.Command{
	Use:   "alert-expression",
	Short: "update select field in sp_alert_expression",
	RunE:  expMigrateCommand,
}

func expMigrateCommand(cmd *cobra.Command, args []string) error {
	db := NewDBConn(constructConfig(cmd))
	err := migrateAlertExpression(db)
	return errors.Wrap(err, "execute failed")
}

func constructConfig(cmd *cobra.Command) *config {
	user, _ := cmd.Flags().GetString("user")
	password, _ := cmd.Flags().GetString("password")
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetString("port")
	db, _ := cmd.Flags().GetString("db")
	return &config{
		MySQLHost:     host,
		MySQLUsername: user,
		MySQLPort:     port,
		MySQLPassword: password,
		MySQLDatabase: db,
	}
}

func init() {
	rootCmd.AddCommand(alertExpressionCmd)
	rootCmd.PersistentFlags().String("user", "", "db user")
	rootCmd.PersistentFlags().String("password", "", "db password")
	rootCmd.PersistentFlags().String("host", "localhost", "db host")
	rootCmd.PersistentFlags().String("port", "3306", "db port")
	rootCmd.PersistentFlags().String("db", "", "db database")
	rootCmd.MarkPersistentFlagRequired("user")
	rootCmd.MarkPersistentFlagRequired("password")
	rootCmd.MarkPersistentFlagRequired("db")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
