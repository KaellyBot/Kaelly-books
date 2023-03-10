package jobs

import (
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-configurator/models/constants"
	"github.com/kaellybot/kaelly-configurator/models/entities"
	"github.com/kaellybot/kaelly-configurator/models/mappers"
	"github.com/rs/zerolog/log"
)

func (service *JobServiceImpl) GetBookRequest(request *amqp.JobGetBookRequest,
	correlationId, answersRoutingkey string, lg amqp.Language) {

	if !isValidJobGetRequest(request) {
		service.publishFailedGetBookAnswer(correlationId, answersRoutingkey, lg)
		return
	}

	log.Info().Str(constants.LogCorrelationId, correlationId).
		Str(constants.LogJobId, request.JobId).
		Str(constants.LogServerId, request.ServerId).
		Msgf("Get job books request received")

	books, err := service.jobBookRepo.GetBooks(request.JobId, request.ServerId,
		request.UserIds, int(request.Limit))
	if err != nil {
		service.publishFailedGetBookAnswer(correlationId, answersRoutingkey, lg)
		return
	}

	service.publishSucceededGetBookAnswer(correlationId, request.JobId,
		request.ServerId, answersRoutingkey, books, lg)
}

func (service *JobServiceImpl) publishSucceededGetBookAnswer(correlationId, jobId, serverId,
	answersRoutingkey string, books []entities.JobBook, lg amqp.Language) {

	message := amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_JOB_GET_BOOK_ANSWER,
		Status:   amqp.RabbitMQMessage_SUCCESS,
		Language: lg,
		JobGetBookAnswer: &amqp.JobGetBookAnswer{
			JobId:     jobId,
			ServerId:  serverId,
			Craftsmen: mappers.MapCraftsmen(books),
		},
	}

	err := service.broker.Publish(&message, amqp.ExchangeAnswer, answersRoutingkey, correlationId)
	if err != nil {
		log.Error().Err(err).
			Str(constants.LogCorrelationId, correlationId).
			Msgf("Cannot publish via broker, request ignored")
	}
}

func (service *JobServiceImpl) publishFailedGetBookAnswer(correlationId, answersRoutingkey string,
	lg amqp.Language) {

	message := amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_JOB_GET_BOOK_ANSWER,
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

func isValidJobGetRequest(request *amqp.JobGetBookRequest) bool {
	return request != nil
}
