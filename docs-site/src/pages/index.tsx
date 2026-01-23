import React from 'react';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <header style={{padding: '4rem 0', textAlign: 'center', background: 'var(--ifm-color-primary-lightest)'}}>
      <div className="container">
        <h1 style={{fontSize: '3rem'}}>{siteConfig.title}</h1>
        <p style={{fontSize: '1.5rem', opacity: 0.8}}>{siteConfig.tagline}</p>
        <p style={{fontSize: '1.1rem', maxWidth: '600px', margin: '1rem auto'}}>
          One MCP server that connects your library, reading tracker, and bookshelf
          into a seamless AI-powered reading assistant.
        </p>
        <div style={{display: 'flex', gap: '1rem', justifyContent: 'center', marginTop: '2rem'}}>
          <Link className="button button--primary button--lg" to="/docs/getting-started">
            Get Started
          </Link>
          <Link className="button button--secondary button--lg" to="/docs/category/tool-reference">
            Tool Reference
          </Link>
        </div>
      </div>
    </header>
  );
}

function Feature({title, description}: {title: string; description: string}) {
  return (
    <div style={{flex: 1, padding: '1.5rem'}}>
      <h3>{title}</h3>
      <p>{description}</p>
    </div>
  );
}

export default function Home(): React.JSX.Element {
  return (
    <Layout title="Home" description="BookLife MCP - Your reading life, unified">
      <HomepageHeader />
      <main>
        <section style={{padding: '3rem 0'}}>
          <div className="container">
            <div style={{display: 'flex', flexWrap: 'wrap', gap: '1rem'}}>
              <Feature
                title="Library Access"
                description="Search your library catalog, check availability, place holds on ebooks and audiobooks through Libby/OverDrive."
              />
              <Feature
                title="Reading Tracker"
                description="Manage your Hardcover reading list, update progress, rate books, and maintain your history."
              />
              <Feature
                title="Unified TBR"
                description="One to-be-read list from Hardcover, Libby holds, tags, and physical books. Filter, search, and prioritize."
              />
            </div>
            <div style={{display: 'flex', flexWrap: 'wrap', gap: '1rem', marginTop: '1rem'}}>
              <Feature
                title="Comprehensive Sync"
                description="One command to import loans, sync history to Hardcover, enrich metadata, and cache tags."
              />
              <Feature
                title="Recommendations"
                description="Content-based book discovery using enriched themes, topics, and mood data from your reading history."
              />
              <Feature
                title="27 Tools"
                description="Progressive discovery with built-in help, workflow guides, and automation-friendly metadata."
              />
            </div>
          </div>
        </section>
        <section style={{padding: '2rem 0', background: 'var(--ifm-background-surface-color)'}}>
          <div className="container" style={{textAlign: 'center'}}>
            <h2>Claude Code Plugin</h2>
            <p>
              Enhanced integration with skills and slash commands via the{' '}
              <a href="https://github.com/andylbrummer/andy-marketplace/tree/main/plugins/booklife">
                BookLife plugin in andy-marketplace
              </a>.
            </p>
            <code style={{display: 'block', margin: '1rem auto', maxWidth: '500px', textAlign: 'left', padding: '1rem'}}>
              /plugin install booklife@andy-marketplace
            </code>
          </div>
        </section>
      </main>
    </Layout>
  );
}
