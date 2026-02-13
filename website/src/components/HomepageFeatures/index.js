import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

const FeatureList = [
  {
    title: 'Monitor AI Agents',
    icon: 'icon-eye',
    description: (
      <>
        Watch Claude Code, Cursor, and other AI coding agents work in real-time.
        See what they're doing without context-switching.
      </>
    ),
  },
  {
    title: 'Terminal Native',
    icon: 'icon-terminal',
    description: (
      <>
        A beautiful TUI that feels at home in your terminal. Vim-style keybindings
        and a keyboard-first workflow.
      </>
    ),
  },
  {
    title: 'Stay in Flow',
    icon: 'icon-rocket',
    description: (
      <>
        Run agents in the background while you work. Get notified when they need
        input or finish their tasks.
      </>
    ),
  },
];

function Feature({icon, title, description}) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <span className={styles.featureIcon}>
          <i className={icon} />
        </span>
      </div>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className={clsx('row', styles.featureRow)}>
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
