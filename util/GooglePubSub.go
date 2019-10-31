package util

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
)

// GooglePubSub TODO
type GooglePubSub struct {
	ProjectID string
	TopicID   string
}

// InProject TODO
func (g *GooglePubSub) InProject(projectID string) *GooglePubSub {
	g.ProjectID = projectID
	return g
}

// Topic TODO
func (g *GooglePubSub) Topic(topicID string) *GooglePubSub {
	g.TopicID = topicID
	return g
}

// Pub TODO
func (g *GooglePubSub) Pub(data []byte) error {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, g.ProjectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}

	t := client.Topic(g.TopicID)
	ok, err := t.Exists(ctx)
	if err != nil {
		return fmt.Errorf("pubsub.Exists: %v", err)
	}
	if !ok {
		t, err = client.CreateTopic(ctx, g.TopicID)
		if err != nil {
			return fmt.Errorf("pubsub.CreateTopic: %v", err)
		}
	}

	result := t.Publish(ctx, &pubsub.Message{
		Data: data,
	})
	// Block until the result is returned and a server-generated
	// ID is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("Get: %v", err)
	}
	log.Println("Published a message; msg ID: ", id)
	return nil
}
