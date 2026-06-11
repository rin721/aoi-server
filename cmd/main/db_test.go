package main

// 本测试文件固定命令行入口和数据库命令的契约，防止注释补全和后续重构改变外部可观察行为。

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/rei0721/go-scaffold/internal/app/dbapp"
	appconfig "github.com/rei0721/go-scaffold/internal/config"
	"github.com/rei0721/go-scaffold/pkg/cli"
	"github.com/rei0721/go-scaffold/types/constants"
)

// TestDBCommandMetadata 固定命令行入口和数据库命令的契约，确保后续注释补全或结构调整不改变该场景。
func TestDBCommandMetadata(t *testing.T) {
	cmd := NewDBCommand()

	if got := cmd.Name(); got != dbCommandName {
		t.Fatalf("Name() = %q, want %q", got, dbCommandName)
	}
	if !strings.Contains(cmd.Description(), "sqlgen-powered") {
		t.Fatalf("Description() = %q, want sqlgen wording", cmd.Description())
	}
	if !strings.Contains(cmd.Usage(), "--operation") {
		t.Fatalf("Usage() = %q, want operation flag", cmd.Usage())
	}
	if !strings.Contains(cmd.Usage(), dbOperationDatabase) {
		t.Fatalf("Usage() = %q, want database operation", cmd.Usage())
	}

	gotFlags := cmd.Flags()
	wantNames := []string{"config", "operation", "apply", "id", "title", "description", "completed", "limit", "offset", "print-sql"}
	if len(gotFlags) != len(wantNames) {
		t.Fatalf("len(Flags()) = %d, want %d", len(gotFlags), len(wantNames))
	}
	for i, want := range wantNames {
		if gotFlags[i].Name != want {
			t.Fatalf("Flags()[%d].Name = %q, want %q", i, gotFlags[i].Name, want)
		}
	}
	if gotFlags[0].Default != constants.AppDefaultConfigPath {
		t.Fatalf("config default = %v, want %q", gotFlags[0].Default, constants.AppDefaultConfigPath)
	}
	if gotFlags[1].Default != defaultDBOperation {
		t.Fatalf("operation default = %v, want %q", gotFlags[1].Default, defaultDBOperation)
	}
	if gotFlags[6].Name != dbCompletedFlagName {
		t.Fatalf("completed flag name = %q, want %q", gotFlags[6].Name, dbCompletedFlagName)
	}
	if gotFlags[7].Default != dbapp.DefaultTodoLimit {
		t.Fatalf("limit default = %v, want %d", gotFlags[7].Default, dbapp.DefaultTodoLimit)
	}
}

// TestDBSQLForPrintDatabaseOperation 固定命令行入口和数据库命令的契约，确保后续注释补全或结构调整不改变该场景。
func TestDBSQLForPrintDatabaseOperation(t *testing.T) {
	sql, err := dbSQLForPrint(dbOptions{Operation: dbOperationDatabase}, appconfig.DatabaseConfig{
		Driver: "mysql",
		DBName: "demo_app",
	})
	if err != nil {
		t.Fatalf("dbSQLForPrint() error = %v", err)
	}
	if sql != "CREATE DATABASE IF NOT EXISTS `demo_app`;" {
		t.Fatalf("dbSQLForPrint() = %q, want sqlgen CREATE DATABASE", sql)
	}
}

// TestDBCommandExecutePassesOptionsToRunner 固定命令行入口和数据库命令的契约，确保后续注释补全或结构调整不改变该场景。
func TestDBCommandExecutePassesOptionsToRunner(t *testing.T) {
	var got dbOptions
	cmd := &DBCommand{
		runner: func(_ *cli.Context, opts dbOptions) error {
			got = opts
			return nil
		},
	}

	var stdout bytes.Buffer
	err := cmd.Execute(&cli.Context{
		Flags: map[string]interface{}{
			"config":      "configs/test.yaml",
			"operation":   dbOperationCreate,
			"apply":       true,
			"id":          7,
			"title":       "demo",
			"description": "created by sqlgen",
			"completed":   "true",
			"limit":       20,
			"offset":      5,
			"print-sql":   true,
		},
		Stdout: &stdout,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	want := dbOptions{
		ConfigPath:  "configs/test.yaml",
		Operation:   dbOperationCreate,
		Apply:       true,
		ID:          7,
		Title:       "demo",
		Description: "created by sqlgen",
		Completed:   "true",
		Limit:       20,
		Offset:      5,
		PrintSQL:    true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("options = %#v, want %#v", got, want)
	}
}

// TestDBCommandExecuteDefaultsOperation 固定命令行入口和数据库命令的契约，确保后续注释补全或结构调整不改变该场景。
func TestDBCommandExecuteDefaultsOperation(t *testing.T) {
	var got dbOptions
	cmd := &DBCommand{
		runner: func(_ *cli.Context, opts dbOptions) error {
			got = opts
			return nil
		},
	}

	err := cmd.Execute(&cli.Context{
		Flags: map[string]interface{}{
			"config":      constants.AppDefaultConfigPath,
			"operation":   "",
			"apply":       false,
			"id":          0,
			"title":       "",
			"description": "",
			"completed":   "",
			"limit":       dbapp.DefaultTodoLimit,
			"offset":      0,
			"print-sql":   false,
		},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got.Operation != defaultDBOperation {
		t.Fatalf("operation = %q, want %q", got.Operation, defaultDBOperation)
	}
}
