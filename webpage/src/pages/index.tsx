// This is handler for old links that were used in v2.
// They were used only as hash links, so we need to redirect them to the new links.
if (typeof window !== 'undefined' && window.location.hash) {
  let hash = window.location.hash.substring(1);
  hash = hash.replace('?id=', '#');
  if (/^[a-z0-9\-#\/]+$/.test(hash)) {
    // if id starts with known path
    if (hash.startsWith('/getting-started')) {
      // remove /getting-started
      hash = hash.replace('/getting-started', '');
      // add /docs/v2
      window.location.href = '/docs/v2' + hash;
    }
  }
}

import type {ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import HomepageUseCases from '@site/src/components/HomepageUseCases';
import HomepageShowcase from '@site/src/components/HomepageShowcase';
import HomepageFeatures from '@site/src/components/HomepageFeatures';
import Heading from '@theme/Heading';

import styles from './index.module.css';

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className="container">
        <Heading as="h1" className="hero__title">
          <img
            alt="n.eko"
            className={styles.heroLogo}
            src="img/logo.png"
            width="450"
          />
        </Heading>
        <p className="hero__subtitle">{siteConfig.tagline}</p>
        <div className={styles.buttons}>
          <Link
            className="button button--secondary button--lg"
            to="/docs/v3/quick-start">
            Get started
          </Link>
          <span className={styles.indexCtasGitHubButtonWrapper}>
            <iframe
              className={styles.indexCtasGitHubButton}
              src="https://ghbtns.com/github-btn.html?user=m1k1o&amp;repo=neko&amp;type=star&amp;count=true&amp;size=large"
              width={160}
              height={30}
              title="GitHub Stars"
            />
          </span>
        </div>
      </div>
    </header>
  );
}

export default function Home(): ReactNode {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      description={siteConfig.tagline}>
      <HomepageHeader />
      <main>
        <section className={styles.description}>
          <div className="container">
            <p className="text--center">
              Welcome to Neko, a self-hosted virtual browser that runs in Docker and uses WebRTC technology. Neko allows you to <strong>run a fully-functional browser in a virtual environment</strong>, providing <strong>secure and private internet access</strong> from anywhere. It's perfect for developers, privacy-conscious users, and anyone needing a <strong>virtual browser</strong>.
            </p>
          </div>
        </section>
        <HomepageUseCases />
        <HomepageShowcase />
        <HomepageFeatures />
      </main>
    </Layout>
  );
}
