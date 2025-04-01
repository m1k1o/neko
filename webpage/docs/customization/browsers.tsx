import React from 'react';
import browsers from './browsers.json';
import Link from '@docusaurus/Link';

export function ProfileDirectoryPaths({flavors, ...props}: { flavors: string[] }) {
    if (!flavors) {
        flavors = [];
    }
    return (
        <table {...props}>
            <thead>
                <tr>
                    <th>Browser</th>
                    <th>Profile Directory Path</th>
                </tr>
            </thead>
            <tbody>
                {browsers.filter(({ flavor }) => flavors.length == 0 || flavors.includes(flavor)).map(({ tag, profileDir }) => (
                    <tr key={tag}>
                        <td><Link to={`/docs/v3/installation/docker-images#${tag}`} ><strong>{tag}</strong></Link></td>
                        <td>
                            {profileDir ? <code>{profileDir}</code> : <i>Does not support profiles.</i>}
                        </td>
                    </tr>
                ))}
            </tbody>
        </table>
    );
}

export function PolicyFilePaths({flavors, ...props}: { flavors: string[] }) {
    if (!flavors) {
        flavors = [];
    }
    return (
        <table {...props}>
            <thead>
                <tr>
                    <th>Browser</th>
                    <th>Policy File Path</th>
                </tr>
            </thead>
            <tbody>
                {browsers.filter(({ flavor }) => flavors.length == 0 || flavors.includes(flavor)).map(({ tag, policiesFile }) => (
                    <tr key={tag}>
                        <td><Link to={`/docs/v3/installation/docker-images#${tag}`} ><strong>{tag}</strong></Link></td>
                        <td>
                            {policiesFile ? <code>{policiesFile}</code> : <i>Does not support policies.</i>}
                        </td>
                    </tr>
                ))}
            </tbody>
        </table>
    );
}
