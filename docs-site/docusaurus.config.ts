import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'BookLife MCP',
  tagline: 'Your reading life, unified',
  favicon: 'img/favicon.ico',

  future: {
    v4: true,
  },

  url: 'https://andylbrummer.github.io',
  baseUrl: '/booklife-mcp/',

  organizationName: 'andylbrummer',
  projectName: 'booklife-mcp',
  deploymentBranch: 'gh-pages',
  trailingSlash: false,

  onBrokenLinks: 'throw',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          editUrl: 'https://github.com/andylbrummer/booklife-mcp/tree/main/docs-site/',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    colorMode: {
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'BookLife MCP',
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docsSidebar',
          position: 'left',
          label: 'Documentation',
        },
        {
          href: 'https://github.com/andylbrummer/andy-marketplace/tree/main/plugins/booklife',
          label: 'Claude Code Plugin',
          position: 'left',
        },
        {
          href: 'https://github.com/andylbrummer/booklife-mcp',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Documentation',
          items: [
            {label: 'Getting Started', to: '/docs/getting-started'},
            {label: 'Tool Reference', to: '/docs/category/tool-reference'},
            {label: 'Workflows', to: '/docs/category/workflows'},
            {label: 'Configuration', to: '/docs/configuration'},
          ],
        },
        {
          title: 'Integrations',
          items: [
            {label: 'Hardcover', href: 'https://hardcover.app'},
            {label: 'Libby', href: 'https://libbyapp.com'},
            {label: 'Open Library', href: 'https://openlibrary.org'},
          ],
        },
        {
          title: 'More',
          items: [
            {label: 'GitHub', href: 'https://github.com/andylbrummer/booklife-mcp'},
            {label: 'Claude Code Plugin', href: 'https://github.com/andylbrummer/andy-marketplace/tree/main/plugins/booklife'},
            {label: 'Andy Marketplace', href: 'https://github.com/andylbrummer/andy-marketplace'},
          ],
        },
      ],
      copyright: `Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'json', 'toml'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
