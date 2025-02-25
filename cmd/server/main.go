package main

import (
	"context"
	"github.com/ft-t/go-money/pkg/accounts"
	"github.com/ft-t/go-money/pkg/appcfg"
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/currency"
	"github.com/ft-t/go-money/pkg/jwt"
	"github.com/ft-t/go-money/pkg/mappers"
	"github.com/ft-t/go-money/pkg/users"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	config := configuration.GetConfiguration()
	_, cancel := context.WithCancel(context.Background())

	logger := log.Logger

	if config.JwtPrivateKey == "" {
		logger.Warn().Msgf("jwt private key is empty. Will create a new temporary key")

		keyGen := jwt.NewKeyGenerator()
		newKey := keyGen.Generate()

		config.JwtPrivateKey = string(keyGen.Serialize(newKey))
	}

	grpcServer := boilerplate.GetDefaultGrpcServerBuilder().Build()

	jwtService, err := users.NewJwtGenerator(config.JwtPrivateKey)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to create jwt service")
	}

	userService := users.NewService(&users.ServiceConfig{
		JwtSvc: jwtService,
	})

	_, err = NewUserApi(grpcServer, userService)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to create user handler")
	}

	_, err = NewConfigApi(grpcServer, appcfg.NewService(&appcfg.ServiceConfig{
		UserSvc: userService,
	}))
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to create config handler")
	}

	decimalSvc := currency.NewDecimalService()

	_, err = NewCurrencyApi(grpcServer, currency.NewService(), currency.NewConverter(), decimalSvc)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to create config handler")
	}

	mapper := mappers.NewMapper(&mappers.MapperConfig{
		DecimalSvc: decimalSvc,
	})

	_, err = NewAccountsApi(grpcServer, accounts.NewService(&accounts.ServiceConfig{
		MapperSvc: mapper,
	}))
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to create accounts handler")
	}

	go func() {
		grpcServer.ServeAsync(config.GrpcPort)

		log.Logger.Info().Msgf("server started on port %v", config.GrpcPort)
	}()

	sg := <-sig

	log.Logger.Info().Msgf("GOT SIGNAL %v", sg.String())
	log.Logger.Info().Msgf("[Graceful Shutdown] GOT SIGNAL %v", sg.String())

	log.Logger.Info().Msgf("[Graceful Shutdown] Shutting down webservers")

	cancel()
	_ = grpcServer.Shutdown(context.TODO())

	log.Logger.Info().Msg("[Graceful Shutdown] Exit")
}
