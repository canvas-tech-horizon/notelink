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
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link rel="icon" type="image/png" sizes="32x32" href="/icon.png">
    <link rel="icon" type="image/png" sizes="16x16" href="/icon.png">
    <link rel="shortcut icon" href="/icon.png">
    <link rel="apple-touch-icon" href="/icon.png">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <title>` + an.config.Title + `</title>
    <style>
        :root {
            --primary: #e9902bff;
            --primary-dark: #e59346ff;
            --success: #10b981;
            --warning: #f59e0b;
            --danger: #ef4444;
            --info: #3b82f6;
            --secondary: #e7a04eff;
            --gray-50: #f9fafb;
            --gray-100: #f3f4f6;
            --gray-200: #e5e7eb;
            --gray-300: #d1d5db;
            --gray-400: #9ca3af;
            --gray-500: #6b7280;
            --gray-600: #4b5563;
            --gray-700: #374151;
            --gray-800: #1f2937;
            --gray-900: #111827;
            --white: #ffffff;
            --radius: 0.75rem;
            --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
            --shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1);
            --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
        }

        * {
            box-sizing: border-box;
        }

        body {
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 0;
            background: linear-gradient(135deg, var(--gray-50) 0%, var(--gray-100) 100%);
            color: var(--gray-800);
            min-height: 100vh;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
            padding: 1rem 2rem;
        }

        .header {
            text-align: center;
            margin-bottom: 1.5rem;
            padding: 1rem 0;
        }

        h1 {
            font-size: 1.75rem;
            font-weight: 700;
            color: var(--gray-900);
            margin: 0 0 0.25rem 0;
            background: linear-gradient(135deg, var(--primary) 0%, var(--secondary) 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }

        .subtitle {
            font-size: 0.875rem;
            color: var(--gray-600);
            margin: 0 0 0.25rem 0;
            max-width: 600px;
            margin-left: auto;
            margin-right: auto;
        }

        .version-badge {
            display: inline-block;
            background: var(--primary);
            color: var(--white);
            padding: 0.15rem 0.5rem;
            border-radius: 9999px;
            font-size: 0.75rem;
            font-weight: 500;
            margin-top: 0.25rem;
        }

        .auth-section {
            background: var(--white);
            border-radius: var(--radius);
            padding: 1rem;
            margin-bottom: 1.5rem;
            box-shadow: var(--shadow);
            border: 1px solid var(--gray-200);
        }

        .auth-section h2 {
            font-size: 1rem;
            font-weight: 600;
            color: var(--gray-900);
            margin: 0 0 0.75rem 0;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .auth-input-group {
            display: flex;
            gap: 0.75rem;
            align-items: stretch;
        }

        .auth-section input {
            flex: 1;
            padding: 0.75rem 1rem;
            border: 1px solid var(--gray-300);
            border-radius: var(--radius);
            font-size: 0.875rem;
            transition: all 0.2s ease;
            background: var(--white);
        }

        .auth-section input:focus {
            outline: none;
            border-color: var(--primary);
            box-shadow: 0 0 0 3px rgb(99 102 241 / 0.1);
        }

        .auth-section button {
            padding: 0.5rem 1rem;
            background: var(--primary);
            color: var(--white);
            border: none;
            border-radius: var(--radius);
            font-weight: 500;
            cursor: pointer;
            transition: all 0.2s ease;
            font-size: 0.875rem;
        }

        .auth-section button:hover {
            background: var(--primary-dark);
            transform: translateY(-1px);
        }

        .monitor-section {
            background: var(--white);
            border-radius: var(--radius);
            padding: 1rem;
            margin-bottom: 1.5rem;
            box-shadow: var(--shadow);
            border: 1px solid var(--gray-200);
            text-align: center;
        }

        .section-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .monitor-button {
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            padding: 0.5rem 1rem;
            background: var(--info);
            color: var(--white);
            border: none;
            border-radius: var(--radius);
            font-weight: 500;
            cursor: pointer;
            transition: all 0.2s ease;
            font-size: 0.75rem;
            text-decoration: none;
        }

        .monitor-button:hover {
            background: #2563eb;
            transform: translateY(-1px);
            box-shadow: var(--shadow-lg);
        }

        .monitor-button i {
            font-size: 1rem;
        }

        .section-title {
            font-size: 1.25rem;
            font-weight: 600;
            color: var(--gray-900);
        }

        .version-group, .top-segment-group {
            background: transparent;
            border: none;
            margin-bottom: 1rem;
            padding-bottom: 0.5rem;
            transition: all 0.3s ease;
        }

        .version-group:hover, .top-segment-group:hover {
            border-bottom-color: var(--primary);
        }

        .segment-group {
            margin: 0.25rem 0;
            background: transparent;
            border: none;
            border-left: 3px solid var(--gray-200);
            transition: all 0.3s ease;
        }

        .segment-group:hover {
            border-left-color: var(--primary);
        }

        .path-group {
            margin: 0.25rem 0;
            background: transparent;
            border: none;
            padding-left: 1rem;
            transition: all 0.3s ease;
        }

        .method-group {
            margin: 0.5rem 0;
            padding-left: 1rem;
            transition: all 0.3s ease;
        }

        .method-group:hover {
            border-left-color: var(--primary);
        }

        /* Beautiful collapse animations */
        details {
            position: relative;
        }

        details > *:not(summary) {
            animation: collapse-open 0.3s ease-out;
            transform-origin: top;
        }

        @keyframes collapse-open {
            0% {
                opacity: 0;
                transform: scaleY(0.8) translateY(-10px);
            }
            100% {
                opacity: 1;
                transform: scaleY(1) translateY(0);
            }
        }

        summary {
            cursor: pointer;
            padding: 0.75rem 1rem;
            font-weight: 500;
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            list-style: none;
            position: relative;
            display: flex;
            align-items: center;
            user-select: none;
            border-radius: inherit;
        }

        summary::-webkit-details-marker {
            display: none;
        }

        /* Modern chevron design */
        summary::before {
            content: '';
            width: 6px;
            height: 6px;
            border-right: 2px solid var(--gray-500);
            border-bottom: 2px solid var(--gray-500);
            transform: rotate(-45deg);
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            margin-right: 0.75rem;
            flex-shrink: 0;
        }

        details[open] > summary::before {
            transform: rotate(45deg);
            border-color: var(--primary);
        }

        summary:hover {
            background: var(--gray-50);
            border-radius: 0.5rem;
        }

        summary:hover::before {
            border-color: var(--primary);
            transform: scale(1.1) rotate(-45deg);
        }

        details[open] > summary:hover::before {
            transform: scale(1.1) rotate(45deg);
        }

        /* Enhanced styling for different levels */
        .version-group > summary, .top-segment-group > summary {
            font-size: 1.25rem;
            font-weight: 700;
            color: var(--primary);
            background: transparent;
            padding: 0.5rem 0;
            border-bottom: none;
        }

        .version-group > summary::before, .top-segment-group > summary::before {
            border-color: var(--primary);
        }

        .version-group > summary:hover, .top-segment-group > summary:hover {
            background: var(--gray-50);
            color: var(--primary-dark);
        }

        .segment-group > summary {
            font-size: 1rem;
            font-weight: 600;
            color: var(--gray-800);
            background: transparent;
            padding: 0.5rem 0 0.5rem 1rem;
        }

        .segment-group > summary:hover {
            background: var(--gray-50);
            color: var(--primary);
            border-radius: 0.5rem;
        }

        .method-group > summary {
            font-weight: 500;
            background: transparent;
            padding: 0.5rem 0 0.5rem 1rem;
        }

        .method-group > summary:hover {
            background: var(--gray-50);
            border-radius: 0.5rem;
        }

        .path-group > summary {
            font-size: 0.9rem;
            font-weight: 500;
            color: var(--gray-700);
            background: transparent;
            padding: 0.4rem 0 0.4rem 1rem;
        }

        .path-group > summary:hover {
            background: var(--gray-50);
            color: var(--info);
            border-radius: 0.5rem;
        }

        /* Content styling with better spacing */
        details[open] > summary + * {
            background: transparent;
            border-top: none;
        }

        .version-group[open] > summary + *,
        .top-segment-group[open] > summary + * {
            background: transparent;
        }

        .method-group[open] > summary + * {
            padding: 1rem 0.75rem;
            background: var(--gray-50);
            border-radius: 0.5rem;
            margin-top: 0.5rem;
        }

        /* Badge indicators for open/closed state */
        summary::after {
            position: absolute;
            right: 1.5rem;
            width: 6px;
            height: 6px;
            background: var(--gray-300);
            border-radius: 50%;
            transition: all 0.3s ease;
        }

        details[open] > summary::after {
            background: var(--success);
            transform: scale(1.3);
        }

        .segment-group > summary::after,
        .path-group > summary::after {
            display: none;
        }

        .version-group > summary::after,
        .top-segment-group > summary::after {
            background: rgba(255, 255, 255, 0.5);
        }

        .version-group[open] > summary::after,
        .top-segment-group[open] > summary::after {
            background: var(--white);
        }

        .method {
            display: inline-flex;
            align-items: center;
            font-weight: 600;
            font-size: 0.75rem;
            padding: 0.5rem 0.75rem;
            border-radius: 9999px;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            margin-right: 1rem;
            min-width: 70px;
            justify-content: center;
            position: relative;
        }

        .method.GET {
            background: linear-gradient(135deg, #10b981 0%, #059669 100%);
            color: var(--white);
        }

        .method.POST {
            background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
            color: var(--white);
        }

        .method.PUT {
            background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
            color: var(--white);
        }

        .method.DELETE {
            background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
            color: var(--white);
        }

        .method.PATCH {
            background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
            color: var(--white);
        }

        .endpoint-path {
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.9rem;
            color: var(--gray-700);
            font-weight: 500;
            background: var(--gray-100);
            padding: 0.375rem 0.75rem;
            border-radius: 0.5rem;
            border: 1px solid var(--gray-200);
            transition: all 0.3s ease;
        }

        .method-group:hover .endpoint-path {
            background: var(--primary);
            color: var(--white);
            border-color: var(--primary);
            transform: translateX(5px);
        }

        .endpoint-description {
            color: var(--gray-500);
            font-style: italic;
            font-size: 0.875rem;
            font-weight: 400;
            margin-left: auto;
            opacity: 0.8;
            transition: all 0.3s ease;
            padding-right: 28px;
        }

        .method-group:hover .endpoint-description {
            color: var(--gray-700);
            opacity: 1;
        }

        .responses, .schemas, .parameters {
            margin: 0.75rem 0;
            padding: 0.5rem 0;
            border-bottom: 1px solid var(--gray-200);
        }

        .api-test {
            margin: 1rem 0;
            padding: 1rem;
            background: var(--white);
            border-radius: var(--radius);
            border: 1px solid var(--gray-200);
            box-shadow: var(--shadow-sm);
        }

        h4, h5 {
            font-size: 0.9rem;
            font-weight: 600;
            color: var(--gray-900);
            margin: 0 0 0.5rem 0;
        }

        h5 {
            font-size: 0.85rem;
            color: var(--gray-700);
        }

        pre {
            background: var(--gray-900);
            color: var(--gray-100);
            padding: 1rem;
            border-radius: var(--radius);
            overflow-x: auto;
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.8rem;
            line-height: 1.5;
        }

        .required {
            color: var(--danger);
            font-weight: 500;
        }

        .api-test h4 {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            color: var(--primary);
        }

        .api-test h4::before {
            content: '🚀';
            font-size: 1.2rem;
        }

        .api-test input,
        .api-test textarea {
            width: 100%;
            padding: 1rem;
            border: 2px solid var(--gray-200);
            border-radius: var(--radius);
            margin: 0.75rem 0;
            font-family: inherit;
            font-size: 0.875rem;
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            background: var(--white);
            position: relative;
        }

        .api-test input:focus,
        .api-test textarea:focus {
            outline: none;
            border-color: var(--primary);
            border-left-style: dashed;
            box-shadow: 0 0 0 4px rgb(99 102 241 / 0.1);
            transform: translateY(-2px);
            background: var(--white);
        }

        .api-test label {
            display: block;
            font-size: 0.875rem;
            font-weight: 600;
            color: var(--gray-800);
            margin: 1.5rem 0 0.5rem 0;
            transition: color 0.3s ease;
        }

        .api-test input:focus + label,
        .api-test textarea:focus + label {
            color: var(--primary);
        }

        .api-test button {
            background: linear-gradient(135deg, var(--primary) 0%, var(--secondary) 100%);
            color: var(--white);
            padding: 0.6rem 1.9rem;
            border: none;
            border-radius: var(--radius);
            font-weight: 600;
            cursor: pointer;
            position: relative;
            overflow: hidden;
            font-size: 0.875rem;
        }

        .api-test button::before {
            content: '';
            position: absolute;
            top: 0;
            left: -100%;
            width: 100%;
            height: 100%;
            background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
        }

        .api-test button:hover::before {
            left: 100%;
        }

        .api-test button:hover {
            background: linear-gradient(135deg, var(--primary-dark) 0%, var(--secondary) 100%);
            transform: translateY(-3px);
        }

        .api-test button:active {
            transform: translateY(-1px);
        }

        .lock-icon {
            color: var(--warning);
            font-size: 1rem;
            background: rgba(245, 158, 11, 0.1);
            padding: 0.50rem;
            border-radius: 50%;
            transition: all 0.3s ease;
        }

        .method-group:hover .lock-icon {
            background: var(--warning);
            color: var(--white);
            transform: scale(1.1);
        }

        ul {
            margin: 0;
            padding-left: 1.25rem;
        }

        li {
            margin: 0.5rem 0;
            color: var(--gray-700);
        }

        @media (max-width: 768px) {
            .container {
                padding: 1rem;
            }
            
            h1 {
                font-size: 2rem;
            }
            
            .auth-input-group {
                flex-direction: column;
            }
            
            .method-group > summary {
                padding-left: 2rem;
            }
            
            .segment-group > summary {
                padding-left: 1.5rem;
            }
            
            .path-group > summary {
                padding-left: 2rem;
            }
        }
        
        /* JSON Editor Styles */
        .json-editor-container {
            position: relative;
            border: 2px solid var(--gray-200);
            border-radius: var(--radius);
            margin: 0.75rem 0;
            overflow: hidden;
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
        }
        
        .json-editor-container:focus-within {
            border-color: var(--primary);
            border-left-style: dashed;
            box-shadow: 0 0 0 4px rgb(99 102 241 / 0.1);
            transform: translateY(-2px);
        }
        
        .json-editor {
            min-height: 120px;
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.875rem;
            line-height: 1.4;
        }
        
        .json-editor-toolbar {
            background: var(--gray-50);
            border-bottom: 1px solid var(--gray-200);
            padding: 0.5rem;
            display: flex;
            gap: 0.5rem;
            align-items: center;
        }
        
        .json-editor-btn {
            background: var(--white);
            border: 1px solid var(--gray-300);
            border-radius: 4px;
            padding: 0.25rem 0.5rem;
            font-size: 0.75rem;
            cursor: pointer;
            transition: all 0.2s ease;
        }
        
        .json-editor-btn:hover {
            background: var(--gray-100);
            border-color: var(--primary);
        }
        
        .json-validation-message {
            padding: 0.5rem;
            font-size: 0.75rem;
            border-top: 1px solid var(--gray-200);
            background: var(--gray-50);
        }
        
        .json-validation-message.error {
            background: #fef2f2;
            color: var(--danger);
            border-color: #fecaca;
        }
        
        .json-validation-message.success {
            background: #f0fdf4;
            color: var(--success);
            border-color: #bbf7d0;
        }
    </style>
    
    <!-- CodeMirror for JSON editing -->
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/codemirror.min.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/theme/default.min.css">
    <script src="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/codemirror.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/mode/javascript/javascript.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/addon/lint/lint.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/addon/lint/json-lint.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/addon/edit/closebrackets.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/addon/edit/matchbrackets.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/addon/fold/foldcode.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/addon/fold/foldgutter.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.2/addon/fold/brace-fold.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/jsonlint/1.6.0/jsonlint.min.js"></script>
    
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>` + an.config.Title + `</h1>
            <p class="subtitle">` + an.config.Description + `</p>
            <span class="version-badge">` + an.config.Version + `</span>
        </div>
        
        <div class="auth-section">
            <h2><i class="fas fa-key"></i> Authorize</h2>
            <div class="auth-input-group">
                <input type="text" id="auth-token" placeholder="Enter JWT Bearer Token (e.g., Bearer eyJ...)" value="` + an.config.AuthToken + `">
                <button onclick="setAuthToken()">Set Token</button>
            </div>
        </div>
        
        <div class="section-header">
            <h2 class="section-title">API Endpoints</h2>
            <a href="/api-docs/metrics" target="_blank" class="monitor-button">
                <i class="fas fa-chart-line"></i>
                Monitor
            </a>
        </div>`)

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
                <summary>
                    <span class="method ` + endpoint.Method + `">` + endpoint.Method + `</span>
                    <span class="endpoint-path">` + endpoint.Path + `</span>
                    <span class="endpoint-description">` + endpoint.Description + `</span>` + lockIcon + `
                </summary>
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
                        <form id="test-form-` + endpoint.Method + "-" + strings.ReplaceAll(strings.ReplaceAll(endpoint.Path, "/", "-"), ":", "_") + `" onsubmit="testApi(event, '` + endpoint.Method + `', '` + endpoint.Path + `', this)" enctype="multipart/form-data">
                            <input type="hidden" name="method" value="` + endpoint.Method + `">`)

						// hasFormData := false
						for _, param := range endpoint.Parameters {
							inputType := "text"
							if param.Type == "number" {
								inputType = "number"
							} else if param.Type == "file" {
								inputType = "file"
								// hasFormData = true
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
							// Generate JSON template from request schema
							jsonTemplate := ""
							if endpoint.RequestSchema != nil {
								if template, err := generateJSONTemplate(endpoint.RequestSchema); err == nil {
									// Properly escape the JSON for HTML attribute
									jsonTemplate = strings.ReplaceAll(template, `"`, `&quot;`)
									jsonTemplate = strings.ReplaceAll(jsonTemplate, `'`, `&#39;`)
									jsonTemplate = strings.ReplaceAll(jsonTemplate, `\`, `\\`)
								}
							}

							if endpoint.Parameters == nil || len(endpoint.Parameters) == 0 {
								html.WriteString(`
                                <label>Request Body (JSON):</label>
                                <div class="json-editor-container" data-template="` + jsonTemplate + `">
                                    <div class="json-editor-toolbar">
                                        <button type="button" class="json-editor-btn" onclick="formatJSON(this)">
                                            <i class="fas fa-magic"></i> Format
                                        </button>
                                        <button type="button" class="json-editor-btn" onclick="validateJSON(this)">
                                            <i class="fas fa-check-circle"></i> Validate
                                        </button>
                                        <button type="button" class="json-editor-btn" onclick="clearJSON(this)">
                                            <i class="fas fa-trash"></i> Clear
                                        </button>
                                        <button type="button" class="json-editor-btn" onclick="loadSchemaTemplate(this)">
                                            <i class="fas fa-file-code"></i> Load Template
                                        </button>
                                    </div>
                                    <textarea name="requestBody" class="json-editor" placeholder="Enter JSON request body..."></textarea>
                                    <div class="json-validation-message" style="display: none;"></div>
                                </div>`)
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

            window.onload = function() {
                const authInput = document.getElementById('auth-token');
                if (authInput) {
                    authInput.value = authToken;
                }
            };

            function setAuthToken() {
                const authInput = document.getElementById('auth-token');
                authToken = authInput.value.trim();
                localStorage.setItem('authToken', authToken);
                alert('Authorization token set: ' + (authToken ? authToken : 'None'));
            }

            function testApi(event, method, path, form) {
                event.preventDefault();
                const resultElement = document.getElementById('test-result-' + method + path.replace(/\//g, '-'));
                resultElement.textContent = 'Sending request...';

                const params = {};
                const queryParams = new URLSearchParams();
                const formData = new FormData();
                let isFormDataRequest = false;

                // Process form inputs
                const inputs = form.querySelectorAll('input, textarea');
                let modifiedPath = path; // Start with the original path
                inputs.forEach(input => {
                    const key = input.name;
                    const value = input.value;
                    const paramIn = input.getAttribute('data-in');

                    if (key && paramIn) {
                        if (paramIn === 'formData') {
                            isFormDataRequest = true;
                            if (input.type === 'file' && input.files.length > 0) {
                                formData.append(key, input.files[0]);
                            } else if (value) {
                                formData.append(key, value);
                            }
                        } else if (paramIn === 'path' && value) {
                            // Replace :key with the value in the path
                            modifiedPath = modifiedPath.replace(':' + key, encodeURIComponent(value));
                        } else if (paramIn === 'query' && value) {
                            queryParams.append(key, value);
                        } else if (paramIn === 'header' && value) {
                            params[key] = value;
                        }
                    }
                });

                const baseUrl = 'http://' + '` + an.config.Host + `';
                const url = baseUrl + modifiedPath + (queryParams.toString() ? '?' + queryParams.toString() : '');

                const options = {
                    method: method,
                    headers: {},
                };

                if (authToken) {
                    const token = authToken.startsWith('Bearer ') ? authToken : 'Bearer ' + authToken;
                    options.headers['Authorization'] = token;
                }

                Object.keys(params).forEach(key => {
                    if (params[key]) {
                        options.headers[key] = params[key];
                    }
                });

                if (isFormDataRequest) {
                    options.body = formData;
                } else if (method === 'POST' || method === 'PUT' || method === 'PATCH') {
                    const requestBodyInput = form.querySelector('textarea[name="requestBody"]');
                    if (requestBodyInput) {
                        // Sync CodeMirror content if it exists
                        if (requestBodyInput.hasAttribute('data-editor-id')) {
                            const editorId = requestBodyInput.getAttribute('data-editor-id');
                            const codeMirrorEditor = codeMirrorEditors[editorId];
                            if (codeMirrorEditor) {
                                codeMirrorEditor.save();
                            }
                        }
                        
                        const bodyContent = requestBodyInput.value.trim();
                        if (bodyContent) {
                            try {
                                const jsonBody = JSON.parse(bodyContent);
                                options.headers['Content-Type'] = 'application/json';
                                options.body = JSON.stringify(jsonBody);
                            } catch (e) {
                                resultElement.textContent = 'Invalid JSON in request body: ' + e.message;
                                return;
                            }
                        }
                        // If bodyContent is empty, don't set any body - this allows requests without bodies
                    }
                }

                fetch(url, options)
                    .then(response => {
                        const contentType = response.headers.get('content-type') || '';
                        const disposition = response.headers.get('content-disposition') || '';
                        let filename = 'download';
                        if (disposition) {
                            const matches = disposition.match(/filename="([^"]+)"/);
                            if (matches && matches[1]) {
                                filename = matches[1];
                            }
                        }

                        // Capture all response headers
                        const headers = {};
                        for (let [key, value] of response.headers.entries()) {
                            headers[key] = value;
                        }

                        if (!response.ok) {
                            return response.text().then(text => ({
                                status: response.status,
                                statusText: response.statusText,
                                body: text,
                                contentType: contentType,
                                headers: headers,
                                isError: true
                            }));
                        } else if (contentType.includes('application/json')) {
                            return response.json().then(data => ({
                                status: response.status,
                                statusText: response.statusText,
                                body: JSON.stringify(data, null, 2),
                                contentType: contentType,
                                headers: headers
                            }));
                        } else if (contentType.startsWith('image/')) {
                            return response.blob().then(blob => ({
                                status: response.status,
                                statusText: response.statusText,
                                body: blob,
                                contentType: contentType,
                                headers: headers,
                                isImage: true,
                                filename: filename
                            }));
                        } else {
                            return response.blob().then(blob => ({
                                status: response.status,
                                statusText: response.statusText,
                                body: blob,
                                contentType: contentType,
                                headers: headers,
                                isBlob: true,
                                filename: filename
                            }));
                        }
                    })
                    .then(result => {
                        resultElement.innerHTML = "Url: " + url + "<br>Status: " + result.status + " " + result.statusText + "<br>";
                        
                        // Display response headers
                        if (result.headers && Object.keys(result.headers).length > 0) {
                            resultElement.innerHTML += "<br><strong>Response Headers:</strong><br>";

                            for (const [key, value] of Object.entries(result.headers)) {
                                resultElement.innerHTML += escapeHtml(key) + ": " + escapeHtml(value) + "\n";
                            }
                        }
                        
                        resultElement.innerHTML += "<br>";
                        
                        if (result.isError) {
                            resultElement.innerHTML += '<strong>Error Response:</strong><br><pre>' + escapeHtml(result.body) + '</pre>';
                        } else if (result.isImage) {
                            const imgUrl = URL.createObjectURL(result.body);
                            resultElement.innerHTML += '<strong>Response (Image):</strong><br><img src="' + imgUrl + '" style="max-width: 100%;" onload="setTimeout(() => URL.revokeObjectURL(this.src), 10000)">';
                            // Also provide download link for images
                            resultElement.innerHTML += '<br><a href="' + imgUrl + '" download="' + escapeHtml(result.filename) + '">Download Image</a>';
                            setTimeout(() => URL.revokeObjectURL(imgUrl), 30000); // Longer timeout for images
                        } else if (result.isBlob) {
                            const blobUrl = URL.createObjectURL(result.body);
                            const downloadId = 'download-link-' + Date.now();
                            resultElement.innerHTML += '<strong>Response (File):</strong><br>';
                            resultElement.innerHTML += '<div class="download-info">';
                            resultElement.innerHTML += '<i class="fas fa-download"></i> ';
                            resultElement.innerHTML += '<a href="' + blobUrl + '" download="' + escapeHtml(result.filename) + '" id="' + downloadId + '">' + escapeHtml(result.filename) + '</a>';
                            resultElement.innerHTML += '<span style="margin-left: 10px; color: var(--info);">(' + (result.body.size ? (result.body.size / 1024).toFixed(1) + ' KB' : 'Unknown size') + ')</span>';
                            resultElement.innerHTML += '</div>';
                            
                            // Automatically trigger download
                            const downloadLink = document.getElementById(downloadId);
                            if (downloadLink) {
                                // Add click event to show download status
                                downloadLink.addEventListener('click', function() {
                                    const statusSpan = document.createElement('span');
                                    statusSpan.style.marginLeft = '10px';
                                    statusSpan.style.color = 'var(--success)';
                                    statusSpan.innerHTML = '<i class="fas fa-check"></i> Download started';
                                    this.parentNode.appendChild(statusSpan);
                                });
                                
                                downloadLink.click();
                                // Revoke object URL after download is triggered with a longer delay
                                setTimeout(() => URL.revokeObjectURL(blobUrl), 5000);
                            }
                        } else {
                            resultElement.innerHTML += '<strong>Response Body:</strong><br><pre>' + escapeHtml(result.body) + '</pre>';
                        }
                    })
                    .catch(error => {
                        console.error('Fetch error:', error);
                        resultElement.innerHTML = '<strong>Error:</strong><br><pre style="color: var(--danger);">' + escapeHtml(error.message) + '</pre>';
                        
                        // Provide more detailed error information
                        if (error.name === 'TypeError' && error.message.includes('fetch')) {
                            resultElement.innerHTML += '<br><small>This might be a network connectivity issue or CORS error.</small>';
                        } else if (error.name === 'AbortError') {
                            resultElement.innerHTML += '<br><small>Request was aborted.</small>';
                        }
                    });
            }

            function escapeHtml(unsafe) {
                if (typeof unsafe !== 'string') return unsafe;
                return unsafe
                    .replace(/&/g, "&amp;")
                    .replace(/</g, "&lt;")
                    .replace(/>/g, "&gt;")
                    .replace(/"/g, "&quot;")
                    .replace(/'/g, "&#039;");
            }

            // JSON Editor functionality
            let codeMirrorEditors = {};

            // Initialize CodeMirror editors for all JSON textareas
            document.addEventListener('DOMContentLoaded', function() {
                document.querySelectorAll('textarea.json-editor').forEach(function(textarea) {
                    const editorId = 'editor_' + Math.random().toString(36).substr(2, 9);
                    
                    const editor = CodeMirror.fromTextArea(textarea, {
                        mode: { name: "javascript", json: true },
                        theme: "default",
                        lineNumbers: true,
                        lineWrapping: true,
                        autoCloseBrackets: true,
                        matchBrackets: true,
                        indentUnit: 2,
                        tabSize: 2,
                        foldGutter: true,
                        gutters: ["CodeMirror-linenumbers", "CodeMirror-foldgutter"],
                        lint: true,
                        placeholder: "Enter JSON request body..."
                    });

                    // Store editor reference
                    codeMirrorEditors[editorId] = editor;
                    textarea.setAttribute('data-editor-id', editorId);

                    // Auto-validate on change
                    editor.on('change', function() {
                        setTimeout(() => validateJSONEditor(editor), 300);
                    });

                    // Set default content if template exists
                    const container = textarea.closest('.json-editor-container');
                    if (container) {
                        const form = container.closest('form');
                        if (form) {
                            const method = form.querySelector('button[type="submit"]').closest('form').id;
                            loadDefaultTemplate(editor, method);
                        }
                    }
                });
            });

            function getEditorFromButton(button) {
                const container = button.closest('.json-editor-container');
                const textarea = container.querySelector('textarea.json-editor');
                const editorId = textarea.getAttribute('data-editor-id');
                return codeMirrorEditors[editorId];
            }

            function formatJSON(button) {
                const editor = getEditorFromButton(button);
                const content = editor.getValue().trim();
                
                if (!content) {
                    showValidationMessage(button, 'No JSON content to format', 'error');
                    return;
                }

                try {
                    const parsed = JSON.parse(content);
                    const formatted = JSON.stringify(parsed, null, 2);
                    editor.setValue(formatted);
                    showValidationMessage(button, 'JSON formatted successfully', 'success');
                } catch (e) {
                    showValidationMessage(button, 'Invalid JSON: ' + e.message, 'error');
                }
            }

            function validateJSON(button) {
                const editor = getEditorFromButton(button);
                validateJSONEditor(editor);
            }

            function validateJSONEditor(editor) {
                const content = editor.getValue().trim();
                const container = editor.getTextArea().closest('.json-editor-container');
                const messageDiv = container.querySelector('.json-validation-message');

                if (!content) {
                    messageDiv.style.display = 'none';
                    return;
                }

                try {
                    JSON.parse(content);
                    showValidationMessage(container, 'Valid JSON ✓', 'success');
                } catch (e) {
                    showValidationMessage(container, 'Invalid JSON: ' + e.message, 'error');
                }
            }

            function clearJSON(button) {
                const editor = getEditorFromButton(button);
                editor.setValue('');
                const container = button.closest('.json-editor-container');
                const messageDiv = container.querySelector('.json-validation-message');
                messageDiv.style.display = 'none';
            }

            function loadSchemaTemplate(button) {
                const container = button.closest('.json-editor-container');
                let template = container.getAttribute('data-template');
                
                if (!template || template === '{}') {
                    showValidationMessage(container, 'No template available for this endpoint', 'error');
                    return;
                }

                // Decode HTML entities
                template = template.replace(/&quot;/g, '"').replace(/&#39;/g, "'").replace(/\\\\/g, '\\');

                const editor = getEditorFromButton(button);
                try {
                    // Parse and reformat the template to ensure proper formatting
                    const parsed = JSON.parse(template);
                    const formatted = JSON.stringify(parsed, null, 2);
                    editor.setValue(formatted);
                    showValidationMessage(container, 'Schema template loaded successfully', 'success');
                } catch (e) {
                    showValidationMessage(container, 'Invalid template: ' + e.message, 'error');
                }
            }

            function loadDefaultTemplate(editor, method) {
                // Auto-load templates based on the schema
                const container = editor.getTextArea().closest('.json-editor-container');
                let template = container.getAttribute('data-template');
                
                if (template && template !== '{}') {
                    try {
                        // Decode HTML entities
                        template = template.replace(/&quot;/g, '"').replace(/&#39;/g, "'").replace(/\\\\/g, '\\');
                        const parsed = JSON.parse(template);
                        const formatted = JSON.stringify(parsed, null, 2);
                        editor.setValue(formatted);
                    } catch (e) {
                        console.warn('Failed to load default template:', e);
                    }
                }
            }

            function showValidationMessage(elementOrContainer, message, type) {
                let container;
                if (elementOrContainer.classList && elementOrContainer.classList.contains('json-editor-container')) {
                    container = elementOrContainer;
                } else {
                    container = elementOrContainer.closest('.json-editor-container');
                }

                const messageDiv = container.querySelector('.json-validation-message');
                messageDiv.textContent = message;
                messageDiv.className = 'json-validation-message ' + type;
                messageDiv.style.display = 'block';

                if (type === 'success') {
                    setTimeout(() => {
                        messageDiv.style.display = 'none';
                    }, 3000);
                }
            }

            // Update the existing form submission to work with CodeMirror
            document.addEventListener('submit', function(e) {
                if (e.target.tagName === 'FORM') {
                    const editors = e.target.querySelectorAll('textarea.json-editor');
                    editors.forEach(function(editor) {
                        if (editor.hasAttribute('data-editor-id')) {
                            const editorId = editor.getAttribute('data-editor-id');
                            const codeMirrorEditor = codeMirrorEditors[editorId];
                            if (codeMirrorEditor) {
                                // Sync CodeMirror content back to textarea
                                codeMirrorEditor.save();
                            }
                        }
                    });
                }
            });
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
