# AgentFlow Hugo Documentation Site

This directory contains the Hugo-based documentation site for AgentFlow, designed to be deployed to GitHub Pages.

## Structure

```
hugo-site/
├── config.toml              # Hugo configuration
├── content/                 # Documentation content
│   ├── _index.md           # Home page
│   ├── docs/               # Core documentation
│   ├── examples/           # Example workflows and code
│   └── contributors/       # Contributor guides
├── layouts/                # Hugo templates
│   └── _default/           # Default layouts
├── static/                 # Static assets
└── public/                 # Generated site (ignored)
```

## Local Development

### Prerequisites

- Hugo Extended v0.119.0 or later
- Git

### Setup

```bash
# From the repository root
cd hugo-site

# Build the site
hugo

# Start development server
hugo server -D

# The site will be available at http://localhost:1313
```

### Building for Production

```bash
hugo --gc --minify --baseURL "https://kunalkushwaha.github.io/agentflow/"
```

## Content Organization

### Docs Section (`content/docs/`)
- Core concepts and guides
- API references  
- Architecture documentation

### Examples Section (`content/examples/`)
- Practical code samples
- Workflow tutorials
- Integration examples

### Contributors Section (`content/contributors/`)
- Development setup guides
- Coding standards
- Testing strategies
- Release processes

## Deployment

The site is automatically deployed to GitHub Pages via GitHub Actions when changes are pushed to the `main` branch. The workflow:

1. Checks out the repository
2. Installs Hugo
3. Builds the site 
4. Deploys to the `gh-pages` branch
5. Site becomes available at https://kunalkushwaha.github.io/agentflow/

## Theme and Styling

The site uses a custom theme built with Bootstrap 5 and includes:

- Responsive design
- Code syntax highlighting
- Navigation sidebar
- Search functionality (planned)
- Dark/light mode (planned)

## Adding Content

1. Create new `.md` files in the appropriate content sections
2. Include proper frontmatter with title, weight, and description
3. Use Hugo's markdown extensions for enhanced formatting
4. Test locally before committing

## Configuration

Key settings in `config.toml`:

- `baseURL`: GitHub Pages URL
- `title`: Site title
- `menu.main`: Navigation menu items
- `params.github_repo`: Repository URL for edit links