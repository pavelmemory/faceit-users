// +build integration

package test

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO: it should have setup procedure as well as cleanup procedure

func TestUserCRUD(t *testing.T) {
	httpClient := http.DefaultClient

	// create
	createResp, err := httpClient.Post(
		"http://localhost:8080/users",
		"application/json",
		strings.NewReader(`{"first_name": "f", "last_name":"l", "nickname":"n", "email":"e@mail.com", "password": "password", "country":"X1"}`),
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	location := createResp.Header.Get("location")
	require.NotEmpty(t, location)

	// retrieve
	getResp, err := httpClient.Get("http://localhost:8080" + location)
	require.NoError(t, err)
	require.Contains(t, getResp.Header.Get("content-type"), "application/json")
	getData, err := ioutil.ReadAll(getResp.Body)
	require.NoError(t, err)
	require.JSONEq(t, `{"first_name":"f","last_name":"l","nickname":"n","email":"e@mail.com","country":"X1"}`, string(getData))

	// update
	putReq, err := http.NewRequest(
		http.MethodPut,
		"http://localhost:8080"+location,
		strings.NewReader(`{"first_name":"f2","last_name":"l2","nickname":"n2","email":"e@mail2.com","country":"X2"}`),
	)
	require.NoError(t, err)
	putReq.Header.Set("content-type", "application/json")
	putResp, err := httpClient.Do(putReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, putResp.StatusCode)

	// delete
	deleteReq, err := http.NewRequest(http.MethodDelete, "http://localhost:8080"+location, nil)
	require.NoError(t, err)
	deleteResp, err := httpClient.Do(deleteReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

	// attempt to find removed entity
	reGetResp, err := httpClient.Get("http://localhost:8080" + location)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, reGetResp.StatusCode)
}
