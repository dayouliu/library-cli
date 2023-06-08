package tests

import (
	"bms/client/cmd"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// compareJSON compares two JSON strings and returns true if they are equal
func compareJSON(str1, str2 string) bool {
	var data1, data2 interface{}

	err := json.Unmarshal([]byte(str1), &data1)
	if err != nil {
		fmt.Println(err)
		return false
	}

	err = json.Unmarshal([]byte(str2), &data2)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// Compare the parsed JSON data using DeepEqual
	equal := reflect.DeepEqual(data1, data2)

	return equal
}

// TestCommands tests the cobra commands for bms cli client
func TestCommands(t *testing.T) {
	mockBookListData, _ := readJsonFile("resources/mock_books.json")

	// table driven tests
	testCases := []struct {
		name               string
		args               []string
		flags              map[string]string
		cmdHandler         func(cmd *cobra.Command, args []string) string
		expectedStatusCode int
		expectedOutput     string
	}{
		{
			name:               "Valid create book",
			args:               []string{"book", "create", "book1"},
			flags:              map[string]string{},
			expectedStatusCode: http.StatusCreated,
			expectedOutput:     "Book created successfully\n",
		},
		{
			name:               "Invalid create book",
			args:               []string{"book", "create", ""},
			flags:              map[string]string{},
			expectedStatusCode: http.StatusBadRequest,
			expectedOutput:     "Error: Title cannot be empty\n",
		},
		{
			name:               "List books",
			args:               []string{"book", "list"},
			flags:              map[string]string{},
			expectedStatusCode: http.StatusOK,
			expectedOutput:     string(mockBookListData),
		},
		// Add more tests for each command as necessary
	}

	// Create a new mock server
	server := httptest.NewServer(http.HandlerFunc(mockRouter))
	defer server.Close()
	cmd.ServerUrl = server.URL

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a buffer to capture the output
			buf := new(bytes.Buffer)
			cmd.RootCmd.SetOut(buf)

			// set flags and args
			for k, v := range tc.flags {
				err := cmd.RootCmd.Flags().Set(k, v)
				if err != nil {
					t.Errorf("Error setting flag %v: %v", k, err)
				}
			}
			cmd.RootCmd.SetArgs(tc.args)

			// Execute the command
			err := cmd.RootCmd.Execute()
			if err != nil {
				t.Errorf("Error executing add command: %v", err)
			}

			// Get the captured output and compare
			cmdOutput := buf.String()
			if cmdOutput != tc.expectedOutput && !compareJSON(cmdOutput, tc.expectedOutput) {
				t.Errorf("Expected body %v, but got %v", tc.expectedOutput, cmdOutput)
			}
		})
	}
}
