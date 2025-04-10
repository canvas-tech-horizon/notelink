package notelink

import (
	"sort"
	"strconv"
	"strings"
)

// getVersion extracts the version from the path (e.g., "v1" from "/api/v1/users")
func getVersion(path string) string {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	for _, seg := range segments {
		if strings.HasPrefix(seg, "v") && len(seg) > 1 {
			return seg
		}
	}
	return "unknown" // Default if no version found
}

// getFullPath extracts the full path including parameters, normalized for grouping
func getFullPath(path string) string {
	return strings.Trim(path, "/")
}

// generateHTML creates documentation with progressive segment grouping and method grouping
func (an *ApiNote) generateHTML() string {
	var html strings.Builder

	html.WriteString(`<!DOCTYPE html>
<html>
<head>
    <link href="https://fonts.googleapis.com/css2?family=Source+Code+Pro:wght@400;600&display=swap" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0-beta3/css/all.min.css" rel="stylesheet">
    <title>` + an.config.Title + `</title>
    <style>
        body {
            font-family: 'Source Code Pro', monospace;
            margin: 20px;
            background-color: #f9f9f9;
            color: #333;
        }

        h1, h2, h4 {
            color: #2c3e50;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 0 15px;
            box-sizing: border-box;
        }

        .version-group, .top-segment-group {
            margin-bottom: 20px;
            padding: 10px 15px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.05);
        }

        .segment-group {
            margin-left: 20px;
            margin-bottom: 10px;
        }

        .path-group {
            margin-left: 40px;
            margin-bottom: 10px;
        }

        .method-group {
            margin-left: 60px;
            margin: 10px 0;
            padding: 10px 15px;
            border: 1px solid #eee;
            border-radius: 6px;
            background-color: #fefefe;
            box-shadow: 0 1px 3px rgba(0,0,0,0.03);
            box-sizing: border-box;
        }

        summary {
            cursor: pointer;
            padding: 8px;
        }

        .version-group > summary, .top-segment-group > summary {
            font-size: 1.3em;
            font-weight: bold;
            color: #2c3e50;
        }

        .segment-group > summary {
            font-size: 1.2em;
            font-weight: 600;
            color: #34495e;
        }

        .path-group > summary {
            font-size: 1.1em;
            font-weight: 500;
            color: #555;
        }

        .method-group > summary {
            font-size: 1em;
            font-weight: 500;
            color: #555;
        }

        .method {
            font-weight: bold;
        }

        .method.GET { color: #27ae60; }
        .method.POST { color: #2980b9; }
        .method.PUT { color: #f39c12; }
        .method.DELETE { color: #e74c3c; }
        .method.PATCH { color: #8e44ad; }

        .path {
            color: #888;
        }

        .responses, .schemas, .parameters, .api-test {
            margin-top: 10px;
            margin-left: 20px;
        }

        pre {
            background: #f4f4f4;
            padding: 10px;
            border-radius: 6px;
            overflow-x: auto;
            font-size: 0.9em;
        }

        .required {
            color: red;
        }

        .api-test input,
        .api-test textarea {
            width: 100%;
            max-width: 100%;
            box-sizing: border-box;
            padding: 8px;
            margin: 6px 0;
            border: 1px solid #ccc;
            border-radius: 4px;
        }

        .api-test button {
            background-color: #3498db;
            color: white;
            padding: 10px 15px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            transition: background 0.2s;
        }

        .api-test button:hover {
            background-color: #2980b9;
        }

        .auth-section {
            margin-bottom: 20px;
            padding: 10px 0;
        }

        .auth-section input {
            width: 60%;
            padding: 8px;
            margin-right: 10px;
            border: 1px solid #ccc;
            border-radius: 4px;
        }

        .auth-section button {
            padding: 8px 12px;
            background-color: #2ecc71;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
        }

        .auth-section button:hover {
            background-color: #27ae60;
        }

        .lock-icon {
            color: rgb(235, 202, 20);
            float: right;
            font-size: 1em;
            margin-left: 10px;
        }

        ul {
            padding-left: 20px;
        }

        @media (max-width: 600px) {
            .auth-section input {
                width: 100%;
                margin-bottom: 10px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
    <h1>` + an.config.Title + `</h1>
    <p>` + an.config.Description + `</p>
    <p>Version: ` + an.config.Version + `</p>
    <div class="auth-section">
        <h2>Authorize</h2>
        <input type="text" id="auth-token" placeholder="Enter JWT Bearer Token (e.g., Bearer eyJ...)" value="` + an.config.AuthToken + `">
        <button onclick="setAuthToken()">Set Token</button>
    </div>
    <h2>Endpoints</h2>`)

	// Build a nested structure: version (if exists) > top-level segment > sub-segments > full path > methods
	type SegmentNode struct {
		Name      string
		Children  map[string]*SegmentNode
		Endpoints []Endpoint
	}

	// Separate endpoints into versioned and non-versioned
	versionGroups := make(map[string]*SegmentNode)
	nonVersionedRoot := &SegmentNode{Name: "", Children: make(map[string]*SegmentNode)}

	for _, endpoint := range an.endpoints {
		version := getVersion(endpoint.Path)
		segments := strings.Split(strings.Trim(endpoint.Path, "/"), "/")
		var versionIdx int = -1
		for i, seg := range segments {
			if strings.HasPrefix(seg, "v") && len(seg) > 1 {
				versionIdx = i
				break
			}
		}

		if version != "unknown" {
			// Versioned endpoint
			if versionGroups[version] == nil {
				versionGroups[version] = &SegmentNode{Name: version, Children: make(map[string]*SegmentNode)}
			}
			current := versionGroups[version]

			// First segment after version
			if versionIdx+1 < len(segments) {
				topSeg := segments[versionIdx+1]
				if current.Children[topSeg] == nil {
					current.Children[topSeg] = &SegmentNode{Name: topSeg, Children: make(map[string]*SegmentNode)}
				}
				current = current.Children[topSeg]

				// Process deeper segments
				for i := versionIdx + 2; i < len(segments)-1; i++ {
					seg := segments[i]
					if current.Children[seg] == nil {
						current.Children[seg] = &SegmentNode{Name: seg, Children: make(map[string]*SegmentNode)}
					}
					current = current.Children[seg]
				}
				// Add endpoint at the deepest segment
				current.Endpoints = append(current.Endpoints, endpoint)
			}
		} else {
			// Non-versioned endpoint, group by first segment
			if len(segments) > 0 {
				topSeg := segments[0]
				if nonVersionedRoot.Children[topSeg] == nil {
					nonVersionedRoot.Children[topSeg] = &SegmentNode{Name: topSeg, Children: make(map[string]*SegmentNode)}
				}
				current := nonVersionedRoot.Children[topSeg]

				// Process deeper segments
				for i := 1; i < len(segments)-1; i++ {
					seg := segments[i]
					if current.Children[seg] == nil {
						current.Children[seg] = &SegmentNode{Name: seg, Children: make(map[string]*SegmentNode)}
					}
					current = current.Children[seg]
				}
				// Add endpoint at the deepest segment
				current.Endpoints = append(current.Endpoints, endpoint)
			}
		}
	}

	// Render segments recursively
	var renderSegments func(node *SegmentNode, depth int, groupClass string)
	renderSegments = func(node *SegmentNode, depth int, groupClass string) {
		// Sort children (segments)
		var segmentNames []string
		for name := range node.Children {
			segmentNames = append(segmentNames, name)
		}
		sort.Strings(segmentNames)

		for _, name := range segmentNames {
			child := node.Children[name]
			html.WriteString(`
        <details class="` + groupClass + `">
            <summary>` + name + `</summary>`)

			// Group endpoints by full path
			if len(child.Endpoints) > 0 {
				// Deduplicate by path
				pathGroups := make(map[string][]Endpoint)
				for _, endpoint := range child.Endpoints {
					fullPath := getFullPath(endpoint.Path)
					pathGroups[fullPath] = append(pathGroups[fullPath], endpoint)
				}

				// Sort full paths
				var fullPaths []string
				for fullPath := range pathGroups {
					fullPaths = append(fullPaths, fullPath)
				}
				sort.Strings(fullPaths)

				for _, fullPath := range fullPaths {
					endpoints := pathGroups[fullPath]
					sort.Slice(endpoints, func(i, j int) bool {
						return endpoints[i].Method < endpoints[j].Method
					})
					html.WriteString(`
            <details class="path-group">
                <summary>` + fullPath + ` (` + strconv.Itoa(len(endpoints)) + ` method` + pluralize(len(endpoints)) + `)</summary>`)

					// Render all methods under this path
					for _, endpoint := range endpoints {
						schemaBaseName := strings.Split(fullPath, "/")[len(strings.Split(fullPath, "/"))-1] // Second-to-last segment
						lockIcon := ""
						if endpoint.AuthRequired {
							lockIcon = `<i class="fas fa-lock lock-icon"></i>`
						}
						html.WriteString(`
                <details class="method-group">
                    <summary><span class="method ` + endpoint.Method + `">` + endpoint.Method + `</span> <b>` + endpoint.Path + `</b> <i>` + endpoint.Description + `</i>` + lockIcon + `</summary>
                    <div>`)

						if len(endpoint.Parameters) > 0 {
							html.WriteString(`
                    <div class="parameters">
                        <h4>Parameters:</h4>
                        <ul>`)
							for _, param := range endpoint.Parameters {
								required := ""
								if param.Required {
									required = `<span class="required"> (required)</span>`
								}
								html.WriteString(`
                        <li><strong>` + param.Name + `</strong> (` + param.In + `, ` + param.Type + `): ` + param.Description + required + `</li>`)
							}
							html.WriteString(`
                        </ul>
                    </div>`)
						}

						html.WriteString(`
                    <div class="responses">
                        <h4>Responses:</h4>`)

						var codes []string
						for code := range endpoint.Responses {
							codes = append(codes, code)
						}
						sort.Strings(codes)

						for _, code := range codes {
							html.WriteString(`
                        <p>` + code + `: ` + endpoint.Responses[code] + `</p>`)
						}

						html.WriteString(`
                    </div>
                    <div class="schemas">
                        <h4>Schemas:</h4>`)

						if reqTs := generateTypeScriptSchema(schemaBaseName+"Request", endpoint.RequestSchema); reqTs != "" {
							html.WriteString(`
                        <h5>Request Body:</h5>
                        <pre>` + reqTs + `</pre>`)
						}
						if respTs := generateTypeScriptSchema(schemaBaseName+"Response", endpoint.ResponseSchema); respTs != "" {
							html.WriteString(`
                        <h5>Response Body:</h5>
                        <pre>` + respTs + `</pre>`)
						}

						html.WriteString(`
                    </div>
                    <div class="api-test">
                        <h4>Test API</h4>
                        <form id="test-form-` + endpoint.Method + strings.ReplaceAll(endpoint.Path, "/", "-") + `" onsubmit="testApi(event, '` + endpoint.Method + `', '` + endpoint.Path + `', this)" enctype="multipart/form-data">
                            <input type="hidden" name="method" value="` + endpoint.Method + `">`)

						hasFormData := false
						for _, param := range endpoint.Parameters {
							inputType := "text"
							if param.Type == "number" {
								inputType = "number"
							} else if param.Type == "file" {
								inputType = "file"
								hasFormData = true
							}
							requiredAttr := ""
							if param.Required {
								requiredAttr = " required"
							}
							labelText := param.Name + ` (` + param.In + `)`
							if param.Required {
								labelText += ` <span class="required">* required</span>`
							}
							html.WriteString(`
                            <label>` + labelText + `:</label>
                            <input type="` + inputType + `" name="` + param.Name + `" placeholder="Enter ` + param.Name + `"` + requiredAttr + ` data-in="` + param.In + `">`)
						}

						if endpoint.Method == "POST" || endpoint.Method == "PUT" {
							if !hasFormData {
								html.WriteString(`
                            <label>Request Body (JSON):</label>
                            <textarea rows="5" name="requestBody" placeholder="Enter JSON request body"></textarea>`)
							}
						}

						html.WriteString(`
                            <button type="submit">Test Request</button>
                            <pre id="test-result-` + endpoint.Method + strings.ReplaceAll(endpoint.Path, "/", "-") + `"></pre>
                        </form>
                    </div>
                </div>
            </details>`)
					}
					html.WriteString(`
            </details>`)
				}
			}

			// Recurse into deeper segments
			renderSegments(child, depth+1, "segment-group")
			html.WriteString(`
        </details>`)
		}
	}

	// Render versioned groups
	var versions []string
	for version := range versionGroups {
		versions = append(versions, version)
	}
	sort.Strings(versions)

	for _, version := range versions {
		node := versionGroups[version]
		html.WriteString(`
    <details class="version-group">
        <summary>` + version + `</summary>`)
		renderSegments(node, 1, "segment-group")
		html.WriteString(`
    </details>`)
	}

	// Render non-versioned groups (directly under top-level segments)
	if len(nonVersionedRoot.Children) > 0 {
		renderSegments(nonVersionedRoot, 0, "top-segment-group")
	}

	html.WriteString(`
    <script>
        let authToken = '` + an.config.AuthToken + `';

        if (!authToken) {
            const storedToken = localStorage.getItem('authToken');
            if (storedToken) {
                authToken = storedToken;
            }
        }

        // Set the input field value on page load
        window.onload = function() {
            const authInput = document.getElementById('auth-token');
            if (authInput) {
                authInput.value = authToken;
            }
        };

        function setAuthToken() {
            const authInput = document.getElementById('auth-token');
            authToken = authInput.value;
            localStorage.setItem('authToken', authToken); // Save to localStorage
            alert('Authorization token set: ' + (authToken ? authToken : 'None'));
        }

        function testApi(event, method, path, form) {
            event.preventDefault();
            const resultElement = document.getElementById('test-result-' + method + path.replace(/\//g, '-'));
            resultElement.textContent = 'Sending request...';

            const params = {};
            const queryParams = new URLSearchParams();
            const formData = new FormData(form);
            let hasFormData = false;

            formData.forEach((value, key) => {
                if (key !== 'method' && key !== 'requestBody') {
                    const input = form.querySelector('input[name="' + key + '"]');
                    const paramIn = input.getAttribute('data-in');
                    if (paramIn === 'path') {
                        path = path.replace(':' + key, value);
                    } else if (paramIn === 'query' && value) {
                        queryParams.append(key, value);
                    } else if (paramIn === 'header' && value) {
                        params[key] = value;
                    }
                    if (input.type === 'file' && value) {
                        hasFormData = true;
                    }
                }
            });

            const baseUrl = 'http://` + an.config.Host + `';
            const url = baseUrl + path + (queryParams.toString() ? '?' + queryParams.toString() : '');

            const options = {
                method: method,
                headers: {},
            };

            if (authToken) {
                options.headers['Authorization'] = authToken;
            }

            Object.keys(params).forEach(key => {
                if (params[key]) options.headers[key] = params[key];
            });

            if (hasFormData) {
                options.body = formData;
            } else {
                const requestBody = formData.get('requestBody');
                if (requestBody && (method === 'POST' || method === 'PUT')) {
                    try {
                        options.headers['Content-Type'] = 'application/json';
                        options.body = requestBody;
                    } catch (e) {
                        resultElement.textContent = 'Invalid JSON in request body: ' + e.message;
                        return;
                    }
                }
            }

            fetch(url, options)
                .then(response => {
                    const contentType = response.headers.get('content-type') || '';
                    const disposition = response.headers.get('content-disposition') || '';
                    let filename = 'download';
                    if (disposition) {
                        const matches = disposition.match(/filename="([^"]+)"/);
                        if (matches && matches[1]) filename = matches[1];
                    }

                    if (!response.ok) {
                        return response.text().then(text => ({
                            status: response.status,
                            statusText: response.statusText,
                            body: text,
                            contentType: contentType,
                            isError: true
                        }));
                    } else if (contentType.includes('application/json')) {
                        return response.json().then(data => ({
                            status: response.status,
                            statusText: response.statusText,
                            body: JSON.stringify(data, null, 2),
                            contentType: contentType
                        }));
                    } else if (contentType.startsWith('image/')) {
                        return response.blob().then(blob => ({
                            status: response.status,
                            statusText: response.statusText,
                            body: blob,
                            contentType: contentType,
                            isImage: true,
                            filename: filename
                        }));
                    } else {
                        return response.blob().then(blob => ({
                            status: response.status,
                            statusText: response.statusText,
                            body: blob,
                            contentType: contentType,
                            isBlob: true,
                            filename: filename
                        }));
                    }
                })
                .then(result => {
                    resultElement.innerHTML = "Url: " + url + "<br>Status: " + result.status + " " + result.statusText + "<br><br>";
                    if (result.isError) {
                        resultElement.innerHTML += '<strong>Error Response:</strong><br><pre>' + result.body + '</pre>';
                    } else if (result.isImage) {
                        const imgUrl = URL.createObjectURL(result.body);
                        resultElement.innerHTML += '<strong>Response (Image):</strong><br><img src="' + imgUrl + '" style="max-width: 100%;" onload="setTimeout(() => URL.revokeObjectURL(this.src), 1000)">';
                    } else if (result.isBlob) {
                        const blobUrl = URL.createObjectURL(result.body);
                        resultElement.innerHTML += '<strong>Response (File):</strong><br><a href="' + blobUrl + '" download="' + result.filename + '">' + result.filename + '</a>';
                    } else {
                        resultElement.innerHTML += '<strong>Response:</strong><br><pre>' + result.body + '</pre>';
                    }
                })
                .catch(error => {
                    resultElement.textContent = 'Error: ' + error.message;
                });
        }
    </script>
    </div>
</body>
</html>`)

	return html.String()
}

// pluralize returns "s" if count > 1, empty string otherwise
func pluralize(count int) string {
	if count > 1 {
		return "s"
	}
	return ""
}
