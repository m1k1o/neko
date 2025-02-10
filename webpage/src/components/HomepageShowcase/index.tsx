import type {ReactNode} from 'react';
import styles from './styles.module.css';

export default function HomepageShowcase(): ReactNode {
  return (
    <section className={styles.showcase}>
      <div className={`container ${styles.center}`}>
        <div className="row">
          <img
            alt="n.eko"
            className={styles.showcaseImage}
            src="img/intro.gif"
          />
        </div>
      </div>
    </section>
  );
}
