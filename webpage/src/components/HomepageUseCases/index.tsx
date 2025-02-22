import type {ReactNode} from 'react';
import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

type FeatureItem = {
  title: string;
  Svg: React.ComponentType<React.ComponentProps<'svg'>>;
  description: ReactNode;
};

const FeatureList: FeatureItem[] = [
  {
    title: 'Watch Party',
    Svg: require('@site/static/img/undraw_video_streaming.svg').default,
    description: (
      <>
        Watch video content together with multiple people and react to it in real-time.
        Perfect for staying connected with friends and family.
      </>
    ),
  },
  {
    title: 'Interactive Presentation',
    Svg: require('@site/static/img/undraw_usability_testing.svg').default,
    description: (
      <>
        Share your screen and allow others to control it.
        Ideal for collaborative work and interactive teaching sessions.
      </>
    ),
  },
  {
    title: 'Collaborative Tool',
    Svg: require('@site/static/img/undraw_online_collaboration.svg').default,
    description: (
      <>
        Brainstorm ideas, co-browse, and debug code together.
        Enhance team collaboration with real-time synchronization.
      </>
    ),
  },
  {
    title: 'Support/Teaching',
    Svg: require('@site/static/img/undraw_instant_support.svg').default,
    description: (
      <>
        Guide people interactively in a controlled environment.
        Perfect for providing support or teaching remotely.
      </>
    ),
  },
  {
    title: 'Embed Anything',
    Svg: require('@site/static/img/undraw_website_builder.svg').default,
    description: (
      <>
        Embed a virtual browser in your web app. Open any third-party
        website or application and synchronize audio and video flawlessly.
      </>
    ),
  },
  {
    title: 'Automated Browser',
    Svg: require('@site/static/img/undraw_data_processing.svg').default,
    description: (
      <>
        Install Playwright or Puppeteer and automate tasks while being
        able to actively intercept them. Enhance productivity with automation.
      </>
    ),
  },
];

function Feature({title, Svg, description}: FeatureItem) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <Svg className={styles.useCaseSvg} role="img" />
      </div>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageUseCases(): ReactNode {
  return (
    <section className={styles.useCases}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
