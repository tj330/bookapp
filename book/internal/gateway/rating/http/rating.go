package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/tj330/bookapp/book/internal/gateway"
	"github.com/tj330/bookapp/pkg/discovery"
	"github.com/tj330/bookapp/rating/pkg/model"
)

type Gateway struct {
	registry discovery.Registry
}

func New(registry discovery.Registry) *Gateway {
	return &Gateway{registry}
}

func (g *Gateway) GetAggregatedRating(ctx context.Context, recordId model.RecordID, recordType model.RecordType) (float64, error) {
	addrs, err := g.registry.ServiceAddresses(ctx, "metadata")
	if err != nil {
		return 0, err
	}
	url := "http://" + addrs[rand.Intn(len(addrs))] + "/metadata"
	log.Println("Calling rating service. Request: GET " + url)

	req, err := http.NewRequest(http.MethodGet, url+"/rating", nil)
	if err != nil {
		return 0, err
	}

	req = req.WithContext(ctx)
	values := req.URL.Query()
	values.Add("id", string(recordId))
	values.Add("type", fmt.Sprintf("%v", recordType))
	req.URL.RawQuery = values.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil
	}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return 0, gateway.ErrNotFound
	} else if resp.StatusCode/100 != 2 {
		return 0, fmt.Errorf("non-2xx response: %v", resp)
	}

	var v float64
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return 0, nil
	}
	return v, nil
}

func (g *Gateway) PutRating(ctx context.Context, recordId model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	addrs, err := g.registry.ServiceAddresses(ctx, "rating")
	if err != nil {
		return err
	}
	url := "http://" + addrs[rand.Intn(len(addrs))] + "/metadata"
	log.Println("Calling metadata service. Request: PUT " + url)
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)
	values := req.URL.Query()
	values.Add("id", string(recordId))
	values.Add("type", fmt.Sprintf("%v", recordType))
	values.Add("userId", string(rating.UserID))
	values.Add("value", fmt.Sprintf("%v", rating.Value))
	req.URL.RawQuery = values.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("non-2xx response: %v", resp)
	}
	return nil
}
