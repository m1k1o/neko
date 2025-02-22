import type {ReactNode} from 'react';
import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

type FeatureItem = {
  title: string;
  Svg: React.ComponentType<React.ComponentProps<'svg'>>;
  description: ReactNode;
  isEven: boolean;
};

let i = 1;

const FeatureList: FeatureItem[] = [
  {
    isEven: i++ % 2 === 0,
    title: 'Ultra Low Latency Streaming',
    Svg: require('@site/static/img/undraw_low_latency.svg').default,
    description: (
      <>
        Near real-time streaming with less than 300ms latency thanks to WebRTC. <br />
        Smooth video playback and synchronized audio for a seamless experience.
      </>
    ),
  },
  {
    isEven: i++ % 2 === 0,
    title: 'Multi-Participant Control',
    Svg: require('@site/static/img/undraw_miro.svg').default,
    description: (
      <>
        Multiple participants in the room can share control with each other. <br />
        The host has the ability to give control to others or deny control requests.
      </>
    ),
  },
  {
    isEven: i++ % 2 === 0,
    title: 'Live Broadcasting',
    Svg: require('@site/static/img/undraw_online_media.svg').default,
    description: (
      <>
        Stream room content live via RTMP to platforms like Twitch, YouTube, and more. <br />
        The host can manage the RTMP URL and stream key, enabling or disabling the broadcast as needed. <br />
        Broadcasting can continue even without active participants, allowing for 24/7 streaming.
      </>
    ),
  },
  {
    isEven: i++ % 2 === 0,
    title: 'Persistent Browser',
    Svg: require('@site/static/img/undraw_safe.svg').default,
    description: (
      <>
        Keep your browser session running even after you close the tab or browser.
        Resume your session at any time from any device.
        Ideal for long-running tasks like downloads, uploads, and monitoring. <br /><br />
        Own a browser with persistent cookies available anywhere.
        No state is left on the host browser after terminating the connection.
        Sensitive data like cookies are not transferred - only video is shared.
        For your ISP, it looks like you are watching a video or having a video call.
      </>
    ),
  },
  {
    isEven: i++ % 2 === 0,
    title: 'Throwaway Browser',
    Svg: require('@site/static/img/undraw_throw_away.svg').default,
    description: (
      <>
        Use a disposable browser to access websites without leaving any trace.
        The browser is destroyed after the session ends, leaving no history, cookies, or cache.
        Ideal for accessing sensitive information or testing websites without affecting your local environment. <br /><br />
        Mitigates the risk of OS fingerprinting and browser vulnerabilities by running in a container.
        Use Tor Browser and VPN for additional anonymity.
      </>
    ),
  },
  {
    isEven: i++ % 2 === 0,
    title: 'Jump Host for Internal Resources',
    Svg: require('@site/static/img/undraw_software_engineer.svg').default,
    description: (
      <>
        Access internal resources like servers, databases, and websites from a remote location.
        You can record all session activities for auditing and compliance.
        Ensuring that no data is left on the client side and minimizing the risk of data leakage.
        Making it harder for attackers to pivot to other systems when they compromise the jump host and reducing the attack surface.
      </>
    ),
  },
  {
    isEven: i++ % 2 === 0,
    title: 'Protect Your Intellectual Property',
    Svg: require('@site/static/img/undraw_security.svg').default,
    description: (
      <>
        Have you ever wanted to share your website, AI model, or software with someone without giving them access to the code?
        With WebRTC, only the video and audio are shared, not the actual data. Nobody can reverse-engineer your code because it is not even running on their machine.
        You have full control over who can access your content and can revoke access at any time.
      </>
    ),
  },
];

function Feature({title, Svg, description, isEven}: FeatureItem) {
  return (
    <div className={clsx('row', styles.featureRow, { [styles.featureRowReverse]: !isEven })}>
      <div className={clsx('col col--5', 'text--center')}>
        <Svg className={styles.featureSvg} role="img" />
      </div>
      <div className={clsx('col col--7', 'padding-horiz--md')}>
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures(): ReactNode {
  return (
    <section className={styles.features}>
      <div className="container">
        <h1 id="features">Features</h1>
        {FeatureList.map((props, idx) => (
          <Feature key={idx} {...props} />
        ))}
      </div>
    </section>
  );
}
