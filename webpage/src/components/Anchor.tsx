import useBrokenLinks from '@docusaurus/useBrokenLinks';
import React from 'react';

export function Anchor(props: {id: string}) {
  useBrokenLinks().collectAnchor(props.id);
  return <span className="anchorOffset" {...props}/>;
}

// <Def id="cookie.http_only" /> -> <a href="#cookie.http_only"><code id="cookie.http_only">http_only</code></a>
export function Def({id, code, ...props}: {id: string, code?: string}) {
  // split by . and get last part
  if (code === undefined) {
    code = id.split('.').pop();
  }
  useBrokenLinks().collectAnchor(id);
  // get current heading id
  return <a href={`#${id}`} {...props}><code id={id} title={id} className="anchorOffset">{code}</code></a>;
}

// <Opt id="cookie.http_only" /> -> <a href="#cookie.http_only"><code>http_only</code></a>
export function Opt({id, code, ...props}: {id: string, code?: string}) {
  // split by . and get last part
  if (code === undefined) {
    code = id.split('.').pop();
  }
  // get current heading id
  return <a href={`#${id}`} {...props}><code title={id} className="anchorOffset">{code}</code></a>;
}
