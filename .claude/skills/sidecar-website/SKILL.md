---
name: sidecar-website
description: Writing and maintaining the Sidecar Docusaurus documentation site, including page structure, doc authoring, blog posts, styling, images, and deployment workflow. Use when writing documentation, updating the docs site, adding pages or blog posts, or working with Docusaurus configuration.
---

# Docusaurus Documentation Site

The documentation site lives in `website/`. It uses Docusaurus with Node.js >= 20.

## Quick Start

```bash
cd website
npm install    # First time only
npm start      # Dev server at http://localhost:3000
```

## Project Structure

```
website/
├── docs/                    # Markdown documentation pages
├── blog/                    # Blog posts (date-prefixed markdown)
│   ├── authors.yml         # Blog author definitions
│   └── tags.yml            # Blog tag definitions
├── src/
│   ├── pages/              # Custom React pages (non-docs)
│   │   ├── index.js        # Front page (/)
│   │   └── index.module.css
│   ├── components/         # Reusable React components
│   └── css/
│       └── custom.css      # Global style overrides
├── static/                  # Static assets (copied as-is to build)
│   └── img/                # Images
├── docusaurus.config.js    # Main site configuration
├── sidebars.js             # Docs sidebar structure
└── package.json
```

## Writing Documentation

### Principles

- **User-first**: Answer "what can I do?" before "how does it work?"
- **Scannable**: Use headers, code blocks, tables for keyboard shortcuts
- **Progressive disclosure**: Quick overview -> detailed usage -> full reference
- **Working examples**: Every feature needs runnable code, not `...` placeholders

### Creating a New Doc

Add a Markdown file in `website/docs/` with YAML frontmatter:

```markdown
---
sidebar_position: 2
title: My New Page
---

# My New Page

Content here. Supports **Markdown** and MDX.
```

**Frontmatter options**:
- `sidebar_position`: Order in sidebar (lower = higher)
- `sidebar_label`: Override sidebar text
- `title`: Page title
- `description`: Meta description for SEO
- `slug`: Custom URL path

### Plugin Documentation Pattern

```markdown
# Plugin Name

One-line description.

![Screenshot](../../docs/screenshots/plugin-name.png)

## Overview
Brief explanation of UI layout and core purpose.

## Feature Section
Description with keyboard shortcut table:

| Key | Action |
|-----|--------|
| `s` | Stage file |
| `d` | View diff |

## Navigation
How to move around.

## Command Reference
Complete shortcut list by context.
```

### Organizing Docs in Folders

```
docs/
├── intro.md
├── guides/
│   ├── _category_.json    # Folder metadata
│   ├── installation.md
│   └── configuration.md
```

`_category_.json` controls folder appearance:

```json
{
  "label": "Guides",
  "position": 2,
  "collapsible": true,
  "collapsed": false
}
```

### Sidebar Configuration

Auto-generates from `docs/` folder structure. To customize, edit `sidebars.js`:

```javascript
const sidebars = {
  tutorialSidebar: [
    'intro',
    {
      type: 'category',
      label: 'Guides',
      items: ['guides/installation', 'guides/usage'],
    },
  ],
};
```

### Using MDX

Docs support MDX (Markdown + JSX):

```mdx
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs>
  <TabItem value="npm">npm install</TabItem>
  <TabItem value="yarn">yarn add</TabItem>
</Tabs>
```

## Front Page

The front page is at `website/src/pages/index.js` -- a React component using Docusaurus Layout and theming.

```jsx
export default function Home() {
  return (
    <Layout title="Home" description="...">
      <HomepageHeader />
      <main className="container">
        {/* Add content here */}
      </main>
    </Layout>
  );
}
```

Styling: `index.module.css` for page-specific, `src/css/custom.css` for global overrides.

## Images and Screenshots

**Website doc screenshots** (for pages in `website/docs/`):
- Store in: `docs/screenshots/` (project root)
- Reference: `![Alt text](../../docs/screenshots/filename.png)`

**README / repo doc screenshots**:
- Store in: `docs/screenshots/` (project root)
- Reference: `![Alt text](docs/screenshots/filename.png)` (from repo root)

**General website images** (logos, icons):
- Store in: `website/static/img/`
- Reference: `![Alt text](/img/filename.png)`

In JSX:
```jsx
import screenshot from '@site/static/img/logo.png';
<img src={screenshot} alt="Logo" />
```

## Blog Posts

Date-prefixed Markdown files in `blog/`:

```markdown
---
slug: my-post
title: Post Title
authors: [default]
tags: [announcement, release]
---

Preview text shown in list.

<!-- truncate -->

Full content below the fold.
```

## Style Guidelines

### No Emoji Policy

Never use emoji in site content, components, or documentation. Use Lucide icons instead.

### Icons (Lucide)

The site uses Lucide icon font (CDN import in `docusaurus.config.js`).

```jsx
<i className="icon-terminal" />
<i className="icon-check" />
<i className="icon-git-branch" />
```

Common icons: `icon-eye`, `icon-terminal`, `icon-rocket`, `icon-check`, `icon-copy`, `icon-external-link`, `icon-git-branch`, `icon-zap`, `icon-keyboard`, `icon-layers`, `icon-code`.

Browse all: https://lucide.dev/icons

### Terminal Aesthetic

- Monospace fonts (`JetBrains Mono`, `Google Sans Code`)
- Dark backgrounds with muted colors
- Bright accents from Monokai palette (green, blue, pink, yellow)
- Clean 1px borders, subtle gradients and glows

## Building and Deploying

```bash
cd website
npm run build      # Outputs to website/build/
npm run serve      # Preview built site locally
```

Deploys automatically via GitHub Actions when changes to `website/` merge to `main`.

- `.github/workflows/deploy-docs.yml` -- Deploys to GitHub Pages
- `.github/workflows/test-docs.yml` -- Validates PR builds
- Live site: https://marcus.github.io/sidecar

## Common Tasks

| Task | Steps |
|------|-------|
| Add docs section | Create folder in `website/docs/`, add `_category_.json`, add Markdown files |
| Change theme colors | Edit `src/css/custom.css` (`:root` and `[data-theme='dark']` variables) |
| Add custom component | Create in `src/components/MyComponent/index.js`, import with `@site/src/components/MyComponent` |

## Troubleshooting

- **Build fails with broken links**: Config uses `onBrokenLinks: 'throw'`. Temporarily change to `'warn'` for local dev.
- **Styles not updating**: `npm run clear && npm start`
- **GitHub Pages 404**: Verify `baseUrl` matches repo name (`/sidecar/`).

## Reference

For detailed site configuration (navbar, footer, theme config, future compatibility), see `references/site-configuration.md`.
