package main

import (
	"flag"
	"log"
	"strings"

	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {

	var (
		dbFile = flag.String("db", "", "SQLite数据库文件路径（例如：./foxflow.db 或 /var/lib/foxflow/foxflow.db）")
	)
	flag.Parse()

	// Load config
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if *dbFile != "" {
		config.GlobalConfig.DBFile = *dbFile
	}

	if config.GlobalConfig == nil {
		log.Fatalf("global config is nil")
	}

	// Connection database
	var err error
	db, err := gorm.Open(sqlite.Open(config.GlobalConfig.DBFile), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("failed to connect database: %w", err)
	}

	if err := db.AutoMigrate(
		&models.FoxAccount{},
		&models.FoxSymbol{},
		&models.FoxOrder{},
		&models.FoxExchange{},
	); err != nil {
		log.Fatalf("failed to auto migrate: %w", err)
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
		"uint":      func(detailType gorm.ColumnType) (dataType string) { return "int64" },
		"integer":   func(detailType gorm.ColumnType) (dataType string) { return "int64" },
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

	g.ApplyBasic(orderModel)
	g.ApplyBasic(symbolModel)
	g.ApplyBasic(allModel...)

	g.Execute()
}
