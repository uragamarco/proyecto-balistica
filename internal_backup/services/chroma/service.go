package chroma

import (
	"context"
	"fmt"

	chromago "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/types"
	"gocv.io/x/gocv"
)

var (
	client     chromago.Client
	collection chromago.Collection
)

func Init(url, collectionName string) error {
	var err error
	client, err = chromago.NewClient(chroma.WithBasePath(url))
	if err != nil {
		return fmt.Errorf("error creating Chroma client: %w", err)
	}

	collection, err = client.GetOrCreateCollection(
		context.Background(),
		collectionName,
		nil,
		types.L2,
	)
	return err
}

func StoreDescriptors(imageID string, desc gocv.Mat) error {
	embeddings := make([][]float32, desc.Rows())
	for i := 0; i < desc.Rows(); i++ {
		vec := make([]float32, desc.Cols())
		for j := 0; j < desc.Cols(); j++ {
			vec[j] = desc.GetFloatAt(i, j)
		}
		embeddings[i] = vec
	}

	_, err := collection.Add(
		context.Background(),
		embeddings,
		[]map[string]interface{}{{"source_image": imageID}},
		[]string{imageID},
		nil,
	)
	return err
}

func QuerySimilar(queryDesc gocv.Mat, topK int) ([]map[string]interface{}, error) {
	queryEmbeddings := make([][]float32, queryDesc.Rows())
	for i := 0; i < queryDesc.Rows(); i++ {
		vec := make([]float32, queryDesc.Cols())
		for j := 0; j < queryDesc.Cols(); j++ {
			vec[j] = queryDesc.GetFloatAt(i, j)
		}
		queryEmbeddings[i] = vec
	}

	results, err := collection.Query(
		context.Background(),
		queryEmbeddings,
		topK,
		nil,
		nil,
		[]types.QueryEnum{"metadatas", "distances"},
	)
	if err != nil {
		return nil, err
	}

	var matches []map[string]interface{}
	for i := range results.Distances {
		matches = append(matches, map[string]interface{}{
			"image":    results.Metadatas[i][0]["source_image"],
			"distance": results.Distances[i][0],
			"score":    1 - results.Distances[i][0],
		})
	}
	return matches, nil
}

func GetService() *chromago.Collection {
	return &collection
}
