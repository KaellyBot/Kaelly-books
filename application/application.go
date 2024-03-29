package application

import (
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-books/models/constants"
	alignRepo "github.com/kaellybot/kaelly-books/repositories/alignments"
	jobRepo "github.com/kaellybot/kaelly-books/repositories/jobs"
	"github.com/kaellybot/kaelly-books/services/alignments"
	"github.com/kaellybot/kaelly-books/services/books"
	"github.com/kaellybot/kaelly-books/services/jobs"
	"github.com/kaellybot/kaelly-books/utils/databases"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func New() (*Impl, error) {
	// misc
	db, err := databases.New()
	if err != nil {
		return nil, err
	}

	broker, err := amqp.New(constants.RabbitMQClientID, viper.GetString(constants.RabbitMQAddress),
		[]amqp.Binding{books.GetBinding()})
	if err != nil {
		return nil, err
	}

	// repositories
	jobBooksRepo := jobRepo.New(db)
	alignBooksRepo := alignRepo.New(db)

	// services
	jobService := jobs.New(broker, jobBooksRepo)
	alignService := alignments.New(broker, alignBooksRepo)
	booksService := books.New(broker, jobService, alignService)

	return &Impl{
		booksService: booksService,
		broker:       broker,
	}, nil
}

func (app *Impl) Run() error {
	return app.booksService.Consume()
}

func (app *Impl) Shutdown() {
	app.broker.Shutdown()
	log.Info().Msgf("Application is no longer running")
}
