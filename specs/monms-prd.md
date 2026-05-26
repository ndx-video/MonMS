# **Product Requirement Document (PRD)**

## **Project Name: MonMS — The Agent-Malleable Monolithic CMS**

## **1\. Executive Summary & Vision**

Traditional Content Management Systems (CMSs) enforce strict boundaries:

* **Developers** write code, schemas, and templates.  
* **Clients/Editors** manage content in predefined input boxes.  
* **Architectures** require complex build steps, CI/CD pipelines, and high-overhead cloud infrastructure.

**MonMS** (Monolithic Management System) is a radical departure from this model. It is an **agent-malleable, single-binary monolith** built on Go and PocketBase that treats database schemas, UI structures, and content as a fluid, singular continuum.

Instead of writing code or manually configuring CMS fields, an engineering-oriented AI agent manages the application's structure by directly mutating a Git-managed folder tree (containing HTML templates, CSS, and schema definitions).

For clients, MonMS provides **Inline Contextual Editing**—the ability to safely edit content in-place directly on the live website using HTMX, secured by PocketBase's lightweight authentication engine. It retains the native PocketBase developer dashboard as a fallback.

### **Key Architectural Pillars**

1. **Single-Binary Monolith:** An entire production-grade CMS, database, file server, and web server compiled into a single executable using \< 30MB RAM.  
2. **Zero-Compilation Malleability:** The application UI/UX, routing, and database schemas are modified on the fly without restarting or rebuilding the Go binary.  
3. **Git-Managed State:** Every AI structural adjustment (new page, updated CSS, modified collection layout) is tracked, versioned, and rolled back via Git.  
4. **Inline Contextual Editing:** Clients log in and click to type directly on the live site, instantly updating SQLite records behind the scenes.

## **2\. Architectural Design & Folder Topology**

The system uses PocketBase as an embedded Go framework. The Go binary remains completely frozen and generic. All application logic, assets, templates, and schemas live in a sibling folder (/workspace) which is a tracked Git repository.

### **Directory Structure**

├── monms-engine          \# Pre-compiled Go monolithic executable  
├── pb\_data/              \# SQLite database (contains client & agent raw data)  
└── workspace/            \# The Git-tracked AI & Human workspace  
    ├── schema/           \# Declared collection schemas (JSON backups)  
    ├── templates/        \# Go HTML templates parsed on demand  
    │   ├── layouts/      \# Global page structures (base layout, head, auth check)  
    │   ├── fragments/    \# Component parts or partials used by HTMX  
    │   └── \*.gohtml      \# Live route templates (index, about, pricing, team)  
    └── assets/           \# Client-facing CSS, JS, fonts, and uploaded media

## **3\. High-Performance Runtime Malleability (Go Backend)**

To allow the AI to mutate the system without service interruption, MonMS implements a custom, performance-optimized, on-demand parsing engine in Go.

package main

import (  
	"html/template"  
	"io/fs"  
	"log"  
	"net/http"  
	"os"  
	"path/filepath"  
	"sync"

	"github.com/fsnotify/fsnotify"  
	"github.com/pocketbase/pocketbase"  
	"github.com/pocketbase/pocketbase/core"  
)

type TemplateCache struct {  
	mu     sync.RWMutex  
	cache  map\[string\]\*template.Template  
	active bool  
}

// Global cache instance to avoid disk reads under heavy traffic,  
// but instantly invalidated on Git pull / Agent updates.  
var tplCache \= \&TemplateCache{  
	cache:  make(map\[string\]\*template.Template),  
	active: os.Getenv("ENV") \== "production",   
}

func main() {  
	app := pocketbase.New()

	// Initialize filesystem watcher to automatically clear cache when the agent commits files  
	go watchWorkspace(tplCache)

	app.OnServe().BindFunc(func(se \*core.ServeEvent) error {  
		// Static asset serving directly from the Git-managed workspace  
		se.Router.GET("/assets/{path...}", func(e \*core.RequestEvent) error {  
			filePath := filepath.Join("./workspace/assets", e.Request.PathValue("path"))  
			return e.File(filePath)  
		})

		// Catch-all SSR Route handler  
		se.Router.GET("/{slug...}", func(e \*core.RequestEvent) error {  
			slug := e.Request.PathValue("slug")  
			if slug \== "" {  
				slug \= "index"  
			}

			tmpl, err := getTemplate(slug)  
			if err \!= nil {  
				return e.NotFoundError("Template variant not found or syntax error", err)  
			}

			// Context Injection for rendering  
			data := map\[string\]any{  
				"IsLoggedIn": e.Auth \!= nil,  
				"User":       e.Auth,  
				"Slug":       slug,  
				"App":        app, // Expose pocketbase helper context safely if needed  
			}

			e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")  
			return tmpl.Execute(e.Response, data)  
		})

		return se.Next()  
	})

	if err := app.Start(); err \!= nil {  
		log.Fatal(err)  
	}  
}

func getTemplate(slug string) (\*template.Template, error) {  
	templatePath := filepath.Join("./workspace/templates", slug+".gohtml")  
	layoutPath := filepath.Join("./workspace/templates/layouts/base.gohtml")

	tplCache.mu.RLock()  
	if tplCache.active {  
		if cached, exists := tplCache.cache\[slug\]; exists {  
			tplCache.mu.RUnlock()  
			return cached, nil  
		}  
	}  
	tplCache.mu.RUnlock()

	// Read directly from disk on cache miss/development mode  
	tmpl, err := template.ParseFiles(layoutPath, templatePath)  
	if err \!= nil {  
		return nil, err  
	}

	tplCache.mu.Lock()  
	if tplCache.active {  
		tplCache.cache\[slug\] \= tmpl  
	}  
	tplCache.mu.Unlock()

	return tmpl, nil  
}

func watchWorkspace(c \*TemplateCache) {  
	watcher, err := fsnotify.NewWatcher()  
	if err \!= nil {  
		return  
	}  
	defer watcher.Close()

	\_ \= watcher.Add("./workspace/templates")

	for {  
		select {  
		case event, ok := \<-watcher.Events:  
			if \!ok {  
				return  
			}  
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {  
				c.mu.Lock()  
				c.cache \= make(map\[string\]\*template.Template) // Invalidate entire cache  
				c.mu.Unlock()  
			}  
		case \<-watcher.Errors:  
			return  
		}  
	}  
}

## **4\. The Agent Mutation Engine (How the AI Modifies the CMS)**

When instructed by a user via an integrated chatbox or API hook, the AI Agent acts as an active workspace editor. It reasons over the current system state and applies atomic mutations.

                  ┌────────────────────────────────────────┐  
                  │          AI Agent Interaction          │  
                  └───────────────────┬────────────────────┘  
                                      │  
                         \[Analyzes Current Schema/UI\]  
                                      │  
            ┌─────────────────────────┴─────────────────────────┐  
            ▼                                                   ▼  
┌───────────────────────┐                           ┌───────────────────────┐  
│     Modifies UI       │                           │    Mutates Schema     │  
│  (Templates & CSS)    │                           │ (PocketBase API/JSON) │  
└───────────┬───────────┘                           └───────────┬───────────┘  
            │                                                   │  
            │                  \[Git Push Commit\]                │  
            └─────────────────────────┬─────────────────────────┘  
                                      ▼  
                        ┌───────────────────────────┐  
                        │      On-Demand Parse      │  
                        │   (Live Update in Go)     │  
                        └───────────────────────────┘

### **1\. Schema Mutation**

Instead of restarting the server, the Agent alters PocketBase schemas programmatically by posting to the native admin collections API (/api/collections), or writing declarative migration files.

*Example Agent Operation:* Create a press\_releases table.

* **Agent Execution:** Issues a POST /api/collections request.  
* **Result:** The SQLite physical tables are generated on-the-fly. PocketBase's dynamic routing exposes the endpoints instantly.

### **2\. UI/UX Mutation**

The Agent accesses the template directory, modifies markup, and creates CSS rules.

*Example Agent Operation:*\* Add a layout block displaying the newly created press\_releases.

* **Agent Execution:** Rewrites ./workspace/templates/press.gohtml injecting an HTMX layout fetching from /api/collections/press\_releases/records.  
* **Result:** On the next browser reload, Go detects the template change, parses the raw .gohtml file, and outputs the brand-new component.

### **3\. Safety Guardrails for Agent Code**

To prevent syntax runtime panics or breaking the UI layout, the Agent platform runs the following checks before committing its files:

* **Validation Stage:** Compiles the updated template in an isolated Go dry-run environment using standard parsing.  
* **HTML Linting:** Runs an HTML structure validator to ensure no broken tags damage the layout.  
* **Rollback Path:** The workspace directory is a tracked Git repo. If the verification fails, the engine runs git checkout \-- . to revert back to the last stable state instantly.

## **5\. Flagship Feature: Inline Contextual Editing**

While the AI handles high-level structural and design mutations, human clients require an intuitive, zero-overhead workflow for day-to-day text updates. MonMS achieves this using HTML5's native contenteditable coupled with **HTMX**.

When an authenticated client views the site, MonMS injects editability directly into the SSR layout.

### **Global Base Layout Layout (/workspace/templates/layouts/base.gohtml)**

This template automatically enables inline saving functionality if the client is logged in.

{{define "base"}}  
\<\!DOCTYPE html\>  
\<html lang="en"\>  
\<head\>  
    \<meta charset="UTF-8"\>  
    \<meta name="viewport" content="width=device-width, initial-scale=1.0"\>  
    \<title\>{{block "title" .}}MonMS{{end}}\</title\>  
    \<\!-- AlpineJS & HTMX loaded globally \--\>  
    \<script src="https://unpkg.com/htmx.org@1.9.10"\>\</script\>  
    \<script src="https://defer.js"\>\</script\>   
    \<link rel="stylesheet" href="/assets/main.css"\>  
\</head\>  
\<body class="bg-slate-50 text-slate-900 min-h-screen"\>

    \<\!-- Embedded Admin Login/Status Indicator Overlay \--\>  
    {{if .IsLoggedIn}}  
    \<div class="fixed top-4 right-4 z-50 bg-indigo-600 text-white py-2 px-4 rounded-full shadow-lg text-sm font-medium flex items-center gap-3"\>  
        \<span class="flex h-2 w-2 relative"\>  
          \<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"\>\</span\>  
          \<span class="relative inline-flex rounded-full h-2 w-2 bg-green-500"\>\</span\>  
        \</span\>  
        Live Editor Active  
        \<a href="/\_/" class="underline text-indigo-200 hover:text-white transition"\>Full Admin Dashboard\</a\>  
    \</div\>  
    {{end}}

    \<main class="max-w-6xl mx-auto px-6 py-12"\>  
        {{template "body" .}}  
    \</main\>

    \<\!-- Configure HTMX request headers to handle PocketBase JWT authorization tokens seamlessly \--\>  
    \<script\>  
        document.addEventListener("DOMContentLoaded", () \=\> {  
            const pbAuthCookie \= document.cookie.split('; ').find(row \=\> row.startsWith('pb\_auth='));  
            if (pbAuthCookie) {  
                const token \= pbAuthCookie.split('=')\[1\];  
                document.body.addEventListener('htmx:configRequest', (event) \=\> {  
                    event.detail.headers\['Authorization'\] \= 'Bearer ' \+ token;  
                });  
            }  
        });  
    \</script\>  
\</body\>  
\</html\>  
{{end}}

### **Page Template Implementation (/workspace/templates/index.gohtml)**

This template uses HTMX attributes to trigger REST API requests back to the dynamic PocketBase collection on edit blur.

{{define "title"}}Welcome to our MonMS Portal{{end}}

{{define "body"}}  
{{$record := .App.FindRecordById "hero\_content" "homepage"}}

\<div class="space-y-6 max-w-2xl"\>  
    \<\!-- Inline Title Editing \--\>  
    \<h1 class="text-4xl font-extrabold tracking-tight"  
        {{if .IsLoggedIn}}  
        contenteditable="true"  
        hx-put="/api/collections/hero\_content/records/homepage"  
        hx-trigger="blur"  
        hx-ext="json-enc"  
        hx-vals='js:{"title": event.target.innerText}'  
        {{end}}\>  
        {{$record.Get "title"}}  
    \</h1\>

    \<\!-- Inline Rich Paragraph Editing \--\>  
    \<p class="text-lg text-slate-600 leading-relaxed"  
        {{if .IsLoggedIn}}  
        contenteditable="true"  
        hx-put="/api/collections/hero\_content/records/homepage"  
        hx-trigger="blur"  
        hx-vals='js:{"body": event.target.innerText}'  
        {{end}}\>  
        {{$record.Get "body"}}  
    \</p\>  
\</div\>  
{{end}}

### **The UX Flow for Humans**

1. **Login:** The client logs in at /admin/login or via the standard PocketBase dashboard interface. A secure HttpOnly cookie containing the Session JWT is written.  
2. **Navigate:** When loading the page, the Go router notes the active Session, passing IsLoggedIn: true to the templates.  
3. **Direct Interaction:** Elements marked with contenteditable="true" become interactive. The user selects a heading, changes a typo, and clicks away.  
4. **Instant Save:** HTMX intercepts the blur event, reads the text buffer, and issues a standard RESTful PUT directly to PocketBase's core system endpoint. A successful database update occurs securely without reloading the page.

## **6\. Feature Matrix & Comparative Positioning**

| Feature | Classic Headless (Contentful/Strapi) | Static Site Gen (Astro/Eleventy) | MonMS (Malleable Monolith) |
| :---- | :---- | :---- | :---- |
| **Server Complexity** | High (Requires separated host \+ CDN \+ API DB) | Moderate (Build servers, webhook runs) | **Extremely Low (Single executable, embedded SQLite)** |
| **Schema Alterations** | Click-heavy admin panels | Manual developer code modifications | **Instantaneous via dynamic API (AI Agent Managed)** |
| **Editing Interface** | Detached administrative grid forms | Git-based markdown file updates | **Inline Contextual Visual WYSIWYG (Via HTMX)** |
| **Deploy Time** | \~1-5 Minutes (Full rebuild cycles) | \~2-10 Minutes (CI pipelines) | **Instantaneous (Disk writes directly loaded in-memory)** |
| **Memory Footprint** | \~512MB \- 1GB | N/A (Build stage, not SSR runtime) | **\< 30 Megabytes RAM under production idle** |

## **7\. Key Non-Functional Requirements (NFRs)**

### **1\. Performance and Resource Profile**

* **Idle Overhead:** Memory footprint of the running Go binary must remain under **30MB RAM**.  
* **Time To First Byte (TTFB):** Standard server-side rendered routes under SQLite reads must process in \< 15ms.  
* **Asset Footprint:** To keep files incredibly lean, the visual framework defaults to native CSS or CDN Tailwind CSS imports, requiring no local Node.js compilation pipeline.

### **2\. Version Control & Audit Trail**

* All mutations done by the AI agent on layouts, routing, or schemas *must* be captured in a git commit.  
* The system schedules regular, automatic git commits when human inline actions modify underlying files or static content properties to prevent local state drift.

### **3\. Permissions & Security**

* **Agent Execution Layer:** The AI agent operates using dedicated SSH keys and REST API tokens restricting permissions strictly to the active workspace subdirectory.  
* **Editor Layer:** Strict PocketBase API rules apply. Unauthenticated users cannot overwrite fields, as validation is enforced at the database layer, blockading unauthorized PUT calls.