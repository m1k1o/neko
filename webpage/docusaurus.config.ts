import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';
import type * as Plugin from "@docusaurus/types/src/plugin";
import type * as OpenApiPlugin from "docusaurus-plugin-openapi-docs";

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

const config: Config = {
  title: 'n.eko',
  tagline: 'A self hosted virtual browser that runs in docker and uses WebRTC.',
  favicon: 'img/favicon.ico',

  // Set the production url of your site here
  url: 'https://neko.m1k1o.net',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/',

  // GitHub pages deployment config.
  organizationName: 'm1k1o',
  projectName: 'neko',
  trailingSlash: false,

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',

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
          docItemComponent: "@theme/ApiItem", // Derived from docusaurus-theme-openapi
          editUrl: 'https://github.com/m1k1o/neko/tree/main/webpage/',
          lastVersion: 'current',
          versions: {
            current: {
              label: 'v3',
              path: 'v3',
            },
          },
        },
        //blog: {
        //  showReadingTime: true,
        //  feedOptions: {
        //    type: ['rss', 'atom'],
        //    xslt: true,
        //  },
        //  editUrl: 'https://github.com/m1k1o/neko/tree/main/docs/',
        //  // Useful options to enforce blogging best practices
        //  onInlineTags: 'warn',
        //  onInlineAuthors: 'warn',
        //  onUntruncatedBlogPosts: 'warn',
        //},
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    image: 'img/neko-social-card.jpg',
    navbar: {
      //title: 'n.eko',
      logo: {
        alt: 'n.eko',
        src: 'img/logo.png',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docsSidebar',
          position: 'left',
          label: 'Docs',
        },
        {
          to: 'contributing',
          label: 'Contributing',
          position: 'left',
        },
        {
          to: 'non-goals',
          label: 'Non-Goals',
          position: 'left',
        },
        {
          to: 'contact',
          label: 'Contact',
          position: 'left',
        },
        {
          href: 'https://github.com/sponsors/m1k1o',
          label: 'Donate',
          position: 'left',
        },
        {
          type: 'docsVersionDropdown',
          position: 'right',
        },
        {
          href: 'https://discord.gg/3U6hWpC',
          label: 'Discord',
          position: 'right',
        },
        {
          href: 'https://github.com/m1k1o/neko',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Other Projects',
          items: [
            {
              label: 'Neko Rooms',
              href: 'https://github.com/m1k1o/neko-rooms',
            },
            {
              label: 'Neko Apps',
              href: 'https://github.com/m1k1o/neko-apps',
            },
            {
              label: 'Neko VPN',
              href: 'https://github.com/m1k1o/neko-vpn',
            },
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'Discord',
              href: 'https://discord.gg/3U6hWpC',
            },
            {
              label: 'Issues',
              href: 'https://github.com/m1k1o/neko/issues',
            },
          ],
        },
        {
          title: 'More',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/m1k1o/neko',
            },
            {
              label: 'Sponsors',
              href: 'https://github.com/sponsors/m1k1o',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} <a href="https://github.com/m1k1o">m1k1o</a>. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash'],
    },
  } satisfies Preset.ThemeConfig,

  plugins: [
    [
      "docusaurus-plugin-openapi-docs",
      {
        id: "openapi",
        docsPluginId: "classic",
        config: {
          api: {
            specPath: "../server/openapi.yaml",
            outputDir: "docs/api",
            downloadUrl: "https://raw.githubusercontent.com/m1k1o/neko/refs/heads/master/server/openapi.yaml",
            baseUrl: "/docs/v3/api",
            sidebarOptions: {
              groupPathsBy: "tag",
              categoryLinkSource: "tag",
              sidebarCollapsed: false,
            },
          } satisfies OpenApiPlugin.Options,
        } satisfies Plugin.PluginOptions,
      },
    ],
  ],

  themes: ["docusaurus-theme-openapi-docs"],
};

export default config;
