package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/lemconn/foxflow/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

func main() {
	// Load config
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if config.GlobalConfig == nil {
		log.Fatalf("global config is nil")
	}

	dbs := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		config.GlobalConfig.DBConfig.Username,
		config.GlobalConfig.DBConfig.Password,
		config.GlobalConfig.DBConfig.Host,
		config.GlobalConfig.DBConfig.Port,
		config.GlobalConfig.DBConfig.DbName,
		config.GlobalConfig.DBConfig.Config)

	// Connection database
	var err error
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dbs,   // DSN data source name
		DefaultStringSize:         256,   // string Default length of type fields
		DisableDatetimePrecision:  true,  // Disable datetime precision, not supported by databases before MySQL 5.6
		DontSupportRenameIndex:    true,  // When renaming the index, delete and create a new one. Databases before MySQL 5.7 and MariaDB do not support renaming indexes.
		DontSupportRenameColumn:   true,  // Use `change` to rename columns. Databases prior to MySQL 8 and MariaDB do not support renaming columns.
		SkipInitializeWithVersion: false, // Automatically configured based on the current MySQL version
	}), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %w", err)
	}

	// Generate instance
	g := gen.NewGenerator(gen.Config{
		OutPath:           "internal/pkg/dao/query",
		ModelPkgPath:      "internal/pkg/dao/model",
		Mode:              gen.WithDefaultQuery | gen.WithoutContext | gen.WithQueryInterface,
		FieldCoverable:    false,
		FieldSignable:     false,
		FieldWithIndexTag: false,
		FieldWithTypeTag:  true,
	})

	// set target DB
	g.UseDB(db)

	// Custom column type
	dataMap := map[string]func(detailType gorm.ColumnType) (dataType string){
		"tinyint": func(detailType gorm.ColumnType) (dataType string) {
			ct, _ := detailType.ColumnType()
			if strings.HasPrefix(ct, "tinyint(1)") {
				return "int"
			}
			return "int64"
		},
		"smallint":  func(detailType gorm.ColumnType) (dataType string) { return "int64" },
		"mediumint": func(detailType gorm.ColumnType) (dataType string) { return "int64" },
		"bigint":    func(detailType gorm.ColumnType) (dataType string) { return "int64" },
		"int":       func(detailType gorm.ColumnType) (dataType string) { return "int64" },
		"float":     func(detailType gorm.ColumnType) (dataType string) { return "float64" },
		"decimal":   func(detailType gorm.ColumnType) (dataType string) { return "decimal.Decimal" },
	}

	// It must be executed before `ApplyBasic`.
	g.WithDataTypeMap(dataMap)

	jsonField := gen.FieldJSONTagWithNS(func(columnName string) (tagContent string) {
		return columnName
	})

	// Model Custom Options Group
	fieldOpts := []gen.ModelOpt{jsonField}
	allModel := g.GenerateAllTable(fieldOpts...)

	exchangeModel := g.GenerateModel("fox_exchanges")
	accountModel := g.GenerateModel("fox_accounts")

	orderModel := g.GenerateModel("fox_orders",
		append(
			fieldOpts,
			gen.FieldRelate(field.BelongsTo, "Account", accountModel,
				&field.RelateConfig{
					GORMTag: field.GormTag{
						"foreignKey": []string{"account_id"},
						"references": []string{"id"},
					},
				},
			),
		)...,
	)
	symbolModel := g.GenerateModel("fox_symbols",
		append(
			fieldOpts,
			gen.FieldRelate(field.BelongsTo, "Account", accountModel,
				&field.RelateConfig{
					GORMTag: field.GormTag{
						"foreignKey": []string{"account_id"},
						"references": []string{"id"},
					},
				},
			),
		)...,
	)

	g.ApplyBasic(exchangeModel)
	g.ApplyBasic(accountModel)
	g.ApplyBasic(orderModel)
	g.ApplyBasic(symbolModel)
	g.ApplyBasic(allModel...)

	g.Execute()
}
