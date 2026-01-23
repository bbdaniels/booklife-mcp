import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    'intro',
    'getting-started',
    'configuration',
    {
      type: 'category',
      label: 'Tool Reference',
      link: {
        type: 'generated-index',
        description: 'Complete reference for all 27 BookLife MCP tools.',
      },
      items: [
        'tools/info',
        {
          type: 'category',
          label: 'Hardcover',
          items: [
            'tools/hardcover-get-my-library',
            'tools/hardcover-update-reading-status',
            'tools/hardcover-add-to-library',
          ],
        },
        {
          type: 'category',
          label: 'Libby',
          items: [
            'tools/libby-search',
            'tools/libby-get-loans',
            'tools/libby-get-holds',
            'tools/libby-place-hold',
            'tools/libby-sync-tag-metadata',
            'tools/libby-tag-metadata-list',
          ],
        },
        {
          type: 'category',
          label: 'TBR Management',
          items: [
            'tools/tbr-list',
            'tools/tbr-search',
            'tools/tbr-add',
            'tools/tbr-remove',
            'tools/tbr-sync',
            'tools/tbr-stats',
          ],
        },
        {
          type: 'category',
          label: 'Unified Actions',
          items: [
            'tools/booklife-find-book-everywhere',
            'tools/booklife-best-way-to-read',
          ],
        },
        {
          type: 'category',
          label: 'History',
          items: [
            'tools/history-import-timeline',
            'tools/history-sync-current-loans',
            'tools/history-get',
            'tools/history-stats',
          ],
        },
        {
          type: 'category',
          label: 'Sync & Enrichment',
          items: [
            'tools/sync',
            'tools/enrichment-enrich-history',
            'tools/enrichment-status',
          ],
        },
        {
          type: 'category',
          label: 'Recommendations & Profile',
          items: [
            'tools/book-find-similar',
            'tools/profile-get',
          ],
        },
      ],
    },
    {
      type: 'category',
      label: 'Workflows',
      link: {
        type: 'generated-index',
        description: 'Step-by-step guides for common reading management tasks.',
      },
      items: [
        'workflows/find-and-read',
        'workflows/sync-history',
        'workflows/tbr-management',
        'workflows/recommendations',
      ],
    },
    'claude-code-plugin',
    'cli-reference',
    'architecture',
  ],
};

export default sidebars;
