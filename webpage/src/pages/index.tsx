import type {ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import HomepageFeatures from '@site/src/components/HomepageFeatures';
import HomepageShowcase from '@site/src/components/HomepageShowcase';
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
            to="/docs/intro">
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
        <HomepageFeatures />
        <HomepageShowcase />
        <section className={styles.description}>
          <div className="container">
            <p>
              Welcome to Neko, a self-hosted virtual browser that runs in Docker and uses WebRTC technology. Neko is a powerful tool that allows you to <strong>run a fully-functional browser in a virtual environment</strong>, giving you the ability to <strong>access the internet securely and privately from anywhere</strong>. With Neko, you can browse the web, <strong>run applications</strong>, and perform other tasks just as you would on a regular browser, all within a <strong>secure and isolated environment</strong>. Whether you are a developer looking to test web applications, a <strong>privacy-conscious user seeking a secure browsing experience</strong>, or simply someone who wants to take advantage of the <strong>convenience and flexibility of a virtual browser</strong>, Neko is the perfect solution.
            </p>
            <p>
              In addition to its security and privacy features, Neko offers the <strong>ability for multiple users to access it simultaneously</strong>. This makes it an ideal solution for teams or organizations that need to share access to a browser, as well as for individuals who want to use <strong>multiple devices to access the same virtual environment</strong>. With Neko, you can <strong>easily and securely share access to a browser with others</strong>, without having to worry about maintaining separate configurations or settings. Whether you need to <strong>collaborate on a project</strong>, access shared resources, or simply want to <strong>share access to a browser with friends or family</strong>, Neko makes it easy to do so.
            </p>
            <p>
              Neko is also a great tool for <strong>hosting watch parties</strong> and interactive presentations. With its virtual browser capabilities, Neko allows you to host watch parties and presentations that are <strong>accessible from anywhere</strong>, without the need for in-person gatherings. This makes it easy to <strong>stay connected with friends and colleagues</strong>, even when you are unable to meet in person. With Neko, you can easily host a watch party or give an <strong>interactive presentation</strong>, whether it's for leisure or work. Simply invite your guests to join the virtual environment, and you can share the screen and <strong>interact with them in real-time</strong>.
            </p>
          </div>
        </section>
      </main>
    </Layout>
  );
}
