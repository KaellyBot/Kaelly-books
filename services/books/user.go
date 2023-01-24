package books

import (
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-configurator/models/constants"
	"github.com/kaellybot/kaelly-configurator/models/entities"
	"github.com/kaellybot/kaelly-configurator/models/mappers"
	"github.com/rs/zerolog/log"
)

func (service *BooksServiceImpl) userRequest(message *amqp.RabbitMQMessage, correlationId string) {
	request := message.JobGetUserRequest
	if !isValidJobGetUserRequest(request) {
		service.publishFailedGetUserAnswer(correlationId, message.Language)
		return
	}

	log.Info().Str(constants.LogCorrelationId, correlationId).
		Msgf("Get job user request received")

	books, err := service.jobBookRepo.GetUserBook(request.UserId, request.ServerId)
	if err != nil {
		service.publishFailedGetBookAnswer(correlationId, message.Language)
		return
	}

	service.publishSucceededGetUserAnswer(correlationId, request.ServerId, books, message.Language)
}

func (service *BooksServiceImpl) publishSucceededGetUserAnswer(correlationId, serverId string,
	books []entities.JobBook, lg amqp.Language) {

	message := amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_JOB_GET_USER_ANSWER,
		Status:   amqp.RabbitMQMessage_SUCCESS,
		Language: lg,
		JobGetUserAnswer: &amqp.JobGetUserAnswer{
			Jobs:     mappers.MapJobExperiences(books),
			ServerId: serverId,
		},
	}

	err := service.broker.Publish(&message, amqp.ExchangeAnswer, answersRoutingkey, correlationId)
	if err != nil {
		log.Error().Err(err).
			Str(constants.LogCorrelationId, correlationId).
			Msgf("Cannot publish via broker, request ignored")
	}
}

func (service *BooksServiceImpl) publishFailedGetUserAnswer(correlationId string, lg amqp.Language) {
	message := amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_JOB_GET_USER_ANSWER,
		Status:   amqp.RabbitMQMessage_FAILED,
		Language: lg,
	}

	err := service.broker.Publish(&message, amqp.ExchangeAnswer,
		answersRoutingkey, correlationId)
	if err != nil {
		log.Error().Err(err).
			Str(constants.LogCorrelationId, correlationId).
			Msgf("Cannot publish via broker, request ignored")
	}
}

func isValidJobGetUserRequest(request *amqp.JobGetUserRequest) bool {
	return request != nil
}