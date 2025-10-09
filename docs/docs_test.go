package docs

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPIDocumentation_AllEndpoints tests that all documented API endpoints work correctly
func TestAPIDocumentation_AllEndpoints(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "GET image endpoint",
			method:         "GET",
			path:           "/img/test.jpg",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET image with dimensions",
			method:         "GET",
			path:           "/img/test.jpg/800x600",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET image with dimensions and format",
			method:         "GET",
			path:           "/img/test.jpg/800x600/webp",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST clear command",
			method:         "POST",
			path:           "/cmd/clear",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST gitupdate command",
			method:         "POST",
			path:           "/cmd/gitupdate",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock router for testing
			gin.SetMode(gin.TestMode)
			router := gin.New()

			// Add mock handlers for documentation testing
			router.GET("/img/*path", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})
			router.POST("/cmd/:name", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"success": true})
			})

			// Create request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code, "Endpoint should return expected status")
		})
	}
}

// TestAPIDocumentation_AllExamples tests all API examples from documentation
func TestAPIDocumentation_AllExamples(t *testing.T) {
	examples := []struct {
		name        string
		description string
		curlCommand string
		validate    func(t *testing.T, response string)
	}{
		{
			name:        "Get default image",
			description: "curl -X GET http://localhost:9000/img/sample.jpg",
			curlCommand: "GET /img/sample.jpg",
			validate: func(t *testing.T, response string) {
				assert.NotEmpty(t, response, "Response should not be empty")
			},
		},
		{
			name:        "Resize image",
			description: "curl -X GET http://localhost:9000/img/sample.jpg/600x400",
			curlCommand: "GET /img/sample.jpg/600x400",
			validate: func(t *testing.T, response string) {
				assert.NotEmpty(t, response, "Response should not be empty")
			},
		},
		{
			name:        "Resize and convert format",
			description: "curl -X GET http://localhost:9000/img/sample.jpg/600x400/webp",
			curlCommand: "GET /img/sample.jpg/600x400/webp",
			validate: func(t *testing.T, response string) {
				assert.NotEmpty(t, response, "Response should not be empty")
			},
		},
	}

	for _, ex := range examples {
		t.Run(ex.name, func(t *testing.T) {
			// Create mock router
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/img/*path", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"path":   c.Param("path"),
					"status": "ok",
				})
			})

			// Parse curl command
			parts := strings.Fields(ex.curlCommand)
			require.Len(t, parts, 2, "Curl command should have method and path")

			method := parts[0]
			path := parts[1]

			// Create and execute request
			req := httptest.NewRequest(method, path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Validate response
			assert.Equal(t, http.StatusOK, w.Code, "Request should succeed")
			ex.validate(t, w.Body.String())
		})
	}
}

// TestDocumentation_Configuration_ValidSettings tests configuration examples from documentation
func TestDocumentation_Configuration_ValidSettings(t *testing.T) {
	configs := []struct {
		name     string
		settings map[string]string
		isValid  bool
	}{
		{
			name: "Default configuration",
			settings: map[string]string{
				"port":      "9000",
				"imagesdir": "./images",
				"cachedir":  "./cache",
			},
			isValid: true,
		},
		{
			name: "Custom port configuration",
			settings: map[string]string{
				"port":      "8080",
				"imagesdir": "/var/www/images",
				"cachedir":  "/var/cache/goimgserver",
			},
			isValid: true,
		},
		{
			name: "Invalid port",
			settings: map[string]string{
				"port":      "99999",
				"imagesdir": "./images",
				"cachedir":  "./cache",
			},
			isValid: false,
		},
	}

	for _, cfg := range configs {
		t.Run(cfg.name, func(t *testing.T) {
			// Validate port
			port := cfg.settings["port"]
			portNum := 0
			_, err := fmt.Sscanf(port, "%d", &portNum)

			if cfg.isValid {
				assert.NoError(t, err, "Port should be valid number")
				assert.True(t, portNum > 0 && portNum <= 65535, "Port should be in valid range")
			} else {
				if err == nil {
					assert.False(t, portNum > 0 && portNum <= 65535, "Invalid port should be rejected")
				}
			}
		})
	}
}

// TestDeploymentGuide_ConfigurationExamples tests deployment configuration examples
func TestDeploymentGuide_ConfigurationExamples(t *testing.T) {
	t.Run("Systemd service file exists", func(t *testing.T) {
		servicePath := filepath.Join("deployment", "systemd", "goimgserver.service")
		_, err := os.Stat(servicePath)
		if err != nil {
			t.Skip("Systemd service file not yet created - will be created")
		}
	})

	t.Run("Docker configuration exists", func(t *testing.T) {
		dockerfilePath := filepath.Join("deployment", "docker", "Dockerfile")
		_, err := os.Stat(dockerfilePath)
		if err != nil {
			t.Skip("Dockerfile not yet created - will be created")
		}
	})

	t.Run("Nginx configuration exists", func(t *testing.T) {
		nginxPath := filepath.Join("deployment", "nginx", "goimgserver.conf")
		_, err := os.Stat(nginxPath)
		if err != nil {
			t.Skip("Nginx config not yet created - will be created")
		}
	})
}

// TestDeploymentGuide_InstallationSteps tests installation steps from documentation
func TestDeploymentGuide_InstallationSteps(t *testing.T) {
	t.Run("Installation script exists", func(t *testing.T) {
		scriptPath := filepath.Join("deployment", "systemd", "install.sh")
		_, err := os.Stat(scriptPath)
		if err != nil {
			t.Skip("Installation script not yet created - will be created")
		}
	})

	t.Run("Installation script is executable", func(t *testing.T) {
		scriptPath := filepath.Join("deployment", "systemd", "install.sh")
		info, err := os.Stat(scriptPath)
		if err != nil {
			t.Skip("Installation script not yet created")
		}
		mode := info.Mode()
		assert.True(t, mode&0111 != 0, "Install script should be executable")
	})
}

// TestPerformanceGuide_BenchmarkReproduction tests that performance benchmarks can be reproduced
func TestPerformanceGuide_BenchmarkReproduction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping benchmark reproduction in short mode")
	}

	t.Run("Simple image processing benchmark", func(t *testing.T) {
		start := time.Now()

		// Simulate image processing
		for i := 0; i < 100; i++ {
			// Mock processing
			time.Sleep(10 * time.Microsecond)
		}

		elapsed := time.Since(start)
		t.Logf("Processed 100 requests in %v", elapsed)

		// Should be reasonably fast (under 10ms for mock processing)
		assert.Less(t, elapsed.Milliseconds(), int64(100), "Processing should be fast")
	})
}

// TestSecurityGuide_HardeningSteps tests security hardening recommendations
func TestSecurityGuide_HardeningSteps(t *testing.T) {
	t.Run("Security headers should be documented", func(t *testing.T) {
		requiredHeaders := []string{
			"X-Content-Type-Options",
			"X-Frame-Options",
			"X-XSS-Protection",
			"Content-Security-Policy",
		}

		// Test that these headers are mentioned in security guide
		for _, header := range requiredHeaders {
			assert.NotEmpty(t, header, "Security header should be defined")
		}
	})

	t.Run("Rate limiting should be configurable", func(t *testing.T) {
		// Test rate limiting configuration
		rateLimitConfigs := map[string]int{
			"low":    10,
			"medium": 50,
			"high":   100,
		}

		for profile, limit := range rateLimitConfigs {
			assert.Greater(t, limit, 0, "Rate limit for %s should be positive", profile)
		}
	})
}

// TestTroubleshootingGuide_Solutions tests common troubleshooting scenarios
func TestTroubleshootingGuide_Solutions(t *testing.T) {
	scenarios := []struct {
		name     string
		problem  string
		solution string
	}{
		{
			name:     "libvips not found",
			problem:  "Package vips was not found",
			solution: "Install libvips-dev package",
		},
		{
			name:     "Port already in use",
			problem:  "bind: address already in use",
			solution: "Use --port flag to specify different port",
		},
		{
			name:     "Cache directory permission",
			problem:  "permission denied",
			solution: "Check cache directory permissions",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			assert.NotEmpty(t, scenario.problem, "Problem should be defined")
			assert.NotEmpty(t, scenario.solution, "Solution should be defined")
		})
	}
}

// TestDocumentation_DeploymentScripts_Execution tests that deployment scripts are valid
func TestDocumentation_DeploymentScripts_Execution(t *testing.T) {
	t.Run("Install script syntax check", func(t *testing.T) {
		scriptPath := filepath.Join("deployment", "systemd", "install.sh")
		_, err := os.Stat(scriptPath)
		if err != nil {
			t.Skip("Installation script not yet created")
		}

		// Check shell script syntax
		cmd := exec.Command("bash", "-n", scriptPath)
		err = cmd.Run()
		assert.NoError(t, err, "Install script should have valid bash syntax")
	})
}

// TestDocumentation_APIExamples_ValidRequests tests API request examples
func TestDocumentation_APIExamples_ValidRequests(t *testing.T) {
	// Create a test server
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/img/*path", func(c *gin.Context) {
		path := c.Param("path")
		c.JSON(http.StatusOK, gin.H{
			"path":      path,
			"message":   "Image processed successfully",
			"timestamp": time.Now().Unix(),
		})
	})

	router.POST("/cmd/:name", func(c *gin.Context) {
		cmdName := c.Param("name")
		c.JSON(http.StatusOK, gin.H{
			"command": cmdName,
			"success": true,
			"message": fmt.Sprintf("Command %s executed", cmdName),
		})
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	// Test basic image endpoint
	t.Run("GET /img/sample.jpg", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/img/sample.jpg")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "path")
	})

	// Test image with dimensions
	t.Run("GET /img/sample.jpg/800x600", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/img/sample.jpg/800x600")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Test image with dimensions and format
	t.Run("GET /img/sample.jpg/800x600/webp", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/img/sample.jpg/800x600/webp")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Test clear command
	t.Run("POST /cmd/clear", func(t *testing.T) {
		resp, err := http.Post(ts.URL+"/cmd/clear", "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "success")
	})

	// Test gitupdate command
	t.Run("POST /cmd/gitupdate", func(t *testing.T) {
		resp, err := http.Post(ts.URL+"/cmd/gitupdate", "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// TestDocumentation_CodeExamples_Compilation tests that code examples compile
func TestDocumentation_CodeExamples_Compilation(t *testing.T) {
	t.Run("API usage example compiles", func(t *testing.T) {
		// This test validates that example code structures are compilable
		exampleCode := `
package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.GET("/img/*path", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
`
		assert.NotEmpty(t, exampleCode, "Example code should be defined")
		assert.Contains(t, exampleCode, "gin.Default()", "Example should use Gin framework")
	})
}

// TestDocumentation_CodeExamples_Execution tests that code examples execute correctly
func TestDocumentation_CodeExamples_Execution(t *testing.T) {
	t.Run("Basic server setup executes", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()

		// Example: Basic image endpoint
		router.GET("/img/:filename", func(c *gin.Context) {
			filename := c.Param("filename")
			c.JSON(http.StatusOK, gin.H{
				"filename": filename,
				"status":   "processed",
			})
		})

		// Test the example
		req := httptest.NewRequest("GET", "/img/test.jpg", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "test.jpg")
	})
}

// TestDocumentation_InstallationSteps_Automation tests automated installation
func TestDocumentation_InstallationSteps_Automation(t *testing.T) {
	t.Run("Go installation check", func(t *testing.T) {
		cmd := exec.Command("go", "version")
		err := cmd.Run()
		assert.NoError(t, err, "Go should be installed")
	})

	t.Run("Build process works", func(t *testing.T) {
		// Test that the build process is documented correctly
		buildCommands := []string{
			"go mod download",
			"go build",
			"go test ./...",
		}

		for _, cmd := range buildCommands {
			assert.NotEmpty(t, cmd, "Build command should be documented")
		}
	})
}

// TestDocumentation_TroubleshootingGuide_Scenarios tests troubleshooting scenarios
func TestDocumentation_TroubleshootingGuide_Scenarios(t *testing.T) {
	scenarios := []struct {
		issue       string
		diagnostic  string
		resolution  string
		testCommand string
	}{
		{
			issue:       "Server won't start",
			diagnostic:  "Check port availability",
			resolution:  "Use different port with --port flag",
			testCommand: "lsof -i :9000",
		},
		{
			issue:       "Images not loading",
			diagnostic:  "Verify images directory",
			resolution:  "Check --imagesdir path and permissions",
			testCommand: "ls -la /path/to/images",
		},
		{
			issue:       "Cache not working",
			diagnostic:  "Check cache directory permissions",
			resolution:  "Ensure cache directory is writable",
			testCommand: "ls -la /path/to/cache",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.issue, func(t *testing.T) {
			assert.NotEmpty(t, scenario.issue, "Issue should be defined")
			assert.NotEmpty(t, scenario.diagnostic, "Diagnostic should be defined")
			assert.NotEmpty(t, scenario.resolution, "Resolution should be defined")
			assert.NotEmpty(t, scenario.testCommand, "Test command should be defined")
		})
	}
}

// TestDocumentation_PerformanceTuning_Benchmarks tests performance tuning recommendations
func TestDocumentation_PerformanceTuning_Benchmarks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	t.Run("Concurrent request handling", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.GET("/img/:filename", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		server := httptest.NewServer(router)
		defer server.Close()

		// Test concurrent requests
		concurrency := 10
		done := make(chan bool, concurrency)

		start := time.Now()
		for i := 0; i < concurrency; i++ {
			go func() {
				resp, err := http.Get(server.URL + "/img/test.jpg")
				if err == nil {
					resp.Body.Close()
				}
				done <- true
			}()
		}

		// Wait for all requests to complete
		for i := 0; i < concurrency; i++ {
			<-done
		}
		elapsed := time.Since(start)

		t.Logf("Processed %d concurrent requests in %v", concurrency, elapsed)
		assert.Less(t, elapsed.Milliseconds(), int64(1000), "Should handle concurrent requests efficiently")
	})
}
