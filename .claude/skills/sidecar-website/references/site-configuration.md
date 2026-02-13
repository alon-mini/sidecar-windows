# Docusaurus Site Configuration Reference

Detailed configuration for `docusaurus.config.js`.

## Basic Info

```javascript
const config = {
  title: 'Sidecar',
  tagline: 'Terminal UI for monitoring AI coding agent sessions',
  url: 'https://marcus.github.io',
  baseUrl: '/sidecar/',
  organizationName: 'marcus',
  projectName: 'sidecar',
};
```

## Navbar

```javascript
themeConfig: {
  navbar: {
    title: 'Sidecar',
    items: [
      { type: 'docSidebar', sidebarId: 'tutorialSidebar', label: 'Docs' },
      { to: '/blog', label: 'Blog' },
      { href: 'https://github.com/marcus/sidecar', label: 'GitHub' },
    ],
  },
}
```

Custom navbar items are defined in `src/theme/NavbarItem/index.js` wrapping components from `src/components/`.

## Theme Configuration

Dark mode is configured with `disableSwitch: true` -- users cannot toggle between light and dark modes.

## Footer

```javascript
footer: {
  style: 'dark',
  links: [
    { title: 'Docs', items: [{ label: 'Getting Started', to: '/docs/intro' }] },
  ],
  copyright: `Copyright (c) ${new Date().getFullYear()} Sidecar.`,
}
```

## Future Compatibility

The config includes `future: { v4: true }` for Docusaurus v4 compatibility.

## Theme Colors

Edit `src/css/custom.css`:

```css
:root {
  --ifm-color-primary: #2e8555;
}
[data-theme='dark'] {
  --ifm-color-primary: #25c2a0;
}
```

## Resources

- [Docusaurus Documentation](https://docusaurus.io/docs)
- [Markdown Features](https://docusaurus.io/docs/markdown-features)
- [Styling and Layout](https://docusaurus.io/docs/styling-layout)
- [Configuration](https://docusaurus.io/docs/configuration)
- [Deployment to GitHub Pages](https://docusaurus.io/docs/deployment#deploying-to-github-pages)
