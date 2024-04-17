package main

import (
	"encoding/json"

	"github.com/NikhilSharmaWe/go-vercel-app/vercel/app"
	"github.com/NikhilSharmaWe/go-vercel-app/vercel/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

func setupUploadSvcRabbitMQForStartup(application *app.Application) (<-chan amqp.Delivery, error) {
	if err := application.UploadResponseClient.CreateBinding(
		"upload-response-"+application.RabbitMQInstanceID,
		"upload-response-"+application.RabbitMQInstanceID,
		"upload",
	); err != nil {
		return nil, err
	}

	uploadRespMSGBus, err := application.UploadResponseClient.Consume("upload-response-"+application.RabbitMQInstanceID, "upload-service", false)
	if err != nil {
		return nil, err
	}

	return uploadRespMSGBus, nil
}

func setupDeploySvcRabbitMQForStartup(application *app.Application) (<-chan amqp.Delivery, error) {
	if err := application.UploadResponseClient.CreateBinding(
		"deploy-response-"+application.RabbitMQInstanceID,
		"deploy-response-"+application.RabbitMQInstanceID,
		"deploy",
	); err != nil {
		return nil, err
	}

	deployRespMSGBus, err := application.UploadResponseClient.Consume("deploy-response-"+application.RabbitMQInstanceID, "deploy-service", false)
	if err != nil {
		return nil, err
	}

	return deployRespMSGBus, nil
}

func handleRabbitMQResponses(application *app.Application, msg amqp.Delivery) (*models.RabbitMQResponse, error) {
	response := &models.RabbitMQResponse{}

	if err := json.Unmarshal(msg.Body, &response); err != nil {
		return nil, err
	}

	_, ok := application.ProjectChannels[response.ProjectID]
	if !ok {
		return nil, models.ErrProjectNotExists
	}

	return response, nil
}
