package chroma

import (
	"context"
	"fmt"
	"time"

	"github.com/amikos-tech/chroma-go"
	"gocv.io/x/gocv"
)

type Service struct {
	collection *chroma.Collection
}

var (
	cache = make(map[string][]MatchResult)
)

func New(chromaURL, collectionName string) (*Service, error) {
	client, err := chroma.NewClient(chroma.WithBasePath(chromaURL))
	if err != nil {
		return nil, err
	}

	collection, err := client.GetOrCreateCollection(
		context.Background(),
		collectionName,
		nil,
		chroma.L2,
	)
	return &Service{collection}, err
}

func (s *Service) StoreDescriptors(imageID string, desc gocv.Mat) error {
	embeddings := matToEmbeddings(desc)
	_, err := s.collection.Add(
		context.Background(),
		embeddings,
		[]map[string]interface{}{{"source_image": imageID}},
		[]string{imageID},
		nil,
	)
	return err
}

func (s *Service) QuerySimilar(desc gocv.Mat, topK int) ([]MatchResult, error) {
	cacheKey := fmt.Sprintf("%x", desc.ToBytes())
	if results, ok := cache[cacheKey]; ok {
		return results, nil
	}

	results, err := s.collection.Query(
		context.Background(),
		matToEmbeddings(desc),
		topK,
		nil,
		nil,
		[]chroma.QueryEnum{"metadatas", "distances"},
	)
	if err != nil {
		return nil, err
	}

	// Almacenar en cache
	cache[cacheKey] = convertResults(results)
	return cache[cacheKey], nil
}

// Helpers...
