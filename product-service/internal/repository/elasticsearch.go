package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"log"
	"product-catalog-service/internal/model"
	"strings"
)

type ElasticRepository struct {
	ElasticClient *elasticsearch.Client
	IndexName     string
}

func NewElasticRepository(elasticClient *elasticsearch.Client, indexName string) *ElasticRepository {
	return &ElasticRepository{
		ElasticClient: elasticClient,
		IndexName:     indexName,
	}
}

func (r *ElasticRepository) GetProducts(ctx context.Context, searchTerm string, from, size int32) ([]*model.Product, error) {
	query := map[string]interface{}{
		"from": from,
		"size": size,
	}

	if searchTerm != "" {
		query["query"] = map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":    searchTerm,
				"fields":   []string{"name^3", "description", "category", "attributes.*"},
				"operator": "and",
			},
		}
	} else {
		query["query"] = map[string]interface{}{
			"match_all": map[string]interface{}{},
		}
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	res, err := r.ElasticClient.Search(
		r.ElasticClient.Search.WithContext(ctx),
		r.ElasticClient.Search.WithIndex(r.IndexName),
		r.ElasticClient.Search.WithBody(&buf),
		r.ElasticClient.Search.WithTrackTotalHits(true),
		r.ElasticClient.Search.WithPretty(),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching documents: %s", res.String())
	}

	var esResponse struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			}
			Hits []struct {
				Source model.Product `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err = json.NewDecoder(res.Body).Decode(&esResponse); err != nil {
		return nil, err
	}

	totalHits := esResponse.Hits.Total.Value
	if totalHits == 0 {
		return []*model.Product{}, nil
	}

	var products []*model.Product
	for _, hit := range esResponse.Hits.Hits {
		product := hit.Source
		products = append(products, &product)
	}

	return products, nil
}

func (r *ElasticRepository) CreateOrUpdateProduct(ctx context.Context, product *model.Product) error {
	docBytes, err := json.Marshal(product)
	if err != nil {
		return err
	}
	res, err := r.ElasticClient.Index(
		"products_idx",
		strings.NewReader(string(docBytes)),
		r.ElasticClient.Index.WithDocumentID(product.ID.Hex()),
		r.ElasticClient.Index.WithRefresh("true"),
		r.ElasticClient.Index.WithContext(ctx),
	)
	if err != nil || res.IsError() {
		log.Printf("Error indexing document in Elasticsearch: %s", err)
	}
	defer res.Body.Close()

	return nil
}

func (r *ElasticRepository) DeleteProduct(ctx context.Context, id string) error {
	res, err := r.ElasticClient.Delete(
		"products_idx",
		id,
		r.ElasticClient.Delete.WithRefresh("true"),
		r.ElasticClient.Delete.WithContext(ctx),
	)
	if err != nil || res.IsError() {
		log.Printf("Error deleting document in Elasticsearch: %s", err)
	}
	defer res.Body.Close()

	return nil
}
