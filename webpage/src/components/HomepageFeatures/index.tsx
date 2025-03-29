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
        Near real-time streaming with less than 300ms latency thanks to <a href="https://webrtc.org/">WebRTC</a>. <br />
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
        Stream your room's content live to platforms like Twitch, YouTube, and more via RTMP. As the host, you have full
        control over the stream; set the RTMP URL and stream key, start or stop the broadcast at any time. Even if no
        participants are online, the stream keeps running, making 24/7 broadcasting effortless.
      </>
    ),
  },
  {
    isEven: i++ % 2 === 0,
    title: 'Persistent Browser',
    Svg: require('@site/static/img/undraw_safe.svg').default,
    description: (
      <>
        Keep your browser session alive, no matter where you are. Resume your work from any device without losing your progress.
        Ideal for long-running tasks like downloads, uploads, and monitoring. No local data is stored; cookies and session data
        stay protected. For your ISP, it just looks like you're watching a video or on a call, keeping your activity private.
      </>
    ),
  },
  {
    isEven: i++ % 2 === 0,
    title: 'Throwaway Browser',
    Svg: require('@site/static/img/undraw_throw_away.svg').default,
    description: (
      <>
        Access websites without leaving a trace. Every session runs in an isolated environment and is destroyed afterward;
        no history, cookies, or cache left behind. Perfect for handling sensitive information or testing websites without
        affecting your local machine. Minimize OS fingerprinting and browser exploits by running in a secure container.
        Need extra privacy? Use Tor Browser and VPN for added anonymity.
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
      <div id="features" className="container">
        <h1 className={"text--center"}>Features</h1>
        {FeatureList.map((props, idx) => (
          <Feature key={idx} {...props} />
        ))}
      </div>
    </section>
  );
}
