package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ft-t/go-money/cmd/server/internal/handlers"
	"github.com/ft-t/go-money/cmd/server/internal/jobs"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/accounts"
	"github.com/ft-t/go-money/pkg/analytics"
	"github.com/ft-t/go-money/pkg/appcfg"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/ft-t/go-money/pkg/categories"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/currency"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/importers"
	"github.com/ft-t/go-money/pkg/maintenance"
	"github.com/ft-t/go-money/pkg/mappers"
	"github.com/ft-t/go-money/pkg/tags"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/ft-t/go-money/pkg/transactions/applicable_accounts"
	"github.com/ft-t/go-money/pkg/transactions/double_entry"
	"github.com/ft-t/go-money/pkg/transactions/rules"
	"github.com/ft-t/go-money/pkg/transactions/validation"
	"github.com/ft-t/go-money/pkg/users"
	"github.com/rs/zerolog/log"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	config := configuration.GetConfiguration()
	_, cancel := context.WithCancel(context.Background())

	boilerplate.SetupZeroLog()

	ops := boilerplate.NewOpsHttpServer().StartAsync(config.OpsHttpPort)

	logger := log.Logger

	if err := database.InitDb(); err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to initialize database")
	}

	if config.JwtPrivateKey == "" {
		logger.Warn().Msgf("jwt private key is empty. Will create a new temporary key")

		keyGen := auth.NewKeyGenerator()
		newKey := keyGen.Generate()

		config.JwtPrivateKey = string(keyGen.Serialize(newKey))
	} else {
		config.JwtPrivateKey = strings.ReplaceAll(config.JwtPrivateKey, "\\n", "\n")
	}

	jwtService, err := auth.NewService(config.JwtPrivateKey, 24*time.Hour)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to create jwt service")
	}

	grpcServer := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).
		AddServerMiddleware(middlewares.GrpcMiddleware(jwtService)).Build()

	if config.StaticFilesDirectory != "" {
		logger.Info().Str("dir", config.StaticFilesDirectory).Msg("serving static files from directory")
		grpcServer.GetMux().Handle("/", handlers.SpaHandler(config.StaticFilesDirectory))
	}

	userService := users.NewService(&users.ServiceConfig{
		JwtSvc: jwtService,
	})

	_, err = handlers.NewUserApi(grpcServer, userService)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to create user handler")
	}

	_, err = handlers.NewConfigApi(grpcServer, appcfg.NewService(&appcfg.ServiceConfig{
		UserSvc: userService,
		AppCfg:  configuration.GetConfiguration(),
	}))
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to create config handler")
	}

	decimalSvc := currency.NewDecimalService()
	currencyConverter := currency.NewConverter(config.CurrencyConfig.BaseCurrency)

	_, err = handlers.NewCurrencyApi(
		grpcServer,
		currency.NewService(config.CurrencyConfig),
		currencyConverter,
		decimalSvc,
	)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to create config handler")
	}

	mapper := mappers.NewMapper(&mappers.MapperConfig{
		DecimalSvc: decimalSvc,
	})

	accountSvc := accounts.NewService(&accounts.ServiceConfig{
		MapperSvc:       mapper,
		DefaultCurrency: config.CurrencyConfig.BaseCurrency,
	})

	_, err = handlers.NewAccountsApi(grpcServer, accounts.NewService(&accounts.ServiceConfig{
		MapperSvc: mapper,
	}))
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to create accounts handler")
	}

	baseAmountSvc := transactions.NewBaseAmountService(config.CurrencyConfig.BaseCurrency)
	ruleInterpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{
		AccountsSvc:          accountSvc,
		CurrencyConverterSvc: currencyConverter,
		DecimalSvc:           decimalSvc,
	})

	ruleEngine := rules.NewExecutor(ruleInterpreter)
	applicableAccountSvc := applicable_accounts.NewApplicableAccountService(accountSvc)
	validationSvc := validation.NewValidationService(&validation.ServiceConfig{
		ApplicableAccountSvc: applicableAccountSvc,
	})

	doubleEntry := double_entry.NewDoubleEntryService(&double_entry.DoubleEntryConfig{
		BaseCurrency: config.CurrencyConfig.BaseCurrency,
	})

	statsSvc := transactions.NewStatService()
	transactionSvc := transactions.NewService(&transactions.ServiceConfig{
		StatsSvc:             statsSvc,
		MapperSvc:            mapper,
		CurrencyConverterSvc: currencyConverter,
		BaseAmountService:    baseAmountSvc,
		RuleSvc:              ruleEngine,
		ValidationSvc:        validationSvc,
		DoubleEntry:          doubleEntry,
		AccountSvc:           accountSvc,
	})

	ruleScheduler := rules.NewScheduler(&rules.SchedulerConfig{
		RuleInterpreter: ruleInterpreter,
		TransactionSvc:  transactionSvc,
	})

	if err = ruleScheduler.Reinit(context.TODO()); err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to reinitialize rule scheduler")
	}

	rulesSvc := rules.NewService(mapper)
	rulesScheduleSvc := rules.NewScheduleService(mapper, ruleScheduler)
	tagSvc := tags.NewService(mapper)
	categoriesSvc := categories.NewService(mapper)

	maintenanceSvc := maintenance.NewService(&maintenance.Config{
		StatsSvc: statsSvc,
	})

	recalculateSvc := maintenance.NewRecalculateService(&maintenance.RecalculateServiceConfig{
		AccountSvc:     accountSvc,
		TransactionSvc: transactionSvc,
	})

	dryRunSvc := rules.NewDryRun(&rules.DryRunConfig{
		Executor:       ruleEngine,
		TransactionSvc: transactionSvc,
		MapperSvc:      mapper,
		ValidationSvc:  validationSvc,
		AccountSvc:     accountSvc,
	})
	_ = handlers.NewTransactionApi(grpcServer, transactionSvc, applicableAccountSvc, mapper)
	_ = handlers.NewTagsApi(grpcServer, tagSvc)
	_ = handlers.NewRulesApi(grpcServer, &handlers.RulesApiConfig{
		RulesScheduleSvc: rulesScheduleSvc,
		RuleSvc:          rulesSvc,
		DryRunSvc:        dryRunSvc,
		SchedulerSvc:     ruleScheduler,
	})

	analyticsSvc := analytics.NewService(&analytics.ServiceConfig{
		DecimalSvc:   decimalSvc,
		BaseCurrency: config.CurrencyConfig.BaseCurrency,
	})

	_ = handlers.NewCategoriesApi(grpcServer, categoriesSvc)
	_ = handlers.NewMaintenanceApi(grpcServer, recalculateSvc)
	_ = handlers.NewAnalyticsApi(grpcServer, analyticsSvc)

	baseParser := importers.NewBaseParser(currencyConverter, transactionSvc, mapper)

	importSvc := importers.NewImporter(
		&importers.ImporterConfig{
			AccountSvc:     accountSvc,
			TagSvc:         tagSvc,
			CategoriesSvc:  categoriesSvc,
			TransactionSvc: transactionSvc,
			MapperSvc:      mapper,
		},
		importers.NewFireflyImporter(
			transactionSvc,
			currencyConverter,
			baseParser,
		),
		importers.NewPrivat24(baseParser),
		importers.NewMono(baseParser),
		importers.NewParibas(baseParser),
	)

	_, err = handlers.NewImportApi(grpcServer, importSvc)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to create import handler")
	}

	exchangeRateUpdater := currency.NewSyncer(http.DefaultClient, baseAmountSvc, config.CurrencyConfig)

	if err = accountSvc.EnsureDefaultAccountsExist(context.TODO()); err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to ensure default accounts exist")
	}

	go func() {
		if len(config.ExchangeRatesUrl) > 0 {
			if currencyErr := exchangeRateUpdater.Sync(context.TODO(), config.ExchangeRatesUrl); currencyErr != nil {
				logger.Err(err).Msg("cannot update exchange rates")
			}
		}
	}()

	go func() {
		if jobErr := maintenanceSvc.FixDailyGaps(context.TODO()); jobErr != nil {
			logger.Err(jobErr).Msg("cannot fix daily gaps")
			return
		}

		logger.Info().Msg("daily gaps fixed successfully")
	}()

	jobScheduler, err := jobs.NewJobScheduler(&jobs.Config{
		Configuration:          *config,
		ExchangeRatesUpdateSvc: exchangeRateUpdater,
		MaintenanceSvc:         maintenanceSvc,
	})
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to create job scheduler")
	}

	if err = jobScheduler.StartAsync(); err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to start job scheduler")
	}

	logger.Info().Msg("job scheduler started")

	go func() {
		grpcServer.ServeAsync(config.GrpcPort)

		log.Logger.Info().Msgf("server started on port %v", config.GrpcPort)
	}()

	go func() {
		time.Sleep(200 * time.Millisecond)
		ops.Ready()
	}()

	sg := <-sig

	log.Logger.Info().Msgf("GOT SIGNAL %v", sg.String())
	log.Logger.Info().Msgf("[Graceful Shutdown] GOT SIGNAL %v", sg.String())

	log.Logger.Info().Msgf("[Graceful Shutdown] Shutting down webservers")
	if err = jobScheduler.Stop(); err != nil {
		log.Logger.Error().Err(err).Msg("failed to stop job scheduler")
	} else {
		log.Logger.Info().Msg("job scheduler stopped")
	}

	cancel()
	_ = grpcServer.Shutdown(context.TODO())

	log.Logger.Info().Msg("[Graceful Shutdown] Exit")
}
