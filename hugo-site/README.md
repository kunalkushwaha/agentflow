# AgentFlow Documentation Site

This directory contains the Hugo-based documentation site for AgentFlow. The site is automatically deployed to GitHub Pages at https://kunalkushwaha.github.io/agentflow/.

## 🏗️ Site Structure

```
hugo-site/
├── config.toml              # Hugo configuration
├── content/
│   ├── _index.md            # Landing page
│   ├── docs/                # Documentation section
│   │   ├── _index.md
│   │   ├── agent-basics.md
│   │   ├── configuration.md
│   │   ├── tool-integration.md
│   │   └── architecture.md
│   ├── examples/            # Examples section
│   │   ├── _index.md
│   │   └── single-agent.md
│   └── contributors/        # Contributors section
│       ├── _index.md
│       └── contributor-guide.md
├── themes/agentflow/        # Custom theme
│   ├── layouts/
│   │   ├── _default/
│   │   │   ├── baseof.html
│   │   │   ├── home.html
│   │   │   ├── list.html
│   │   │   └── single.html
│   │   └── partials/
│   │       ├── nav-docs.html
│   │       ├── nav-examples.html
│   │       └── nav-contributors.html
│   └── static/
│       └── css/
│           └── style.css
└── README.md               # This file
```

## 🚀 Local Development

### Prerequisites

- [Hugo Extended](https://gohugo.io/installation/) v0.119.0 or later
- [Node.js](https://nodejs.org/) (for MCP server examples)

### Running Locally

```bash
# Navigate to the hugo-site directory
cd hugo-site

# Start the development server
hugo server -D

# Open your browser to http://localhost:1313
```

The site will automatically reload when you make changes to content or theme files.

## 📝 Adding Content

### Adding a New Documentation Page

1. Create a new Markdown file in `content/docs/`:
   ```bash
   hugo new content/docs/your-new-page.md
   ```

2. Add front matter:
   ```yaml
   ---
   title: "Your Page Title"
   description: "Brief description of the page content"
   weight: 50
   ---
   ```

3. Update the navigation in `themes/agentflow/layouts/partials/nav-docs.html`:
   ```html
   <a class="nav-link" href="/docs/your-new-page/">Your Page Title</a>
   ```

### Adding a New Example

1. Create a new file in `content/examples/`:
   ```bash
   hugo new content/examples/your-example.md
   ```

2. Follow the same pattern as existing examples with complete code samples and explanations.

### Adding Contributor Documentation

1. Create a new file in `content/contributors/`:
   ```bash
   hugo new content/contributors/your-guide.md
   ```

2. Focus on development-related content for contributors.

## 🎨 Theme Customization

The site uses a custom Bootstrap 5-based theme with the following features:

- **Responsive Design**: Mobile-friendly layout
- **Syntax Highlighting**: Prism.js for code examples
- **Navigation**: Sidebar navigation for documentation sections
- **Search**: Built-in search functionality
- **Dark Mode**: Automatic dark/light mode support

### Customizing Styles

Edit `themes/agentflow/static/css/style.css` to modify the appearance.

### Customizing Layouts

The theme uses Hugo's template system:

- `baseof.html`: Base template with common HTML structure
- `home.html`: Homepage layout with hero section
- `single.html`: Individual page layout with sidebar
- `list.html`: Section listing pages

## 🔧 Configuration

The main configuration is in `config.toml`:

```toml
baseURL = "https://kunalkushwaha.github.io/agentflow/"
languageCode = "en-us"
title = "AgentFlow Documentation"
theme = "agentflow"

# Menu configuration
[menu]
  [[menu.main]]
    name = "Docs"
    url = "/docs/"
    weight = 10

# Theme parameters
[params]
  description = "The Go SDK for building production-ready multi-agent AI systems"
  github_repo = "https://github.com/kunalkushwaha/agentflow"
  edit_page = true
```

## 📤 Deployment

The site is automatically deployed to GitHub Pages using GitHub Actions:

1. **Trigger**: Pushes to `main` branch that modify files in `hugo-site/`
2. **Build**: Hugo Extended v0.119.0 builds the static site
3. **Deploy**: Files are deployed to the `gh-pages` branch
4. **Live**: Site is available at https://kunalkushwaha.github.io/agentflow/

### Manual Deployment

You can also build and deploy manually:

```bash
# Build the site
hugo --gc --minify

# The generated files will be in the public/ directory
# You can then deploy these files to any static hosting service
```

## 📋 Content Guidelines

### Writing Style

- **Clear and Concise**: Use simple, direct language
- **Code Examples**: Include working code examples for all concepts
- **Step-by-Step**: Break complex topics into numbered steps
- **Cross-References**: Link to related pages and sections

### Code Examples

- Always include complete, runnable examples
- Use proper Go formatting with `gofmt`
- Include necessary imports and error handling
- Provide context and explanation for each example

### Images and Diagrams

- Use descriptive alt text for accessibility
- Keep file sizes reasonable (< 1MB)
- Use consistent styling for diagrams
- Store images in `static/images/`

## 🤝 Contributing

To contribute to the documentation:

1. Fork the repository
2. Create a branch for your changes
3. Make your changes in the `hugo-site/` directory
4. Test locally with `hugo server`
5. Submit a pull request

For major changes, please open an issue first to discuss the proposed changes.

## 📚 Resources

- [Hugo Documentation](https://gohugo.io/documentation/)
- [Bootstrap 5 Documentation](https://getbootstrap.com/docs/5.3/)
- [Prism.js Documentation](https://prismjs.com/)
- [GitHub Pages Documentation](https://docs.github.com/en/pages)

## 🐛 Issues

If you find issues with the documentation site:

1. Check if it's already reported in [GitHub Issues](https://github.com/kunalkushwaha/agentflow/issues)
2. Create a new issue with:
   - Clear description of the problem
   - Steps to reproduce
   - Expected vs actual behavior
   - Screenshots if applicable