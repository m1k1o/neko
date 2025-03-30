import useBrokenLinks from '@docusaurus/useBrokenLinks';

export function AppIcon({id, ...props}: {id: string}) {
  useBrokenLinks().collectAnchor(id);
  return <a href={`#${id}`} {...props}>
    <img src={`/img/icons/${id}.svg`} id={id} title={id} className="anchorOffset" width="60" height="60" />
  </a>
}
