package pagination_test

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	page "github.com/forkyid/go-utils/pagination"
	"github.com/stretchr/testify/assert"
)

const (
	// maximum repetition number for benchmarking
	// recommended value : > 100000
	benchmarkRep = 100000

	// minimum & maximum value for limit
	limitMinimum = -1000000
	limitMaximum = 1000000

	// minimum & maximum value for page
	pageMinimum = -100000
	pageMaximum = 1000000

	// minimum & maximum value for total data
	totalDataMinimum = 0
	totalDataMaximum = 10000000
)

// type for available parameter(s)
type ParamTypeStr string

var (
	pageParam      ParamTypeStr = "PAGE"
	limitParam     ParamTypeStr = "LIMIT"
	totalDataParam ParamTypeStr = "TOTAL-PAGE"

	// Testing instance
	pageObj = page.Pagination{}

	// Test case for ValidatePagination
	validatePaginationTestCase = []struct {
		name          string
		argsPage      int
		argsLimit     int
		expectedPage  int
		expectedLimit int
	}{
		{
			name:          "Test Case 1",
			argsPage:      1,
			argsLimit:     10,
			expectedPage:  1,
			expectedLimit: 10,
		},
		{
			name:          "Test Case 2",
			argsPage:      12,
			argsLimit:     8,
			expectedPage:  12,
			expectedLimit: 8,
		},
		{
			name:          "Test Case 3",
			argsPage:      0,
			argsLimit:     10,
			expectedPage:  page.DefaultPage,
			expectedLimit: page.DefaultLimit,
		},
		{
			name:          "Test Case 4",
			argsPage:      2,
			argsLimit:     2000,
			expectedPage:  page.DefaultPage,
			expectedLimit: page.DefaultLimit,
		},
		{
			name:          "Test Case 5",
			argsPage:      -2,
			argsLimit:     1000,
			expectedPage:  page.DefaultPage,
			expectedLimit: page.DefaultLimit,
		},
		{
			name:          "Test Case 6",
			argsPage:      -50,
			argsLimit:     -1000,
			expectedPage:  page.DefaultPage,
			expectedLimit: page.DefaultLimit,
		},
	}

	// Test case for Paginate
	PaginateTestCase = []struct {
		name           string
		argsPage       int
		argsLimit      int
		expectedOffset int
	}{
		{
			name:           "Test Case 1",
			argsPage:       1,
			argsLimit:      15,
			expectedOffset: 0,
		},
		{
			name:           "Test Case 2",
			argsPage:       2,
			argsLimit:      15,
			expectedOffset: 15,
		},
		{
			name:           "Test Case 3",
			argsPage:       10,
			argsLimit:      20,
			expectedOffset: 180,
		},
		{
			name:           "Test Case 4",
			argsPage:       -5,
			argsLimit:      10,
			expectedOffset: page.DefaultLimit * (page.DefaultPage - 1),
		},
		{
			name:           "Test Case 5",
			argsPage:       8,
			argsLimit:      -100,
			expectedOffset: page.DefaultLimit * (page.DefaultPage - 1),
		},
		{
			name:           "Test Case 6",
			argsPage:       -50,
			argsLimit:      -50,
			expectedOffset: page.DefaultLimit * (page.DefaultPage - 1),
		},
		{
			name:           "Test Case 7",
			argsPage:       7,
			argsLimit:      1000,
			expectedOffset: 0,
		},
	}

	// Test case for SetTotalPage
	setTotalPageTestCase = []struct {
		name              string
		argsLimit         int
		argsTotalData     int
		expectedLimit     int
		expectedTotalPage int
	}{
		{
			name:              "Test Case 1",
			argsLimit:         500,
			argsTotalData:     10,
			expectedLimit:     10,
			expectedTotalPage: 1,
		},
		{
			name:              "Test Case 2",
			argsLimit:         10,
			argsTotalData:     15000,
			expectedLimit:     10,
			expectedTotalPage: 1500,
		},
		{
			name:              "Test Case 3",
			argsLimit:         70,
			argsTotalData:     1200,
			expectedLimit:     70,
			expectedTotalPage: 18,
		},
		{
			name:              "Test Case 4",
			argsLimit:         -100,
			argsTotalData:     8000,
			expectedLimit:     page.DefaultLimit,
			expectedTotalPage: 800,
		},
	}
)

// Generate random value based on given param
func generateRandomValueByParam(paramType ParamTypeStr) (int, error) {
	rand.Seed(time.Now().Unix())

	// check param type (page / limit / total data)
	switch paramType {
	case limitParam:
		return limitMinimum + rand.Intn(limitMaximum-limitMinimum+1), nil
	case pageParam:
		return pageMinimum + rand.Intn(pageMaximum-pageMinimum+1), nil
	case totalDataParam:
		return totalDataMinimum + rand.Intn(totalDataMaximum-totalDataMinimum+1), nil
	}

	return 0, errors.New("Unregistered benchmark parameter")
}

func TestValidatePagination(t *testing.T) {
	t.Log("Validate Pagination Test")
	for _, test := range validatePaginationTestCase {
		// Assign test case value to Pagination instance
		pageObj.Page = test.argsPage
		pageObj.Limit = test.argsLimit

		// Validate
		pageObj.ValidatePagination()

		// Assert result based on test case
		assert.Equal(t, pageObj.Page, test.expectedPage, fmt.Sprintf("Unexpected page value on %s", test.name))
		assert.Equal(t, pageObj.Limit, test.expectedLimit, fmt.Sprintf("Unexpected limit value on %s", test.name))
	}
}

func TestPaginate(t *testing.T) {
	t.Log("Paginate Test")
	for _, test := range PaginateTestCase {
		// Assign test case value to Pagination instance
		pageObj.Page = test.argsPage
		pageObj.Limit = test.argsLimit

		// Paginate
		pageObj.Paginate()

		// Assert result based on test case
		assert.Equal(t, pageObj.Offset, test.expectedOffset, fmt.Sprintf("Unexpected offset value on %s", test.name))
	}
}

func TestSetTotalPage(t *testing.T) {
	t.Log("Set Total Page Test")
	for _, test := range setTotalPageTestCase {
		// Assign test case value to Pagination instance
		pageObj.Limit = test.argsLimit
		pageObj.TotalData = test.argsTotalData

		// Set Total Page
		pageObj.SetTotalPage()

		// Assert result based on test case
		assert.Equal(t, pageObj.Limit, test.expectedLimit, fmt.Sprintf("Unexpected limit value on %s", test.name))
		assert.Equal(t, pageObj.TotalPage, test.expectedTotalPage, fmt.Sprintf("Unexpected total page value on %s", test.name))
	}
}

func BenchmarkPaginate(b *testing.B) {
	startTime := time.Now()
	for i := 0; i < benchmarkRep; i++ {
		// Fatalf will invoke failNow() and stop the testing process immediately
		tempPage, err := generateRandomValueByParam(pageParam)
		if err != nil {
			b.Fatalf("Error while generating random page for benchmarking : %s", err.Error())
		}

		tempLimit, err := generateRandomValueByParam(limitParam)
		if err != nil {
			b.Fatalf("Error while generating random limit for benchmarking : %s", err.Error())
		}

		pageObj.Page = tempPage
		pageObj.Limit = tempLimit
		pageObj.Paginate()
	}

	// Calculate benchmark duration
	totalDuration := time.Since(startTime)
	b.Logf("Benchmark completed in %v seconds", totalDuration.Seconds())
}
